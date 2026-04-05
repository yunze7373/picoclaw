package embedding_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sipeed/picoclaw/pkg/memory/embedding"
)

// — NoopProvider tests —

func TestNoopProvider(t *testing.T) {
	p := &embedding.NoopProvider{}

	if p.Model() != "none" {
		t.Errorf("Model() = %q, want %q", p.Model(), "none")
	}
	if p.Dims() != 0 {
		t.Errorf("Dims() = %d, want 0", p.Dims())
	}

	vectors, err := p.Embed(context.Background(), []string{"hello", "world"})
	if err != nil {
		t.Fatalf("Embed() error: %v", err)
	}
	if len(vectors) != 2 {
		t.Errorf("Embed() len = %d, want 2", len(vectors))
	}
	for _, v := range vectors {
		if v != nil {
			t.Errorf("NoopProvider should return nil vectors, got %v", v)
		}
	}
}

func TestNoopProvider_EmptyInput(t *testing.T) {
	p := &embedding.NoopProvider{}
	vectors, err := p.Embed(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(vectors) != 0 {
		t.Errorf("expected 0 vectors for empty input, got %d", len(vectors))
	}
}

// — OpenAI provider tests (mock server) —

func TestOpenAIProvider_Embed(t *testing.T) {
	mockResp := map[string]any{
		"object": "list",
		"data": []map[string]any{
			{"index": 0, "object": "embedding", "embedding": []float32{0.1, 0.2, 0.3}},
			{"index": 1, "object": "embedding", "embedding": []float32{0.4, 0.5, 0.6}},
		},
		"model": "text-embedding-3-small",
		"usage": map[string]any{"prompt_tokens": 4, "total_tokens": 4},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer srv.Close()

	p, err := embedding.NewOpenAIProvider(embedding.Config{
		APIKey:  "test-key",
		BaseURL: srv.URL,
	})
	if err != nil {
		t.Fatalf("NewOpenAIProvider: %v", err)
	}

	vectors, err := p.Embed(context.Background(), []string{"hello", "world"})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(vectors) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(vectors))
	}
	if vectors[0][0] != 0.1 {
		t.Errorf("vectors[0][0] = %v, want 0.1", vectors[0][0])
	}
	if p.Dims() != 3 {
		t.Errorf("Dims() = %d, want 3", p.Dims())
	}
}

func TestOpenAIProvider_EmptyInput(t *testing.T) {
	p, _ := embedding.NewOpenAIProvider(embedding.Config{APIKey: "test"})
	vectors, err := p.Embed(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if vectors != nil {
		t.Errorf("expected nil for empty input, got %v", vectors)
	}
}

func TestOpenAIProvider_MissingAPIKey(t *testing.T) {
	_, err := embedding.NewOpenAIProvider(embedding.Config{})
	if err == nil {
		t.Fatal("expected error for missing api_key")
	}
}

func TestOpenAIProvider_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "rate limit exceeded", "type": "rate_limit_error"},
		})
	}))
	defer srv.Close()

	p, _ := embedding.NewOpenAIProvider(embedding.Config{APIKey: "test", BaseURL: srv.URL})
	_, err := p.Embed(context.Background(), []string{"test"})
	if err == nil {
		t.Fatal("expected error on API failure")
	}
}

// — Ollama provider tests —

func TestOllamaProvider_Embed(t *testing.T) {
	mockResp := map[string]any{
		"model":      "nomic-embed-text",
		"embeddings": [][]float32{{0.1, 0.2}, {0.3, 0.4}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer srv.Close()

	p, err := embedding.NewOllamaProvider(embedding.Config{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("NewOllamaProvider: %v", err)
	}

	vectors, err := p.Embed(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(vectors) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(vectors))
	}
	if p.Dims() != 2 {
		t.Errorf("Dims() = %d, want 2", p.Dims())
	}
}

// — CachedProvider tests —

func TestCachedProvider_CacheHit(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"index": 0, "embedding": []float32{0.1, 0.2, 0.3}},
			},
		})
	}))
	defer srv.Close()

	inner, _ := embedding.NewOpenAIProvider(embedding.Config{APIKey: "test", BaseURL: srv.URL})
	cached := embedding.NewCachedProvider(inner, 100)

	ctx := context.Background()
	texts := []string{"same text"}

	_, err := cached.Embed(ctx, texts)
	if err != nil {
		t.Fatal(err)
	}
	_, err = cached.Embed(ctx, texts)
	if err != nil {
		t.Fatal(err)
	}

	// Second call should use cache — inner called only once
	if callCount != 1 {
		t.Errorf("expected 1 HTTP call (cache hit), got %d", callCount)
	}
}

