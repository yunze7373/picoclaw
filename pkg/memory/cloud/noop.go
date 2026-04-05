package cloud

import "context"

// NoopStore is a no-op implementation of CloudMemoryStore.
// Used when cloud memory is disabled (the default). All methods return
// success immediately with zero allocations.
type NoopStore struct{}

var _ CloudMemoryStore = (*NoopStore)(nil)

func (*NoopStore) UpsertMemory(context.Context, Memory) error { return nil }

func (*NoopStore) UpsertBatch(_ context.Context, memories []Memory) (int, error) {
	return len(memories), nil
}

func (*NoopStore) DeleteMemory(context.Context, string) error { return nil }

func (*NoopStore) SimilaritySearch(context.Context, string, int, float64) ([]SearchResult, error) {
	return nil, nil
}

func (*NoopStore) SyncFromLocal(_ context.Context, memories []Memory) (*SyncStats, error) {
	return &SyncStats{Upserted: len(memories)}, nil
}

func (*NoopStore) HealthCheck(context.Context) (*HealthStatus, error) {
	return &HealthStatus{
		OK:      true,
		Backend: "noop",
	}, nil
}

func (*NoopStore) Close() error { return nil }
