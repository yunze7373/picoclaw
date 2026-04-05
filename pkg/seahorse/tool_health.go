package seahorse

import (
	"context"

	"github.com/sipeed/picoclaw/pkg/tools"
)

// HealthTool checks the health of the seahorse memory database.
type HealthTool struct {
	engine *Engine
}

// NewHealthTool creates a new memory health check tool.
func NewHealthTool(engine *Engine) *HealthTool {
	return &HealthTool{engine: engine}
}

func (t *HealthTool) Name() string {
	return "memory_health"
}

func (t *HealthTool) Description() string {
	return `Check the health of the memory database.

Runs diagnostics including SQLite integrity check, WAL status, FTS5 index
health, database file size, and connection status.

Parameters: none required.

Returns:
{
  "ok": true,
  "integrity": "ok",
  "journal_mode": "wal",
  "wal_pages": 42,
  "db_size_bytes": 2097152,
  "fts5_ok": true,
  "latency_ms": 12,
  "issues": []
}`
}

func (t *HealthTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

type healthResult struct {
	OK          bool     `json:"ok"`
	Integrity   string   `json:"integrity"`
	JournalMode string   `json:"journal_mode"`
	WALPages    int      `json:"wal_pages"`
	DBSizeBytes int64    `json:"db_size_bytes"`
	FTS5OK      bool     `json:"fts5_ok"`
	LatencyMs   int64    `json:"latency_ms"`
	Issues      []string `json:"issues"`
}

func (t *HealthTool) Execute(ctx context.Context, _ map[string]any) *tools.ToolResult {
	h, err := t.engine.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult("health check failed: " + err.Error())
	}

	result := healthResult{
		OK:          h.OK,
		Integrity:   h.Integrity,
		JournalMode: h.JournalMode,
		WALPages:    h.WALPages,
		DBSizeBytes: h.DBSizeBytes,
		FTS5OK:      h.FTS5OK,
		LatencyMs:   h.LatencyMs,
		Issues:      h.Issues,
	}
	if result.Issues == nil {
		result.Issues = []string{}
	}

	return marshalResult(result)
}