func TestCachedProvider_LRUEviction(t *testing.T) {
	p := embedding.NewCachedProvider(&embedding.NoopProvider{}, 2)
	ctx := context.Background()

	// Fill cache with 2 entries
	_, _ = p.Embed(ctx, []string{"text1"})
	_, _ = p.Embed(ctx, []string{"text2"})
	// Add a third — should evict "text1"
	_, _ = p.Embed(ctx, []string{"text3"})

	// Verify cache size is bounded (no panic, no OOM)
}

// — New() factory tests —

func TestNew_Noop(t *testing.T) {
	p, err := embedding.New(embedding.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if p.Model() != "none" {
		t.Errorf("expected noop provider, got %q", p.Model())
	}
}

func TestNew_UnknownBackend(t *testing.T) {
	_, err := embedding.New(embedding.Config{Backend: "unknown_backend_xyz"})
	if err == nil {
		t.Fatal("expected error for unknown backend")
	}
}

func TestNew_OpenAI_MissingKey(t *testing.T) {
	_, err := embedding.New(embedding.Config{Backend: "openai"})
	if err == nil {
		t.Fatal("expected error for openai without api_key")
	}
}

// — CacheStats tests —

func TestCachedProvider_Stats(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"index": 0, "embedding": []float32{0.1, 0.2}},
			},
		})
	}))
	defer srv.Close()

	inner, _ := embedding.NewOpenAIProvider(embedding.Config{APIKey: "test", BaseURL: srv.URL})
	cached := embedding.NewCachedProvider(inner, 10)
	ctx := context.Background()

	// First embed: miss
	_, _ = cached.Embed(ctx, []string{"hello"})
	// Second embed: hit
	_, _ = cached.Embed(ctx, []string{"hello"})
	// Third embed: new text, miss
	_, _ = cached.Embed(ctx, []string{"world"})

	stats := cached.Stats()
	if stats.Hits != 1 {
		t.Errorf("Stats.Hits = %d, want 1", stats.Hits)
	}
	if stats.Misses != 2 {
		t.Errorf("Stats.Misses = %d, want 2", stats.Misses)
	}
	if stats.Evictions != 0 {
		t.Errorf("Stats.Evictions = %d, want 0", stats.Evictions)
	}
	if stats.Size != 2 {
		t.Errorf("Stats.Size = %d, want 2", stats.Size)
	}
	if stats.MaxSize != 10 {
		t.Errorf("Stats.MaxSize = %d, want 10", stats.MaxSize)
	}
}

func TestCachedProvider_Stats_Eviction(t *testing.T) {
	p := embedding.NewCachedProvider(&embedding.NoopProvider{}, 2)
	ctx := context.Background()

	_, _ = p.Embed(ctx, []string{"a"})
	_, _ = p.Embed(ctx, []string{"b"})
	_, _ = p.Embed(ctx, []string{"c"}) // evicts "a"

	stats := p.Stats()
	if stats.Evictions != 1 {
		t.Errorf("Stats.Evictions = %d, want 1", stats.Evictions)
	}
	if stats.Size != 2 {
		t.Errorf("Stats.Size = %d, want 2 after eviction", stats.Size)
	}
}

// — Google Vertex AI bearer auth test —

func TestGoogleProvider_VertexAI_Bearer(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		// Ensure no ?key= in URL
		if r.URL.RawQuery != "" {
			t.Errorf("Vertex AI request should have no query params, got: %s", r.URL.RawQuery)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"embeddings": []map[string]any{
				{"values": []float32{0.1, 0.2}},
			},
		})
	}))
	defer srv.Close()

	// Fake a Vertex AI URL by injecting aiplatform.googleapis.com substring
	vertexURL := srv.URL + "/aiplatform.googleapis.com"
	p, err := embedding.NewGoogleProvider(embedding.Config{
		APIKey:  "ya29.faketoken",
		BaseURL: vertexURL,
	})
	if err != nil {
		t.Fatalf("NewGoogleProvider: %v", err)
	}

	_, err = p.Embed(context.Background(), []string{"test"})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if gotAuth != "Bearer ya29.faketoken" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer ya29.faketoken")
	}
}
