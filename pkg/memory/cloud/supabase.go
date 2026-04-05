package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/memory/embedding"
)

// SupabaseStore implements CloudMemoryStore using Supabase REST API with pgvector.
//
// Table schema expected in Supabase:
//
//	CREATE TABLE memories (
//	    id TEXT PRIMARY KEY,
//	    session_key TEXT NOT NULL,
//	    content TEXT NOT NULL,
//	    embedding vector(384),  -- adjust dimension to match your model
//	    kind TEXT NOT NULL DEFAULT 'message',
//	    token_count INTEGER NOT NULL DEFAULT 0,
//	    metadata JSONB DEFAULT '{}',
//	    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
//	    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
//	);
//	CREATE INDEX ON memories USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
type SupabaseStore struct {
	baseURL   string // e.g. "https://xxx.supabase.co"
	apiKey    string
	tableName string
	client    *http.Client
	embedder  embedding.Provider // nil or NoopProvider → text fallback
	mu        sync.RWMutex
	closed    bool
}

// SupabaseConfig holds configuration for the Supabase cloud memory backend.
type SupabaseConfig struct {
	// BaseURL is the Supabase project URL (e.g., "https://xxx.supabase.co").
	BaseURL string `json:"base_url"`

	// APIKey is the Supabase anon or service role key.
	APIKey string `json:"api_key"`

	// TableName overrides the default table name ("memories").
	TableName string `json:"table_name,omitempty"`

	// Timeout for HTTP requests. Default: 30s.
	Timeout time.Duration `json:"timeout,omitempty"`

	// Embedder is an optional embedding provider used to generate dense vectors
	// for semantic similarity search. If nil or NoopProvider, the store uses
	// text-based search (match_memories RPC) as fallback.
	Embedder embedding.Provider
}

var _ CloudMemoryStore = (*SupabaseStore)(nil)

