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
	defaultGoogleBaseURL = "https://generativelanguage.googleapis.com/v1beta"
	defaultGoogleModel   = "text-embedding-004"
)

// GoogleProvider calls the Google AI Studio (Gemini) embeddings API.
// For Vertex AI, set BaseURL to the Vertex endpoint.
var _ Provider = (*GoogleProvider)(nil)

// GoogleProvider implements Provider using Google's text embedding models.
type GoogleProvider struct {
	apiKey  string
	baseURL string
	model   string
	dims    int
	client  *http.Client
}

// NewGoogleProvider creates a Google embedding provider.
// Required: cfg.APIKey (Google AI Studio key).
// Optional: cfg.Model (default: text-embedding-004), cfg.BaseURL.
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
	return &GoogleProvider{
		apiKey:  cfg.APIKey,
		baseURL: base,
		model:   model,
		client:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Model returns the model identifier.
func (p *GoogleProvider) Model() string { return p.model }

// Dims returns the cached embedding dimension (0 until first call).
func (p *GoogleProvider) Dims() int { return p.dims }

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
		}
	}

	reqBody := googleBatchEmbedRequest{Requests: requests}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("embedding/google: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:batchEmbedContents?key=%s", p.baseURL, p.model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embedding/google: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
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

	if p.dims == 0 && len(vectors) > 0 && len(vectors[0]) > 0 {
		p.dims = len(vectors[0])
	}

	return vectors, nil
}

// — request/response types —

type googleBatchEmbedRequest struct {
	Requests []googleEmbedContentRequest `json:"requests"`
}

type googleEmbedContentRequest struct {
	Model   string        `json:"model"`
	Content googleContent `json:"content"`
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
