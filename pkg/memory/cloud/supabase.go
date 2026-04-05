package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
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
	baseURL    string // e.g. "https://xxx.supabase.co"
	apiKey     string
	tableName  string
	client     *http.Client
	mu         sync.RWMutex
	closed     bool
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
}

var _ CloudMemoryStore = (*SupabaseStore)(nil)

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

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &SupabaseStore{
		baseURL:   strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:    cfg.APIKey,
		tableName: tableName,
		client:    &http.Client{Timeout: timeout},
	}, nil
}

func (s *SupabaseStore) UpsertMemory(ctx context.Context, m Memory) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return fmt.Errorf("supabase: store is closed")
	}
	s.mu.RUnlock()

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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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

	url := fmt.Sprintf("%s/rest/v1/%s?id=eq.%s", s.baseURL, s.tableName, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("supabase: create request: %w", err)
	}

	s.setAuthHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("supabase: delete: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return s.readError(resp)
	}
	return nil
}

// SimilaritySearch uses Supabase's RPC endpoint to call a pgvector similarity function.
//
// Expected Supabase function:
//
//	CREATE OR REPLACE FUNCTION match_memories(query_text TEXT, match_count INT, min_similarity FLOAT)
//	RETURNS TABLE (id TEXT, content TEXT, session_key TEXT, kind TEXT, similarity FLOAT)
//	AS $$ ... $$ LANGUAGE plpgsql;
func (s *SupabaseStore) SimilaritySearch(ctx context.Context, query string, topK int, minScore float64) ([]SearchResult, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, fmt.Errorf("supabase: store is closed")
	}
	s.mu.RUnlock()

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
func (s *SupabaseStore) setAuthHeaders(req *http.Request) {
	req.Header.Set("apikey", s.apiKey)
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
}

// readError reads an error response body and returns a formatted error.
func (s *SupabaseStore) readError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("supabase: HTTP %d: %s", resp.StatusCode, string(body))
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
