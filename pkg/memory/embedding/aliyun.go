package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultAliyunBaseURL = "https://dashscope.aliyuncs.com/api/v1"
	defaultAliyunModel   = "text-embedding-v3"
)

// AliyunProvider calls the Alibaba Cloud DashScope (Bailian) embeddings API.
var _ Provider = (*AliyunProvider)(nil)

// AliyunProvider implements Provider using Aliyun DashScope text-embedding models.
type AliyunProvider struct {
	apiKey  string
	baseURL string
	model   string
	dims    int
	client  *http.Client
}

// NewAliyunProvider creates an Aliyun DashScope embedding provider.
// Required: cfg.APIKey. Optional: cfg.Model (default: text-embedding-v3).
func NewAliyunProvider(cfg Config) (*AliyunProvider, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("embedding/aliyun: api_key is required")
	}
	base := cfg.BaseURL
	if base == "" {
		base = defaultAliyunBaseURL
	}
	model := cfg.Model
	if model == "" {
		model = defaultAliyunModel
	}
	return &AliyunProvider{
		apiKey:  cfg.APIKey,
		baseURL: base,
		model:   model,
		client:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Model returns the model identifier.
func (p *AliyunProvider) Model() string { return p.model }

// Dims returns the cached embedding dimension (0 until first call).
func (p *AliyunProvider) Dims() int { return p.dims }

// Embed calls DashScope's text-embedding API and returns one vector per input.
// DashScope limits batch size to 25 texts; we chunk automatically.
func (p *AliyunProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	const batchLimit = 25
	all := make([][]float32, len(texts))

	for start := 0; start < len(texts); start += batchLimit {
		end := start + batchLimit
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[start:end]

		vectors, err := p.embedBatch(ctx, batch)
		if err != nil {
			return nil, err
		}
		copy(all[start:], vectors)
	}

	if p.dims == 0 && len(all) > 0 && len(all[0]) > 0 {
		p.dims = len(all[0])
	}

	return all, nil
}

func (p *AliyunProvider) embedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := aliyunEmbedRequest{
		Model: p.model,
		Input: aliyunEmbedInput{Texts: texts},
		Parameters: aliyunEmbedParams{
			TextType: "query",
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("embedding/aliyun: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/services/embeddings/text-embedding/text-embedding", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embedding/aliyun: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding/aliyun: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr aliyunErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		return nil, fmt.Errorf("embedding/aliyun: API error %d: %s", resp.StatusCode, apiErr.Message)
	}

	var result aliyunEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("embedding/aliyun: decode response: %w", err)
	}

	vectors := make([][]float32, len(texts))
	for _, item := range result.Output.Embeddings {
		if item.TextIndex < len(vectors) {
			vectors[item.TextIndex] = item.Embedding
		}
	}
	return vectors, nil
}

// — request/response types —

type aliyunEmbedRequest struct {
	Model      string             `json:"model"`
	Input      aliyunEmbedInput   `json:"input"`
	Parameters aliyunEmbedParams  `json:"parameters"`
}

type aliyunEmbedInput struct {
	Texts []string `json:"texts"`
}

type aliyunEmbedParams struct {
	TextType string `json:"text_type"`
}

type aliyunEmbedResponse struct {
	Output struct {
		Embeddings []struct {
			TextIndex int       `json:"text_index"`
			Embedding []float32 `json:"embedding"`
		} `json:"embeddings"`
	} `json:"output"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

type aliyunErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}
