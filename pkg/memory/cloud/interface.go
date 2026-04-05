// Package cloud provides optional cloud-backed memory storage for PicoClaw.
//
// All implementations are behind a configuration flag and default to disabled.
// When disabled, the NoopStore is used, which silently discards all operations
// with zero overhead.
package cloud

import (
	"context"
	"time"
)

// Memory represents a single memory entry that can be synced to a cloud backend.
type Memory struct {
	// ID is a unique identifier (typically a UUID or hash).
	ID string `json:"id"`

	// SessionKey links back to the local seahorse session.
	SessionKey string `json:"session_key"`

	// Content is the textual memory content (summary, message, etc.).
	Content string `json:"content"`

	// Embedding is the vector representation for similarity search.
	// May be nil if the backend handles embedding internally.
	Embedding []float32 `json:"embedding,omitempty"`

	// Kind classifies the memory: "message", "summary", "condensed".
	Kind string `json:"kind"`

	// TokenCount is the approximate token count of Content.
	TokenCount int `json:"token_count"`

	// CreatedAt is the original creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// SyncedAt records when this memory was last synced to cloud.
	SyncedAt time.Time `json:"synced_at,omitempty"`

	// Metadata holds arbitrary key-value pairs (model, depth, etc.).
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SearchResult wraps a Memory with its similarity score.
type SearchResult struct {
	Memory     Memory  `json:"memory"`
	Similarity float64 `json:"similarity"`
}

// SyncStats contains statistics about a sync operation.
type SyncStats struct {
	Upserted int           `json:"upserted"`
	Deleted  int           `json:"deleted"`
	Errors   int           `json:"errors"`
	Duration time.Duration `json:"duration"`
}

// HealthStatus represents the health of the cloud memory backend.
type HealthStatus struct {
	OK           bool   `json:"ok"`
	Backend      string `json:"backend"`
	Latency      time.Duration `json:"latency"`
	MemoryCount  int    `json:"memory_count"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// CloudMemoryStore defines the interface for cloud-backed memory storage.
// Implementations must be safe for concurrent use.
type CloudMemoryStore interface {
	// UpsertMemory inserts or updates a memory entry in the cloud store.
	// The ID field of m determines whether this is an insert or update.
	UpsertMemory(ctx context.Context, m Memory) error

	// UpsertBatch inserts or updates multiple memories in one operation.
	// Returns the number of successfully upserted entries.
	UpsertBatch(ctx context.Context, memories []Memory) (int, error)

	// DeleteMemory removes a memory by ID.
	DeleteMemory(ctx context.Context, id string) error

	// SimilaritySearch finds memories similar to the query string.
	// topK limits the number of results. minScore filters by minimum similarity.
	SimilaritySearch(ctx context.Context, query string, topK int, minScore float64) ([]SearchResult, error)

	// SyncFromLocal pushes local changes to the cloud store.
	// Returns statistics about the sync operation.
	SyncFromLocal(ctx context.Context, memories []Memory) (*SyncStats, error)

	// HealthCheck verifies the cloud backend is reachable and operational.
	HealthCheck(ctx context.Context) (*HealthStatus, error)

	// Close releases resources held by the store.
	Close() error
}