// validTableName restricts table names to safe SQL identifiers.
var validTableName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,62}$`)

// NewSupabaseStore creates a new Supabase-backed cloud memory store.
func NewSupabaseStore(cfg SupabaseConfig) (*SupabaseStore, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("supabase: base_url is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("supabase: api_key is required")
	}

	tableName := cfg.TableName
	if tableName == "" {
		tableName = "memories"
	}
	if !validTableName.MatchString(tableName) {
		return nil, fmt.Errorf("supabase: invalid table name %q", tableName)
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &SupabaseStore{
		baseURL:   strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:    cfg.APIKey,
		tableName: tableName,
		client:    &http.Client{Timeout: timeout},
		embedder:  cfg.Embedder,
	}, nil
}

func (s *SupabaseStore) UpsertMemory(ctx context.Context, m Memory) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return fmt.Errorf("supabase: store is closed")
	}
	s.mu.RUnlock()

	// Generate embedding if provider is set and memory has no embedding yet.
	if s.hasEmbedder() && len(m.Embedding) == 0 {
		vectors, err := s.embedder.Embed(ctx, []string{m.Content})
		if err == nil && len(vectors) > 0 && len(vectors[0]) > 0 {
			m.Embedding = vectors[0]
		}
		// Non-fatal: continue without embedding on error
	}

	body := memoryToRow(m)
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("supabase: marshal memory: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/%s", s.baseURL, s.tableName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("supabase: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "resolution=merge-duplicates")
	s.setAuthHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("supabase: upsert: %w", err)
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode >= 300 {
		return s.readError(resp)
	}
	return nil
}

func (s *SupabaseStore) UpsertBatch(ctx context.Context, memories []Memory) (int, error) {
	if len(memories) == 0 {
		return 0, nil
	}

	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return 0, fmt.Errorf("supabase: store is closed")
	}
	s.mu.RUnlock()

	rows := make([]map[string]any, len(memories))

	// Batch-embed memories that lack embeddings.
	if s.hasEmbedder() {
		needsEmbed := make([]int, 0, len(memories))
		texts := make([]string, 0, len(memories))
		for i, m := range memories {
			if len(m.Embedding) == 0 {
				needsEmbed = append(needsEmbed, i)
				texts = append(texts, m.Content)
			}
		}
		if len(texts) > 0 {
			if vectors, err := s.embedder.Embed(ctx, texts); err == nil {
				for j, idx := range needsEmbed {
					if j < len(vectors) && len(vectors[j]) > 0 {
						memories[idx].Embedding = vectors[j]
					}
				}
			}
			// Non-fatal: proceed without embeddings on error
		}
	}

	for i, m := range memories {
		rows[i] = memoryToRow(m)
	}

	data, err := json.Marshal(rows)
	if err != nil {
		return 0, fmt.Errorf("supabase: marshal batch: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/%s", s.baseURL, s.tableName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("supabase: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "resolution=merge-duplicates")
	s.setAuthHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("supabase: upsert batch: %w", err)
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode >= 300 {
		return 0, s.readError(resp)
	}
	return len(memories), nil
}

func (s *SupabaseStore) DeleteMemory(ctx context.Context, id string) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return fmt.Errorf("supabase: store is closed")
	}
	s.mu.RUnlock()

	if id == "" {
		return fmt.Errorf("supabase: id is required for delete")
	}

	u, err := url.Parse(fmt.Sprintf("%s/rest/v1/%s", s.baseURL, s.tableName))
	if err != nil {
		return fmt.Errorf("supabase: build URL: %w", err)
	}
	q := u.Query()
	q.Set("id", "eq."+id)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
	if err != nil {
		return fmt.Errorf("supabase: create request: %w", err)
	}

	s.setAuthHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("supabase: delete: %w", err)
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode >= 300 {
		return s.readError(resp)
	}
	return nil
}

// SimilaritySearch uses Supabase's RPC endpoint to call a pgvector similarity function.
//
// When an embedding provider is configured, it uses the vector-based RPC:
//
//	CREATE OR REPLACE FUNCTION match_memories_vector(query_embedding vector, match_count INT, min_similarity FLOAT)
//	RETURNS TABLE (id TEXT, content TEXT, session_key TEXT, kind TEXT, similarity FLOAT)
//
// Without an embedding provider, it falls back to the text-based RPC:
//
//	CREATE OR REPLACE FUNCTION match_memories(query_text TEXT, match_count INT, min_similarity FLOAT)
//	RETURNS TABLE (id TEXT, content TEXT, session_key TEXT, kind TEXT, similarity FLOAT)
func (s *SupabaseStore) SimilaritySearch(ctx context.Context, query string, topK int, minScore float64) ([]SearchResult, error) {
	if topK <= 0 {
		return nil, fmt.Errorf("supabase: topK must be positive, got %d", topK)
	}
	if minScore < 0 || minScore > 1 {
		return nil, fmt.Errorf("supabase: minScore must be in [0,1], got %f", minScore)
	}
	if query == "" {
		return nil, nil
	}

	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, fmt.Errorf("supabase: store is closed")
	}
	s.mu.RUnlock()

	// Use vector search if embedder is configured.
	if s.hasEmbedder() {
		vectors, err := s.embedder.Embed(ctx, []string{query})
		if err == nil && len(vectors) > 0 && len(vectors[0]) > 0 {
			return s.vectorSearch(ctx, vectors[0], topK, minScore)
		}
		// Fall through to text search on embedding failure
	}

	return s.textSearch(ctx, query, topK, minScore)
}

func (s *SupabaseStore) vectorSearch(ctx context.Context, queryVec []float32, topK int, minScore float64) ([]SearchResult, error) {
	payload := map[string]any{
		"query_embedding": queryVec,
		"match_count":     topK,
		"min_similarity":  minScore,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("supabase: marshal vector query: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/match_memories_vector", s.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("supabase: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	s.setAuthHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("supabase: vector search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, s.readError(resp)
	}
	return s.decodeSearchResults(resp)
}

func (s *SupabaseStore) textSearch(ctx context.Context, query string, topK int, minScore float64) ([]SearchResult, error) {
	payload := map[string]any{
		"query_text":     query,
		"match_count":    topK,
		"min_similarity": minScore,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("supabase: marshal query: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/match_memories", s.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("supabase: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	s.setAuthHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("supabase: similarity search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, s.readError(resp)
	}
	return s.decodeSearchResults(resp)
}

func (s *SupabaseStore) decodeSearchResults(resp *http.Response) ([]SearchResult, error) {
	var rpcResults []struct {
		ID         string  `json:"id"`
		Content    string  `json:"content"`
		SessionKey string  `json:"session_key"`
		Kind       string  `json:"kind"`
		Similarity float64 `json:"similarity"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpcResults); err != nil {
		return nil, fmt.Errorf("supabase: decode results: %w", err)
	}

	results := make([]SearchResult, len(rpcResults))
	for i, r := range rpcResults {
		results[i] = SearchResult{
			Memory: Memory{
				ID:         r.ID,
				SessionKey: r.SessionKey,
				Content:    r.Content,
				Kind:       r.Kind,
			},
			Similarity: r.Similarity,
		}
	}
	return results, nil
}

