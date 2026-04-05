package cloud

import (
	"context"
	"sync"
	"testing"
	"time"
)

type syncTracker struct {
	NoopStore
	mu      sync.Mutex
	synced  []Memory
	calls   int
}

func (t *syncTracker) SyncFromLocal(_ context.Context, memories []Memory) (*SyncStats, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls++
	t.synced = append(t.synced, memories...)
	return &SyncStats{Upserted: len(memories)}, nil
}

func (t *syncTracker) getSynced() []Memory {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]Memory, len(t.synced))
	copy(out, t.synced)
	return out
}

func (t *syncTracker) getCalls() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.calls
}

func TestSyncManager_BasicFlush(t *testing.T) {
	tracker := &syncTracker{}
	mgr := NewSyncManager(tracker, SyncManagerConfig{
		Interval:  50 * time.Millisecond,
		BatchSize: 10,
		QueueSize: 100,
	})

	ctx := context.Background()
	mgr.Start(ctx)

	mgr.Enqueue(
		Memory{ID: "a", Content: "hello"},
		Memory{ID: "b", Content: "world"},
	)

	// Wait for timer-based flush
	time.Sleep(150 * time.Millisecond)

	mgr.Stop()

	synced := tracker.getSynced()
	if len(synced) != 2 {
		t.Fatalf("expected 2 synced memories, got %d", len(synced))
	}
	if synced[0].ID != "a" || synced[1].ID != "b" {
		t.Errorf("unexpected synced IDs: %v", synced)
	}
}

func TestSyncManager_BatchSizeFlush(t *testing.T) {
	tracker := &syncTracker{}
	mgr := NewSyncManager(tracker, SyncManagerConfig{
		Interval:  10 * time.Second, // long interval — shouldn't trigger
		BatchSize: 3,
		QueueSize: 100,
	})

	ctx := context.Background()
	mgr.Start(ctx)

	mgr.Enqueue(
		Memory{ID: "1"},
		Memory{ID: "2"},
		Memory{ID: "3"}, // triggers batch flush
	)

	// Give time for flush
	time.Sleep(50 * time.Millisecond)

	mgr.Stop()

	synced := tracker.getSynced()
	if len(synced) < 3 {
		t.Fatalf("expected at least 3 synced, got %d", len(synced))
	}
}

func TestSyncManager_StopFlushesRemaining(t *testing.T) {
	tracker := &syncTracker{}
	mgr := NewSyncManager(tracker, SyncManagerConfig{
		Interval:  10 * time.Second, // won't trigger
		BatchSize: 100,              // won't trigger
		QueueSize: 100,
	})

	ctx := context.Background()
	mgr.Start(ctx)

	mgr.Enqueue(Memory{ID: "pending"})
	time.Sleep(20 * time.Millisecond)

	mgr.Stop()

	synced := tracker.getSynced()
	if len(synced) != 1 {
		t.Fatalf("expected 1 synced on stop, got %d", len(synced))
	}
	if synced[0].ID != "pending" {
		t.Errorf("expected pending memory, got %s", synced[0].ID)
	}
}

func TestSyncManager_DoubleStop(t *testing.T) {
	tracker := &syncTracker{}
	mgr := NewSyncManager(tracker, SyncManagerConfig{
		Interval: 100 * time.Millisecond,
	})

	mgr.Start(context.Background())
	mgr.Stop()
	mgr.Stop() // should not panic
}

func TestSyncManagerConfig_Defaults(t *testing.T) {
	cfg := SyncManagerConfig{}
	cfg = cfg.withDefaults()

	if cfg.Interval != 5*time.Minute {
		t.Errorf("default interval = %v, want 5m", cfg.Interval)
	}
	if cfg.BatchSize != 100 {
		t.Errorf("default batch size = %d, want 100", cfg.BatchSize)
	}
	if cfg.QueueSize != 1000 {
		t.Errorf("default queue size = %d, want 1000", cfg.QueueSize)
	}
}
