package embedding

import (
	"context"
	"errors"
)

const (
	defaultDeepSeekBaseURL = "https://api.deepseek.com/v1"
	defaultDeepSeekModel   = "deepseek-embedding"
)

// DeepSeekProvider calls the DeepSeek Embeddings API.
// DeepSeek uses an OpenAI-compatible REST API, so this is a thin wrapper.
var _ Provider = (*DeepSeekProvider)(nil)

// DeepSeekProvider wraps OpenAIProvider with DeepSeek defaults.
type DeepSeekProvider struct {
	inner *OpenAIProvider
}

// NewDeepSeekProvider creates a DeepSeek embedding provider.
// Required: cfg.APIKey. Optional: cfg.Model (default: deepseek-embedding).
func NewDeepSeekProvider(cfg Config) (*DeepSeekProvider, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("embedding/deepseek: api_key is required")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultDeepSeekBaseURL
	}
	if cfg.Model == "" {
		cfg.Model = defaultDeepSeekModel
	}
	inner, err := NewOpenAIProvider(cfg)
	if err != nil {
		return nil, err
	}
	return &DeepSeekProvider{inner: inner}, nil
}

// Model returns the model identifier.
func (p *DeepSeekProvider) Model() string { return p.inner.Model() }

// Dims returns the cached embedding dimension.
func (p *DeepSeekProvider) Dims() int { return p.inner.Dims() }

// Embed delegates to the underlying OpenAI-compatible provider.
func (p *DeepSeekProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	return p.inner.Embed(ctx, texts)
}