// hasEmbedder returns true if a real (non-noop) embedding provider is set.
func (s *SupabaseStore) hasEmbedder() bool {
	return s.embedder != nil && s.embedder.Dims() != 0 || (s.embedder != nil && s.embedder.Model() != "none")
}

func (s *SupabaseStore) SyncFromLocal(ctx context.Context, memories []Memory) (*SyncStats, error) {
	start := time.Now()
	n, err := s.UpsertBatch(ctx, memories)
	stats := &SyncStats{
		Upserted: n,
		Duration: time.Since(start),
	}
	if err != nil {
		stats.Errors = len(memories) - n
		return stats, err
	}
	return stats, nil
}

func (s *SupabaseStore) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	start := time.Now()

	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return &HealthStatus{
			OK:           false,
			Backend:      "supabase",
			ErrorMessage: "store is closed",
		}, nil
	}
	s.mu.RUnlock()

	// Simple query to check connectivity
	url := fmt.Sprintf("%s/rest/v1/%s?select=id&limit=1", s.baseURL, s.tableName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return &HealthStatus{
			OK:           false,
			Backend:      "supabase",
			ErrorMessage: err.Error(),
		}, nil
	}

	s.setAuthHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return &HealthStatus{
			OK:           false,
			Backend:      "supabase",
			Latency:      time.Since(start),
			ErrorMessage: err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	latency := time.Since(start)

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return &HealthStatus{
			OK:           false,
			Backend:      "supabase",
			Latency:      latency,
			ErrorMessage: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(bodyBytes)),
		}, nil
	}

	return &HealthStatus{
		OK:      true,
		Backend: "supabase",
		Latency: latency,
	}, nil
}

func (s *SupabaseStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.client.CloseIdleConnections()
	return nil
}

// setAuthHeaders adds Supabase authorization headers.
// Supabase PostgREST requires both: "apikey" for rate limiting/routing,
// "Authorization: Bearer" for actual authentication.
func (s *SupabaseStore) setAuthHeaders(req *http.Request) {
	req.Header.Set("apikey", s.apiKey)
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
}

// readError reads an error response body and returns a formatted error.
func (s *SupabaseStore) readError(resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil || len(body) == 0 {
		return fmt.Errorf("supabase: HTTP %d: <body unreadable>", resp.StatusCode)
	}
	return fmt.Errorf("supabase: HTTP %d: %s", resp.StatusCode, string(body))
}

// drainAndClose fully reads and closes an HTTP response body to allow
// connection reuse by the HTTP client pool.
func drainAndClose(body io.ReadCloser) {
	io.Copy(io.Discard, body)
	body.Close()
}

// memoryToRow converts a Memory to a Supabase-compatible row map.
func memoryToRow(m Memory) map[string]any {
	row := map[string]any{
		"id":          m.ID,
		"session_key": m.SessionKey,
		"content":     m.Content,
		"kind":        m.Kind,
		"token_count": m.TokenCount,
		"synced_at":   time.Now().UTC().Format(time.RFC3339),
	}
	if !m.CreatedAt.IsZero() {
		row["created_at"] = m.CreatedAt.Format(time.RFC3339)
	}
	if len(m.Embedding) > 0 {
		row["embedding"] = m.Embedding
	}
	if len(m.Metadata) > 0 {
		row["metadata"] = m.Metadata
	}
	return row
}
