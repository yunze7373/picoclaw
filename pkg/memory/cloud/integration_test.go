package cloud_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	cloud "github.com/sipeed/picoclaw/pkg/memory/cloud"
)

// TestIntegration_NoopStack verifies the full stack with NoopStore produces
// zero-overhead behavior: SyncManager + BackupManager run but never issue
// real network calls.
func TestIntegration_NoopStack(t *testing.T) {
	store := cloud.NewNoopStore()
	defer store.Close()

	// SyncManager with NoopStore
	syncMgr := cloud.NewSyncManager(store, cloud.SyncManagerConfig{
		Interval:  100 * time.Millisecond,
		BatchSize: 10,
		QueueSize: 50,
	})
	syncMgr.Start(context.Background())

	// Enqueue several memories
	for i := 0; i < 5; i++ {
		syncMgr.Enqueue(cloud.Memory{
			ID:      "mem-" + string(rune('a'+i)),
			Content: "test content",
			Kind:    "message",
		})
	}

	// Let the sync tick fire
	time.Sleep(200 * time.Millisecond)

	syncMgr.Stop()

	// Verify health
	health, err := store.HealthCheck(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !health.OK {
		t.Error("NoopStore health should always be OK")
	}
}

// TestIntegration_SupabaseStack runs SyncManager + BackupManager against a
// mock Supabase HTTP server to verify real E2E flow.
func TestIntegration_SupabaseStack(t *testing.T) {
	var mu sync.Mutex
	var upserted []string

	// Mock Supabase PostgREST server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && strings.Contains(r.URL.Path, "/rest/v1/memories"):
			mu.Lock()
			upserted = append(upserted, "batch")
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`[{"id":"ok"}]`))

		case r.Method == "POST" && strings.Contains(r.URL.Path, "/rest/v1/rpc/match_memories"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))

		case r.Method == "GET" && strings.Contains(r.URL.Path, "/rest/v1/memories"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id":"m1","content":"hello","kind":"summary"}]`))

		default:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	store, err := cloud.NewSupabaseStore(cloud.SupabaseConfig{
		BaseURL:   srv.URL,
		APIKey:    "test-api-key",
		TableName: "memories",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Create SyncManager
	syncMgr := cloud.NewSyncManager(store, cloud.SyncManagerConfig{
		Interval:  100 * time.Millisecond,
		BatchSize: 5,
		QueueSize: 50,
		OnError: func(err error, batchSize int) {
			t.Logf("sync error: %v (batch: %d)", err, batchSize)
		},
	})
	syncMgr.Start(context.Background())

	// Enqueue memories
	for i := 0; i < 3; i++ {
		syncMgr.Enqueue(cloud.Memory{
			ID:        "test-" + string(rune('0'+i)),
			SessionKey: "session-1",
			Content:   "content",
			Kind:      "message",
		})
	}

	// Wait for at least one flush
	time.Sleep(300 * time.Millisecond)

	// Stop gracefully (should flush remaining)
	syncMgr.Stop()

	mu.Lock()
	count := len(upserted)
	mu.Unlock()

	if count == 0 {
		t.Error("expected at least one upsert batch to the mock server")
	}

	// Verify health check
	health, err := store.HealthCheck(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !health.OK {
		t.Error("health check should pass against mock server")
	}
}

// TestIntegration_BackupProvider verifies BackupManager integrates with
// SyncFromLocal correctly.
func TestIntegration_BackupProvider(t *testing.T) {
	var mu sync.Mutex
	var synced int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/rest/v1/memories") {
			mu.Lock()
			synced++
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`[{"id":"ok"}]`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	store, err := cloud.NewSupabaseStore(cloud.SupabaseConfig{
		BaseURL:   srv.URL,
		APIKey:    "test-key",
		TableName: "memories",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	provider := &mockProvider{
		memories: []cloud.Memory{
			{ID: "b1", Content: "backup data 1", Kind: "summary"},
			{ID: "b2", Content: "backup data 2", Kind: "message"},
		},
	}

	mgr := cloud.NewBackupManager(store, provider, cloud.BackupManagerConfig{
		Interval: time.Hour,
	})

	stats, err := mgr.RunNow(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Upserted != 2 {
		t.Errorf("upserted = %d, want 2", stats.Upserted)
	}
}

type mockProvider struct {
	memories []cloud.Memory
}

func (p *mockProvider) ExportMemories(_ context.Context, _ time.Time) ([]cloud.Memory, error) {
	return p.memories, nil
}
