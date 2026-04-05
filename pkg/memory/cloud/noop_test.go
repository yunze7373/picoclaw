package cloud

import (
	"context"
	"testing"
)

func TestNoopStore_ImplementsInterface(t *testing.T) {
	var store CloudMemoryStore = &NoopStore{}
	_ = store
}

func TestNoopStore_UpsertMemory(t *testing.T) {
	store := &NoopStore{}
	err := store.UpsertMemory(context.Background(), Memory{ID: "test", Content: "hello"})
	if err != nil {
		t.Fatalf("UpsertMemory should not fail: %v", err)
	}
}

func TestNoopStore_UpsertBatch(t *testing.T) {
	store := &NoopStore{}
	n, err := store.UpsertBatch(context.Background(), []Memory{
		{ID: "a"}, {ID: "b"}, {ID: "c"},
	})
	if err != nil {
		t.Fatalf("UpsertBatch should not fail: %v", err)
	}
	if n != 3 {
		t.Fatalf("UpsertBatch should return count of input: got %d, want 3", n)
	}
}

func TestNoopStore_DeleteMemory(t *testing.T) {
	store := &NoopStore{}
	err := store.DeleteMemory(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("DeleteMemory should not fail: %v", err)
	}
}

func TestNoopStore_SimilaritySearch(t *testing.T) {
	store := &NoopStore{}
	results, err := store.SimilaritySearch(context.Background(), "query", 10, 0.5)
	if err != nil {
		t.Fatalf("SimilaritySearch should not fail: %v", err)
	}
	if results != nil {
		t.Fatalf("SimilaritySearch should return nil: got %v", results)
	}
}

func TestNoopStore_SyncFromLocal(t *testing.T) {
	store := &NoopStore{}
	stats, err := store.SyncFromLocal(context.Background(), []Memory{{ID: "x"}, {ID: "y"}})
	if err != nil {
		t.Fatalf("SyncFromLocal should not fail: %v", err)
	}
	if stats.Upserted != 2 {
		t.Fatalf("SyncFromLocal upserted count: got %d, want 2", stats.Upserted)
	}
}

func TestNoopStore_HealthCheck(t *testing.T) {
	store := &NoopStore{}
	status, err := store.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("HealthCheck should not fail: %v", err)
	}
	if !status.OK {
		t.Fatal("HealthCheck should return OK")
	}
	if status.Backend != "noop" {
		t.Fatalf("HealthCheck backend: got %q, want %q", status.Backend, "noop")
	}
}

func TestNoopStore_Close(t *testing.T) {
	store := &NoopStore{}
	if err := store.Close(); err != nil {
		t.Fatalf("Close should not fail: %v", err)
	}
}
