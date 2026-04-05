package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	defaultOpenAIBaseURL = "https://api.openai.com/v1"
	defaultOpenAIModel   = "text-embedding-3-small"
)

// OpenAIProvider calls the OpenAI (or compatible) Embeddings API.
// Supports Azure OpenAI and any OpenAI-compatible endpoint via BaseURL.
var _ Provider = (*OpenAIProvider)(nil)

// OpenAIProvider implements Provider using the OpenAI embeddings endpoint.
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	model   string
	dims    atomic.Int32
	client  *http.Client
}

// NewOpenAIProvider creates an OpenAI embedding provider from the given config.
// Required: cfg.APIKey. Optional: cfg.Model (default: text-embedding-3-small),
// cfg.BaseURL (default: https://api.openai.com/v1).
func NewOpenAIProvider(cfg Config) (*OpenAIProvider, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("embedding/openai: api_key is required")
	}
	base := cfg.BaseURL
	if base == "" {
		base = defaultOpenAIBaseURL
	}
	model := cfg.Model
	if model == "" {
		model = defaultOpenAIModel
	}
	return &OpenAIProvider{
		apiKey:  cfg.APIKey,
		baseURL: base,
		model:   model,
		client:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Model returns the model identifier.
func (p *OpenAIProvider) Model() string { return p.model }

// Dims returns the cached embedding dimension (0 until first call).
func (p *OpenAIProvider) Dims() int { return int(p.dims.Load()) }

// Embed calls the OpenAI Embeddings API and returns one vector per input.
func (p *OpenAIProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	reqBody := openAIEmbedRequest{
		Input:          texts,
		Model:          p.model,
		EncodingFormat: "float",
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("embedding/openai: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embedding/openai: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding/openai: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr openAIErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		return nil, fmt.Errorf("embedding/openai: API error %d: %s", resp.StatusCode, apiErr.Error.Message)
	}

	var result openAIEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("embedding/openai: decode response: %w", err)
	}

	// Sort data by index to preserve input order
	vectors := make([][]float32, len(texts))
	for _, d := range result.Data {
		if d.Index < len(vectors) {
			vectors[d.Index] = d.Embedding
		}
	}

	// Cache dims from first successful response
	if p.dims.Load() == 0 && len(vectors) > 0 && len(vectors[0]) > 0 {
		p.dims.Store(int32(len(vectors[0])))
	}

	return vectors, nil
}

// — request/response types —

type openAIEmbedRequest struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format"`
}

type openAIEmbedResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
		Object    string    `json:"object"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}
