package seahorse

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sipeed/picoclaw/pkg/tools"
)

// HealthTool checks the health of the seahorse memory database.
type HealthTool struct {
	engine *RetrievalEngine
}

// NewHealthTool creates a new memory health check tool.
func NewHealthTool(engine *RetrievalEngine) *HealthTool {
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

func (t *HealthTool) Parameters() json.RawMessage {
	return json.RawMessage(`{"type": "object", "properties": {}}`)
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

func (t *HealthTool) Execute(ctx context.Context, _ map[string]any) tools.ToolResult {
	store := t.engine.Store()
	if store == nil {
		return tools.ToolResult{IsError: true, ForLLM: "memory store not available"}
	}
	if store.db == nil {
		return tools.ToolResult{IsError: true, ForLLM: "database not initialized"}
	}

	start := time.Now()
	result := healthResult{OK: true}

	// 1. SQLite integrity check (quick mode)
	var integrity string
	if err := store.db.QueryRowContext(ctx, "PRAGMA quick_check").Scan(&integrity); err != nil {
		result.Issues = append(result.Issues, fmt.Sprintf("integrity check failed: %v", err))
		result.OK = false
		integrity = "error"
	}
	result.Integrity = integrity
	if integrity != "ok" {
		result.OK = false
		result.Issues = append(result.Issues, "database integrity check returned: "+integrity)
	}

	// 2. Journal mode
	var journalMode string
	if err := store.db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode); err != nil {
		result.Issues = append(result.Issues, fmt.Sprintf("journal mode check failed: %v", err))
	} else {
		result.JournalMode = journalMode
		if journalMode != "wal" {
			result.Issues = append(result.Issues, "expected WAL journal mode, got: "+journalMode)
		}
	}

	// 3. WAL file size (read-only, no checkpoint side effects)
	walSize := getWALFileSize(store)
	result.WALPages = int(walSize / 4096) // approximate page count

	// 4. Database file size
	result.DBSizeBytes = getDBFileSize(store)

	// 5. FTS5 health check
	fts5OK := true
	if _, err := store.db.ExecContext(ctx, "INSERT INTO summaries_fts(summaries_fts) VALUES('integrity-check')"); err != nil {
		fts5OK = false
		result.Issues = append(result.Issues, fmt.Sprintf("FTS5 summaries index issue: %v", err))
	}
	if _, err := store.db.ExecContext(ctx, "INSERT INTO messages_fts(messages_fts) VALUES('integrity-check')"); err != nil {
		fts5OK = false
		result.Issues = append(result.Issues, fmt.Sprintf("FTS5 messages index issue: %v", err))
	}
	result.FTS5OK = fts5OK
	if !fts5OK {
		result.OK = false
	}

	// 6. Check WAL file size (warn if large)
	walSize := getWALFileSize(store)
	if walSize > 50*1024*1024 { // 50MB
		result.Issues = append(result.Issues, fmt.Sprintf("WAL file is large: %d bytes, consider checkpoint", walSize))
	}

	result.LatencyMs = time.Since(start).Milliseconds()

	return marshalResult(result)
}

func getWALFileSize(store *Store) int64 {
	if store == nil || store.db == nil {
		return 0
	}
	var dbPath string
	row := store.db.QueryRow("PRAGMA database_list")
	var seq int
	var name string
	if err := row.Scan(&seq, &name, &dbPath); err != nil {
		return 0
	}
	walPath := dbPath + "-wal"
	info, err := os.Stat(walPath)
	if err != nil {
		return 0
	}
	return info.Size()
}
