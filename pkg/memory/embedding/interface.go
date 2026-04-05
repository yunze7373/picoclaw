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
}

// New creates a Provider from the given config.
// Returns NoopProvider if backend is "" or "none".
// Returns an error only if the backend name is unknown.
func New(cfg Config) (Provider, error) {
	switch cfg.Backend {
	case "", "none":
		return &NoopProvider{}, nil
	case "openai":
		return NewOpenAIProvider(cfg)
	case "ollama":
		return NewOllamaProvider(cfg)
	case "google":
		return NewGoogleProvider(cfg)
	case "aliyun":
		return NewAliyunProvider(cfg)
	case "deepseek":
		return NewDeepSeekProvider(cfg)
	default:
		return nil, &UnknownBackendError{Backend: cfg.Backend}
	}
}

// UnknownBackendError is returned by New when an unrecognized backend name is given.
type UnknownBackendError struct {
	Backend string
}

func (e *UnknownBackendError) Error() string {
	return "embedding: unknown backend " + e.Backend
}
