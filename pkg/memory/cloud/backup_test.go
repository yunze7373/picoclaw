package cloud

import (
	"context"
	"sync"
	"testing"
	"time"
)

type mockBackupProvider struct {
	mu       sync.Mutex
	memories []Memory
	calls    int
}

func (p *mockBackupProvider) ExportMemories(_ context.Context, since time.Time) ([]Memory, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	return p.memories, nil
}

func (p *mockBackupProvider) getCalls() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

func TestBackupManager_RunNow(t *testing.T) {
	tracker := &syncTracker{}
	provider := &mockBackupProvider{
		memories: []Memory{
			{ID: "m1", Content: "hello", Kind: "summary"},
			{ID: "m2", Content: "world", Kind: "message"},
		},
	}

	mgr := NewBackupManager(tracker, provider, BackupManagerConfig{
		Interval: time.Hour,
	})

	stats, err := mgr.RunNow(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Upserted != 2 {
		t.Errorf("upserted = %d, want 2", stats.Upserted)
	}

	synced := tracker.getSynced()
	if len(synced) != 2 {
		t.Fatalf("synced = %d, want 2", len(synced))
	}
}

func TestBackupManager_RunNow_NoData(t *testing.T) {
	tracker := &syncTracker{}
	provider := &mockBackupProvider{memories: nil}

	mgr := NewBackupManager(tracker, provider, BackupManagerConfig{})

	stats, err := mgr.RunNow(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Upserted != 0 {
		t.Errorf("upserted = %d, want 0", stats.Upserted)
	}
	if tracker.getCalls() != 0 {
		t.Error("should not call SyncFromLocal with empty data")
	}
}

func TestBackupManager_StartStop(t *testing.T) {
	tracker := &syncTracker{}
	provider := &mockBackupProvider{
		memories: []Memory{{ID: "x"}},
	}

	mgr := NewBackupManager(tracker, provider, BackupManagerConfig{
		Interval: time.Hour,
	})

	ctx := context.Background()
	mgr.Start(ctx)

	// Stop should not panic
	mgr.Stop()
	mgr.Stop() // double stop safe
}

func TestBackupManager_DefaultInterval(t *testing.T) {
	tracker := &syncTracker{}
	provider := &mockBackupProvider{}

	mgr := NewBackupManager(tracker, provider, BackupManagerConfig{})
	if mgr.interval != 6*time.Hour {
		t.Errorf("default interval = %v, want 6h", mgr.interval)
	}
}
