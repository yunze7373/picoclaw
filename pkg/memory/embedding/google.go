package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

const (
	defaultGoogleBaseURL = "https://generativelanguage.googleapis.com/v1beta"
	defaultGoogleModel   = "text-embedding-004"
)

// GoogleProvider calls the Google AI Studio (Gemini) embeddings API.
// For Vertex AI, set BaseURL to the Vertex endpoint — Bearer token auth is
// used automatically when the URL contains "aiplatform.googleapis.com".
var _ Provider = (*GoogleProvider)(nil)

// GoogleProvider implements Provider using Google's text embedding models.
type GoogleProvider struct {
	apiKey    string
	baseURL   string
	model     string
	taskType  string // Google task_type enum value; empty = server default
	useBearer bool   // true for Vertex AI; false for AI Studio (?key= param)
	dims      atomic.Int32
	client    *http.Client
}

// NewGoogleProvider creates a Google embedding provider.
// Required: cfg.APIKey (Google AI Studio key or Vertex AI Bearer token).
// Optional: cfg.Model (default: text-embedding-004), cfg.BaseURL, cfg.TextType.
//
// Vertex AI: set BaseURL to the Vertex publisher root (e.g.
// https://us-central1-aiplatform.googleapis.com/v1/projects/PROJECT/locations/us-central1/publishers/google).
// Bearer token auth is used automatically when BaseURL contains "aiplatform.googleapis.com".
//
// WARNING (Vertex AI): cfg.APIKey holds a short-lived OAuth2 access token (~1h TTL).
// Long-running processes must rotate the token externally. Consider Workload Identity
// Federation or a custom http.RoundTripper with oauth2.TokenSource for production.
//
// TextType values: "document" → RETRIEVAL_DOCUMENT, "query" → RETRIEVAL_QUERY.
// Empty string uses the server default (SEMANTIC_SIMILARITY).
func NewGoogleProvider(cfg Config) (*GoogleProvider, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("embedding/google: api_key is required")
	}
	base := cfg.BaseURL
	if base == "" {
		base = defaultGoogleBaseURL
	}
	model := cfg.Model
	if model == "" {
		model = defaultGoogleModel
	}
	useBearer := strings.Contains(base, "aiplatform.googleapis.com")
	taskType := googleTaskType(cfg.TextType)
	return &GoogleProvider{
		apiKey:    cfg.APIKey,
		baseURL:   base,
		model:     model,
		taskType:  taskType,
		useBearer: useBearer,
		client:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// googleTaskType maps TextType values to Google's task_type enum strings.
func googleTaskType(textType string) string {
	switch strings.ToLower(textType) {
	case "document":
		return "RETRIEVAL_DOCUMENT"
	case "query":
		return "RETRIEVAL_QUERY"
	default:
		return "" // server picks SEMANTIC_SIMILARITY
	}
}

// Model returns the model identifier.
func (p *GoogleProvider) Model() string { return p.model }

// Dims returns the cached embedding dimension (0 until first call).
func (p *GoogleProvider) Dims() int { return int(p.dims.Load()) }

// Embed calls Google's batchEmbedContents endpoint and returns vectors.
func (p *GoogleProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	requests := make([]googleEmbedContentRequest, len(texts))
	for i, t := range texts {
		requests[i] = googleEmbedContentRequest{
			Model: "models/" + p.model,
			Content: googleContent{
				Parts: []googlePart{{Text: t}},
			},
			TaskType: p.taskType,
		}
	}

	reqBody := googleBatchEmbedRequest{Requests: requests}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("embedding/google: marshal request: %w", err)
	}

	var reqURL string
	if p.useBearer {
		// Vertex AI: Bearer token in Authorization header; no ?key= query param.
		// baseURL is the publisher root (e.g. .../publishers/google);
		// the model path is appended here to form the full endpoint.
		reqURL = fmt.Sprintf("%s/models/%s:batchEmbedContents", p.baseURL, p.model)
	} else {
		// AI Studio: API key as query parameter.
		reqURL = fmt.Sprintf("%s/models/%s:batchEmbedContents?key=%s", p.baseURL, p.model, p.apiKey)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embedding/google: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if p.useBearer {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		// url.Error.Error() embeds the full URL including ?key=... — redact it.
		var ue *url.Error
		if errors.As(err, &ue) {
			return nil, fmt.Errorf("embedding/google: http %s: %w", ue.Op, ue.Err)
		}
		return nil, fmt.Errorf("embedding/google: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr googleErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		return nil, fmt.Errorf("embedding/google: API error %d: %s", resp.StatusCode, apiErr.Error.Message)
	}

	var result googleBatchEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("embedding/google: decode response: %w", err)
	}

	vectors := make([][]float32, len(texts))
	for i, emb := range result.Embeddings {
		if i < len(vectors) {
			vectors[i] = emb.Values
		}
	}

	if p.dims.Load() == 0 && len(vectors) > 0 && len(vectors[0]) > 0 {
		p.dims.Store(int32(len(vectors[0])))
	}

	return vectors, nil
}

// — request/response types —

type googleBatchEmbedRequest struct {
	Requests []googleEmbedContentRequest `json:"requests"`
}

type googleEmbedContentRequest struct {
	Model    string        `json:"model"`
	Content  googleContent `json:"content"`
	TaskType string        `json:"taskType,omitempty"` // RETRIEVAL_DOCUMENT, RETRIEVAL_QUERY, etc.
}

type googleContent struct {
	Parts []googlePart `json:"parts"`
}

type googlePart struct {
	Text string `json:"text"`
}

type googleBatchEmbedResponse struct {
	Embeddings []struct {
		Values []float32 `json:"values"`
	} `json:"embeddings"`
}

type googleErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
