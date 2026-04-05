package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
	cloud "github.com/sipeed/picoclaw/pkg/memory/cloud"
)

// cloudMemoryStack holds all cloud memory components for lifecycle management.
type cloudMemoryStack struct {
	store   cloud.CloudMemoryStore
	sync    *cloud.SyncManager
	backup  *cloud.BackupManager
	subID   uint64
	eventBus *EventBus
}

// initCloudMemory creates and wires the cloud memory subsystem based on config.
// Returns nil if cloud memory is disabled (default). The returned stack must
// be closed via stack.Close() when the agent shuts down.
func initCloudMemory(cfg config.CloudMemoryConfig, bus *EventBus) (*cloudMemoryStack, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	// Create cloud store based on backend
	var store cloud.CloudMemoryStore
	switch cfg.Backend {
	case "supabase":
		if cfg.BaseURL == "" || cfg.APIKey == "" {
			return nil, fmt.Errorf("cloud memory: supabase backend requires base_url and api_key")
		}
		tableName := cfg.TableName
		if tableName == "" {
			tableName = "memories"
		}
		var err error
		store, err = cloud.NewSupabaseStore(cloud.SupabaseConfig{
			BaseURL:   cfg.BaseURL,
			APIKey:    cfg.APIKey,
			TableName: tableName,
		})
		if err != nil {
			return nil, fmt.Errorf("cloud memory: create supabase store: %w", err)
		}
	case "", "none":
		store = cloud.NewNoopStore()
	default:
		return nil, fmt.Errorf("cloud memory: unknown backend %q", cfg.Backend)
	}

	stack := &cloudMemoryStack{
		store:    store,
		eventBus: bus,
	}

	// Create SyncManager for real-time event-driven sync
	syncInterval := time.Duration(cfg.SyncIntervalSeconds) * time.Second
	if syncInterval <= 0 {
		syncInterval = 5 * time.Minute
	}
	stack.sync = cloud.NewSyncManager(store, cloud.SyncManagerConfig{
		Interval: syncInterval,
		OnError: func(err error, batchSize int) {
			logger.WarnCF("cloud_memory", "sync flush failed", map[string]any{
				"error":      err.Error(),
				"batch_size": batchSize,
			})
		},
	})

	// Subscribe to session_summarize events and feed summaries to SyncManager
	if bus != nil {
		sub := bus.Subscribe(0)
		stack.subID = sub.ID
		go stack.eventLoop(sub)
	}

	// Start SyncManager
	stack.sync.Start(context.Background())

	logger.InfoCF("cloud_memory", "initialized", map[string]any{
		"backend":       cfg.Backend,
		"sync_interval": syncInterval.String(),
	})

	return stack, nil
}

// eventLoop listens for session_summarize events and enqueues memories for sync.
func (s *cloudMemoryStack) eventLoop(sub EventSubscription) {
	for evt := range sub.C {
		if evt.Kind != EventKindSessionSummarize {
			continue
		}

		payload, ok := evt.Payload.(SessionSummarizePayload)
		if !ok {
			continue
		}

		sessionKey := evt.Meta.SessionKey
		if sessionKey == "" {
			continue
		}

		mem := cloud.Memory{
			ID:         fmt.Sprintf("summary-%s-%d", sessionKey, time.Now().UnixMilli()),
			SessionKey: sessionKey,
			Content:    fmt.Sprintf("Session summarization: %d messages summarized, %d kept, summary length %d", payload.SummarizedMessages, payload.KeptMessages, payload.SummaryLen),
			Kind:       "summary",
		}
		s.sync.Enqueue(mem)
	}
}

// Close gracefully shuts down all cloud memory components.
func (s *cloudMemoryStack) Close() {
	if s == nil {
		return
	}
	// Unsubscribe from events first
	if s.eventBus != nil && s.subID > 0 {
		s.eventBus.Unsubscribe(s.subID)
	}
	// Stop SyncManager (flushes pending)
	if s.sync != nil {
		s.sync.Stop()
	}
	// Stop BackupManager if running
	if s.backup != nil {
		s.backup.Stop()
	}
	// Close cloud store
	if s.store != nil {
		s.store.Close()
	}
}
