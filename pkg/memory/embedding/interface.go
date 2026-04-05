// Package embedding provides a uniform interface for text embedding providers
// used to generate semantic vectors for memory similarity search.
//
// All providers are optional — if none is configured the NoopProvider is used
// and Supabase falls back to text-based search.
package embedding

import "context"

// Provider generates dense vector embeddings for text inputs.
// Implementations must be safe for concurrent use.
type Provider interface {
	// Embed generates embeddings for a batch of texts.
	// Returns one vector per input text, in the same order.
	// Each vector has Dims() elements.
	Embed(ctx context.Context, texts []string) ([][]float32, error)

	// Model returns the canonical model identifier used by this provider.
	// Example: "text-embedding-3-small", "nomic-embed-text", "text-embedding-v3"
	Model() string

	// Dims returns the dimensionality of the embedding vectors produced.
	// Returns 0 if the dimension is unknown before the first call.
	Dims() int
}

// Config holds configuration for building an embedding provider.
type Config struct {
	// Backend selects which provider to use. One of:
	//   "none" (default) — no-op, disables semantic search
	//   "openai"
	//   "ollama"
	//   "google"
	//   "aliyun"
	//   "deepseek"
	Backend string `json:"backend" yaml:"backend"`

	// Model overrides the default model for the selected backend.
	Model string `json:"model" yaml:"model"`

	// APIKey for cloud providers (OpenAI, Google, Aliyun, DeepSeek).
	APIKey string `json:"api_key" yaml:"api_key"`

	// BaseURL overrides the endpoint (for Azure, Ollama, or custom deployments).
	BaseURL string `json:"base_url" yaml:"base_url"`

	// CacheSize sets the max entries in the in-memory LRU embedding cache.
	// Default: 10000. Set 0 to disable.
	CacheSize int `json:"cache_size" yaml:"cache_size"`

	// TextType controls asymmetric embedding roles for providers that support it
	// (currently: Aliyun DashScope). Supported values: "document" (default, for
	// stored corpus texts) or "query" (for search queries). Leave empty for
	// providers that do not support this parameter.
	TextType string `json:"text_type,omitempty" yaml:"text_type"`
}

// New creates a Provider from the given config.
// Returns NoopProvider if backend is "" or "none".
// When cfg.CacheSize > 0, the provider is automatically wrapped with an LRU cache.
// Returns an error only if the backend name is unknown or required fields are missing.
func New(cfg Config) (Provider, error) {
	var p Provider
	var err error

	switch cfg.Backend {
	case "", "none":
		p = &NoopProvider{}
	case "openai":
		p, err = NewOpenAIProvider(cfg)
	case "ollama":
		p, err = NewOllamaProvider(cfg)
	case "google":
		p, err = NewGoogleProvider(cfg)
	case "aliyun":
		p, err = NewAliyunProvider(cfg)
	case "deepseek":
		p, err = NewDeepSeekProvider(cfg)
	default:
		return nil, &UnknownBackendError{Backend: cfg.Backend}
	}

	if err != nil {
		return nil, err
	}

	// Apply LRU cache if requested (default: disabled for noop, caller-controlled otherwise).
	if cfg.CacheSize > 0 && cfg.Backend != "" && cfg.Backend != "none" {
		p = NewCachedProvider(p, cfg.CacheSize)
	}

	return p, nil
}

// UnknownBackendError is returned by New when an unrecognized backend name is given.
type UnknownBackendError struct {
	Backend string
}

func (e *UnknownBackendError) Error() string {
	return "embedding: unknown backend " + e.Backend
}
