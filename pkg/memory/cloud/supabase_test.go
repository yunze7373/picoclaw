package cloud

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewSupabaseStore_Validation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     SupabaseConfig
		wantErr string
	}{
		{
			name:    "missing base_url",
			cfg:     SupabaseConfig{APIKey: "key"},
			wantErr: "base_url is required",
		},
		{
			name:    "missing api_key",
			cfg:     SupabaseConfig{BaseURL: "https://example.supabase.co"},
			wantErr: "api_key is required",
		},
		{
			name: "valid config",
			cfg:  SupabaseConfig{BaseURL: "https://example.supabase.co", APIKey: "key"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store, err := NewSupabaseStore(tc.cfg)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error %q should contain %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if store == nil {
				t.Fatal("store should not be nil")
			}
		})
	}
}

func TestSupabaseStore_UpsertMemory(t *testing.T) {
	var receivedBody []byte
	var receivedHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	store, err := NewSupabaseStore(SupabaseConfig{
		BaseURL: srv.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = store.UpsertMemory(context.Background(), Memory{
		ID:         "mem-1",
		SessionKey: "sess-1",
		Content:    "hello world",
		Kind:       "message",
		TokenCount: 5,
	})
	if err != nil {
		t.Fatalf("UpsertMemory failed: %v", err)
	}

	// Verify auth headers
	if receivedHeaders.Get("apikey") != "test-key" {
		t.Errorf("missing apikey header")
	}
	if receivedHeaders.Get("Authorization") != "Bearer test-key" {
		t.Errorf("missing Authorization header")
	}
	if receivedHeaders.Get("Prefer") != "resolution=merge-duplicates" {
		t.Errorf("missing Prefer header for upsert")
	}

	// Verify body
	var body map[string]any
	if err := json.Unmarshal(receivedBody, &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["id"] != "mem-1" {
		t.Errorf("id = %v, want mem-1", body["id"])
	}
	if body["content"] != "hello world" {
		t.Errorf("content = %v, want hello world", body["content"])
	}
}

func TestSupabaseStore_UpsertBatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	store, _ := NewSupabaseStore(SupabaseConfig{
		BaseURL: srv.URL,
		APIKey:  "key",
	})

	n, err := store.UpsertBatch(context.Background(), []Memory{
		{ID: "a", Content: "hello"},
		{ID: "b", Content: "world"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("upserted = %d, want 2", n)
	}
}

func TestSupabaseStore_UpsertBatch_Empty(t *testing.T) {
	store, _ := NewSupabaseStore(SupabaseConfig{
		BaseURL: "https://unused.supabase.co",
		APIKey:  "key",
	})

	n, err := store.UpsertBatch(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("upserted = %d, want 0", n)
	}
}

func TestSupabaseStore_DeleteMemory(t *testing.T) {
	var requestURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURL = r.URL.String()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	store, _ := NewSupabaseStore(SupabaseConfig{
		BaseURL: srv.URL,
		APIKey:  "key",
	})

	err := store.DeleteMemory(context.Background(), "mem-1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(requestURL, "id=eq.mem-1") {
		t.Errorf("URL should contain id filter, got: %s", requestURL)
	}
}

func TestSupabaseStore_SimilaritySearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "rpc/match_memories") {
			t.Errorf("expected RPC endpoint, got %s", r.URL.Path)
		}

		resp := []map[string]any{
			{"id": "m1", "content": "related content", "session_key": "s1", "kind": "summary", "similarity": 0.95},
			{"id": "m2", "content": "partial match", "session_key": "s1", "kind": "message", "similarity": 0.72},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	store, _ := NewSupabaseStore(SupabaseConfig{
		BaseURL: srv.URL,
		APIKey:  "key",
	})

	results, err := store.SimilaritySearch(context.Background(), "test query", 5, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Memory.ID != "m1" {
		t.Errorf("first result ID = %q, want m1", results[0].Memory.ID)
	}
	if results[0].Similarity != 0.95 {
		t.Errorf("first result similarity = %f, want 0.95", results[0].Similarity)
	}
}

func TestSupabaseStore_HealthCheck_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("[]"))
	}))
	defer srv.Close()

	store, _ := NewSupabaseStore(SupabaseConfig{
		BaseURL: srv.URL,
		APIKey:  "key",
	})

	status, err := store.HealthCheck(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !status.OK {
		t.Fatal("expected OK")
	}
	if status.Backend != "supabase" {
		t.Errorf("backend = %q, want supabase", status.Backend)
	}
}

func TestSupabaseStore_HealthCheck_Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Invalid API key"}`))
	}))
	defer srv.Close()

	store, _ := NewSupabaseStore(SupabaseConfig{
		BaseURL: srv.URL,
		APIKey:  "bad-key",
	})

	status, err := store.HealthCheck(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.OK {
		t.Fatal("expected not OK")
	}
	if !strings.Contains(status.ErrorMessage, "401") {
		t.Errorf("error should mention status code: %s", status.ErrorMessage)
	}
}

func TestSupabaseStore_ErrorAfterClose(t *testing.T) {
	store, _ := NewSupabaseStore(SupabaseConfig{
		BaseURL: "https://example.supabase.co",
		APIKey:  "key",
	})
	store.Close()

	err := store.UpsertMemory(context.Background(), Memory{ID: "x"})
	if err == nil {
		t.Fatal("expected error after close")
	}
	if !strings.Contains(err.Error(), "closed") {
		t.Errorf("error should mention closed: %v", err)
	}
}

func TestSupabaseStore_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Internal Server Error"}`))
	}))
	defer srv.Close()

	store, _ := NewSupabaseStore(SupabaseConfig{
		BaseURL: srv.URL,
		APIKey:  "key",
	})

	err := store.UpsertMemory(context.Background(), Memory{ID: "x"})
	if err == nil {
		t.Fatal("expected error on 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should mention status code: %v", err)
	}
}

func TestSupabaseStore_SyncFromLocal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	store, _ := NewSupabaseStore(SupabaseConfig{
		BaseURL: srv.URL,
		APIKey:  "key",
	})

	stats, err := store.SyncFromLocal(context.Background(), []Memory{
		{ID: "a"}, {ID: "b"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Upserted != 2 {
		t.Errorf("upserted = %d, want 2", stats.Upserted)
	}
	if stats.Duration <= 0 {
		t.Error("duration should be positive")
	}
}
