package cloud

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BackupProvider abstracts the source of memory data for backup.
// Typically implemented by wrapping seahorse.Engine.
type BackupProvider interface {
	// ExportMemories returns all memories that should be backed up to cloud.
	// sinceLastBackup filters to only memories created/updated after the given time.
	// Returns nil slice if nothing to export.
	ExportMemories(ctx context.Context, sinceLastBackup time.Time) ([]Memory, error)
}

// BackupManager periodically exports local memories to the cloud store.
type BackupManager struct {
	store     CloudMemoryStore
	provider  BackupProvider
	interval  time.Duration
	mu        sync.Mutex
	runMu     sync.Mutex
	lastRun   time.Time
	startOnce sync.Once
	stopOnce  sync.Once
	cancel    context.CancelFunc
	done      chan struct{}
	onError   func(err error)
	started   bool
}

// BackupManagerConfig configures the backup manager.
type BackupManagerConfig struct {
	// Interval between backup runs. Default: 6 hours.
	Interval time.Duration

	// OnError is called when a backup fails. Optional.
	OnError func(err error)
}

// NewBackupManager creates a new periodic backup manager.
func NewBackupManager(store CloudMemoryStore, provider BackupProvider, cfg BackupManagerConfig) *BackupManager {
	interval := cfg.Interval
	if interval <= 0 {
		interval = 6 * time.Hour
	}
	return &BackupManager{
		store:    store,
		provider: provider,
		interval: interval,
		done:     make(chan struct{}),
		onError:  cfg.OnError,
	}
}

// Start begins the periodic backup loop.
func (m *BackupManager) Start(ctx context.Context) {
	m.startOnce.Do(func() {
		ctx, m.cancel = context.WithCancel(ctx)
		m.started = true
		go m.loop(ctx)
	})
}

// Stop gracefully shuts down the backup manager.
// Safe to call even if Start was never called.
func (m *BackupManager) Stop() {
	m.stopOnce.Do(func() {
		if !m.started {
			return
		}
		if m.cancel != nil {
			m.cancel()
		}
		<-m.done
	})
}

// RunNow triggers an immediate backup, returning stats.
func (m *BackupManager) RunNow(ctx context.Context) (*SyncStats, error) {
	return m.runBackup(ctx)
}

func (m *BackupManager) loop(ctx context.Context) {
	defer close(m.done)

	// Run initial backup shortly after start
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			if _, err := m.runBackup(ctx); err != nil && m.onError != nil {
				m.onError(fmt.Errorf("periodic backup: %w", err))
			}
			timer.Reset(m.interval)

		case <-ctx.Done():
			return
		}
	}
}

func (m *BackupManager) runBackup(ctx context.Context) (*SyncStats, error) {
	// runMu prevents concurrent backups (TOCTOU race between RunNow and the loop).
	m.runMu.Lock()
	defer m.runMu.Unlock()

	m.mu.Lock()
	since := m.lastRun
	m.mu.Unlock()

	memories, err := m.provider.ExportMemories(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("export memories: %w", err)
	}

	if len(memories) == 0 {
		return &SyncStats{}, nil
	}

	stats, err := m.store.SyncFromLocal(ctx, memories)
	if err != nil {
		return stats, fmt.Errorf("sync to cloud: %w", err)
	}

	m.mu.Lock()
	m.lastRun = time.Now()
	m.mu.Unlock()

	return stats, nil
}
