package cloud

import (
	"context"
	"sync"
	"time"
)

// SyncManager coordinates asynchronous cloud memory sync operations.
// It batches memories and debounces syncs to avoid excessive network calls.
//
// Usage:
//
//	mgr := NewSyncManager(store, SyncManagerConfig{Interval: 5 * time.Minute})
//	mgr.Start(ctx)
//	mgr.Enqueue(memory1, memory2) // non-blocking
//	mgr.Stop()                     // flushes pending and stops
type SyncManager struct {
	store    CloudMemoryStore
	config    SyncManagerConfig
	queue     chan Memory
	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	cancel    context.CancelFunc
}

// SyncManagerConfig configures the SyncManager behavior.
type SyncManagerConfig struct {
	// Interval between automatic sync flushes. Default: 5 minutes.
	Interval time.Duration

	// BatchSize is the max number of memories per sync batch. Default: 100.
	BatchSize int

	// QueueSize is the buffer size for the enqueue channel. Default: 1000.
	QueueSize int

	// OnError is called when a sync flush fails. Optional.
	OnError func(err error, batchSize int)
}

func (c *SyncManagerConfig) withDefaults() SyncManagerConfig {
	out := *c
	if out.Interval <= 0 {
		out.Interval = 5 * time.Minute
	}
	if out.BatchSize <= 0 {
		out.BatchSize = 100
	}
	if out.QueueSize <= 0 {
		out.QueueSize = 1000
	}
	return out
}

// NewSyncManager creates a new SyncManager. Call Start to begin background syncing.
func NewSyncManager(store CloudMemoryStore, cfg SyncManagerConfig) *SyncManager {
	cfg = cfg.withDefaults()
	return &SyncManager{
		store:  store,
		config: cfg,
		queue:  make(chan Memory, cfg.QueueSize),
		done:   make(chan struct{}),
	}
}

// Start begins the background sync loop. Call Stop to shut down gracefully.
// Safe to call multiple times; only the first call starts the loop.
func (m *SyncManager) Start(ctx context.Context) {
	m.startOnce.Do(func() {
		ctx, m.cancel = context.WithCancel(ctx)
		go m.loop(ctx)
	})
}

// Enqueue adds memories to the sync queue without blocking.
// If the queue is full or the manager is stopped, memories are silently dropped.
func (m *SyncManager) Enqueue(memories ...Memory) {
	select {
	case <-m.done:
		return // manager stopped
	default:
	}
	for _, mem := range memories {
		select {
		case m.queue <- mem:
		default:
			// Queue full — drop silently to avoid blocking the agent loop.
		}
	}
}

// Stop gracefully shuts down the sync manager, flushing any pending memories.
func (m *SyncManager) Stop() {
	m.stopOnce.Do(func() {
		if m.cancel != nil {
			m.cancel()
		}
		<-m.done
	})
}

func (m *SyncManager) loop(ctx context.Context) {
	defer close(m.done)

	ticker := time.NewTicker(m.config.Interval)
	defer ticker.Stop()

	batch := make([]Memory, 0, m.config.BatchSize)

	for {
		select {
		case mem := <-m.queue:
			batch = append(batch, mem)
			if len(batch) >= m.config.BatchSize {
				m.flush(ctx, batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				m.flush(ctx, batch)
				batch = batch[:0]
			}

		case <-ctx.Done():
			// Drain remaining queue items
			for {
				select {
				case mem := <-m.queue:
					batch = append(batch, mem)
				default:
					goto done
				}
			}
		done:
			if len(batch) > 0 {
				// Use a short deadline for final flush
				flushCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				m.flush(flushCtx, batch)
				cancel()
			}
			return
		}
	}
}

func (m *SyncManager) flush(ctx context.Context, batch []Memory) {
	if len(batch) == 0 {
		return
	}

	// Copy batch to avoid data races
	toSync := make([]Memory, len(batch))
	copy(toSync, batch)

	_, err := m.store.SyncFromLocal(ctx, toSync)
	if err != nil && m.config.OnError != nil {
		m.config.OnError(err, len(toSync))
	}
}
