package embedding

import "context"

// NoopProvider is a zero-overhead embedding provider that always returns
// zero-length zero vectors. Used when no embedding backend is configured.
//
// Compile-time interface check ensures NoopProvider satisfies Provider.
var _ Provider = (*NoopProvider)(nil)

// NoopProvider returns nil vectors (empty slice per text) for every call.
// Supabase store detects nil/empty vectors and falls back to text search.
type NoopProvider struct{}

// Embed returns a nil vector for each input text (no network call, no allocation).
func (p *NoopProvider) Embed(_ context.Context, texts []string) ([][]float32, error) {
	// make([][]float32, n) already initialises each element to nil.
	return make([][]float32, len(texts)), nil
}

// Model returns the noop provider identifier.
func (p *NoopProvider) Model() string { return "none" }

// Dims returns 0 — noop vectors have no dimensions.
func (p *NoopProvider) Dims() int { return 0 }
