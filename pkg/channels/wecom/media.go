package wecom

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"

	"github.com/sipeed/picoclaw/pkg/media"
)

func decodeMediaAESKey(value string) ([]byte, error) {
	if value == "" {
		return nil, nil
	}
	key, err := base64.StdEncoding.DecodeString(value)
	if err == nil && len(key) == 32 {
		return key, nil
	}
	key, err = base64.StdEncoding.DecodeString(value + "=")
	if err != nil {
		return nil, fmt.Errorf("decode AES key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid AES key length %d", len(key))
	}
	return key, nil
}

func decryptAESCBC(key, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, fmt.Errorf("ciphertext is empty")
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext length %d is not a multiple of block size", len(ciphertext))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	plaintext := make([]byte, len(ciphertext))
	iv := key[:aes.BlockSize]
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, ciphertext)
	return pkcs7Unpad(plaintext)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty plaintext")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > 32 || padding > len(data) {
		return nil, fmt.Errorf("invalid padding size %d", padding)
	}
	for i := 0; i < padding; i++ {
		if data[len(data)-1-i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding byte")
		}
	}
	return data[:len(data)-padding], nil
}

func inferMediaExt(contentType, fallback string) string {
	contentType = normalizeWeComContentType(contentType)
	switch contentType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	case "video/mp4":
		return ".mp4"
	default:
		return fallback
	}
}

func normalizeWeComContentType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if idx := strings.Index(value, ";"); idx >= 0 {
		value = strings.TrimSpace(value[:idx])
	}
	return value
}

func isGenericWeComContentType(value string) bool {
	switch normalizeWeComContentType(value) {
	case "", "application/octet-stream", "binary/octet-stream", "application/unknown", "application/binary":
		return true
	default:
		return false
	}
}

func sanitizeWeComFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "." || name == "/" || name == "" {
		return ""
	}
	return name
}

func candidateWeComFilename(resourceURL, contentDisposition, fallbackName string) string {
	if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
		if name := sanitizeWeComFilename(params["filename"]); name != "" {
			return name
		}
		if name := sanitizeWeComFilename(params["filename*"]); name != "" {
			return name
		}
	}

	if parsed, err := url.Parse(resourceURL); err == nil {
		query := parsed.Query()
		for _, key := range []string{"filename", "file_name", "name"} {
			if name := sanitizeWeComFilename(query.Get(key)); name != "" {
				return name
			}
		}
		if name := sanitizeWeComFilename(parsed.Path); name != "" {
			return name
		}
	}

	return sanitizeWeComFilename(fallbackName)
}

func detectWeComFiletype(data []byte) (string, string) {
	kind, err := filetype.Match(data)
	if err != nil || kind == filetype.Unknown {
		return "", ""
	}
	ext := ""
	if kind.Extension != "" {
		ext = "." + strings.ToLower(kind.Extension)
	}
	return normalizeWeComContentType(kind.MIME.Value), ext
}

func detectWeComMediaMetadata(data []byte, fallbackName, fallbackContentType, resourceURL, contentDisposition string) (string, string) {
	filename := candidateWeComFilename(resourceURL, contentDisposition, fallbackName)
	if filename == "" {
		filename = "media"
	}

	ext := strings.ToLower(filepath.Ext(filename))
	contentType := normalizeWeComContentType(fallbackContentType)
	detectedType, detectedExt := detectWeComFiletype(data)

	if ext != "" && isGenericWeComContentType(contentType) {
		if byExt := normalizeWeComContentType(mime.TypeByExtension(ext)); byExt != "" {
			contentType = byExt
		}
	}

	if detectedType != "" {
		switch {
		case contentType == "":
			contentType = detectedType
		case isGenericWeComContentType(contentType):
			contentType = detectedType
		case strings.HasPrefix(detectedType, "image/") && !strings.HasPrefix(contentType, "image/"):
			contentType = detectedType
		case strings.HasPrefix(detectedType, "audio/") && !strings.HasPrefix(contentType, "audio/"):
			contentType = detectedType
		case strings.HasPrefix(detectedType, "video/") && !strings.HasPrefix(contentType, "video/"):
			contentType = detectedType
		}
	}

	if contentType == "" && ext != "" {
		contentType = normalizeWeComContentType(mime.TypeByExtension(ext))
	}
	if contentType == "" {
		contentType = normalizeWeComContentType(http.DetectContentType(data))
	}

	if ext == "" {
		ext = detectedExt
	}
	if ext == "" && contentType != "" {
		if exts, err := mime.ExtensionsByType(contentType); err == nil && len(exts) > 0 {
			ext = strings.ToLower(exts[0])
		}
	}

	if filepath.Ext(filename) == "" && ext != "" {
		filename += ext
	}
	return filename, contentType
}

func (c *WeComChannel) storeRemoteMedia(
	ctx context.Context,
	scope, msgID, resourceURL, aesKey, fallbackExt string,
) (string, error) {
	store := c.GetMediaStore()
	if store == nil {
		return "", fmt.Errorf("no media store available")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resourceURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	resp, err := c.mediaClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download media: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download media returned HTTP %d", resp.StatusCode)
	}

	const maxSize = 20 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSize+1))
	if err != nil {
		return "", fmt.Errorf("read media: %w", err)
	}
	if len(data) > maxSize {
		return "", fmt.Errorf("media too large")
	}

	if aesKey != "" {
		key, keyErr := decodeMediaAESKey(aesKey)
		if keyErr != nil {
			return "", keyErr
		}
		data, err = decryptAESCBC(key, data)
		if err != nil {
			return "", fmt.Errorf("decrypt media: %w", err)
		}
	}

	filename, contentType := detectWeComMediaMetadata(
		data,
		msgID+fallbackExt,
		resp.Header.Get("Content-Type"),
		resourceURL,
		resp.Header.Get("Content-Disposition"),
	)
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = inferMediaExt(contentType, fallbackExt)
	}
	mediaDir := filepath.Join(os.TempDir(), "picoclaw_media")
	if mkdirErr := os.MkdirAll(mediaDir, 0o700); mkdirErr != nil {
		return "", fmt.Errorf("mkdir media dir: %w", mkdirErr)
	}
	tmpFile, err := os.CreateTemp(mediaDir, msgID+"-*"+ext)
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	if _, writeErr := tmpFile.Write(data); writeErr != nil {
		tmpFile.Close()
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("write temp file: %w", writeErr)
	}
	if closeErr := tmpFile.Close(); closeErr != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("close temp file: %w", closeErr)
	}

	ref, err := store.Store(tmpPath, media.MediaMeta{
		Filename:      filename,
		ContentType:   contentType,
		Source:        "wecom",
		CleanupPolicy: media.CleanupPolicyDeleteOnCleanup,
	}, scope)
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	return ref, nil
}
