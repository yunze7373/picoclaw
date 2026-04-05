package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultOllamaBaseURL = "http://localhost:11434"
	defaultOllamaModel   = "nomic-embed-text"
)

// OllamaProvider calls a local Ollama instance for embeddings.
// Ideal for offline / Termux / airgapped deployments — no API key required.
var _ Provider = (*OllamaProvider)(nil)

// OllamaProvider implements Provider using the Ollama embed endpoint.
type OllamaProvider struct {
	baseURL string
	model   string
	dims    int
	client  *http.Client
}

// NewOllamaProvider creates an Ollama embedding provider.
// Optional: cfg.BaseURL (default: http://localhost:11434),
// cfg.Model (default: nomic-embed-text).
func NewOllamaProvider(cfg Config) (*OllamaProvider, error) {
	base := cfg.BaseURL
	if base == "" {
		base = defaultOllamaBaseURL
	}
	model := cfg.Model
	if model == "" {
		model = defaultOllamaModel
	}
	return &OllamaProvider{
		baseURL: base,
		model:   model,
		client:  &http.Client{Timeout: 60 * time.Second}, // local model may be slower
	}, nil
}

// Model returns the model identifier.
func (p *OllamaProvider) Model() string { return p.model }

// Dims returns the cached embedding dimension (0 until first call).
func (p *OllamaProvider) Dims() int { return p.dims }

// Embed calls Ollama's /api/embed endpoint and returns one vector per input.
// Ollama processes one text at a time; we call sequentially to keep it simple.
func (p *OllamaProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	// Ollama >= 0.1.34 supports batch embed via /api/embed
	reqBody := ollamaEmbedRequest{
		Model: p.model,
		Input: texts,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("embedding/ollama: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embedding/ollama: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding/ollama: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ollamaErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("embedding/ollama: API error %d: %s", resp.StatusCode, errResp.Error)
	}

	var result ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("embedding/ollama: decode response: %w", err)
	}

	if len(result.Embeddings) != len(texts) {
		return nil, fmt.Errorf("embedding/ollama: expected %d vectors, got %d", len(texts), len(result.Embeddings))
	}

	// Cache dims
	if p.dims == 0 && len(result.Embeddings) > 0 {
		p.dims = len(result.Embeddings[0])
	}

	return result.Embeddings, nil
}

// — request/response types —

type ollamaEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type ollamaEmbedResponse struct {
	Model      string      `json:"model"`
	Embeddings [][]float32 `json:"embeddings"`
}

type ollamaErrorResponse struct {
	Error string `json:"error"`
}
