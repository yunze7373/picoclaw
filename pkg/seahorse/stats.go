package seahorse

import (
	"context"
	"os"
	"time"
)

// EngineStats contains aggregate memory statistics.
type EngineStats struct {
	TotalSessions  int            `json:"total_sessions"`
	TotalMessages  int            `json:"total_messages"`
	TotalTokens    int            `json:"total_tokens"`
	TotalSummaries int            `json:"total_summaries"`
	DBSizeBytes    int64          `json:"db_size_bytes"`
	Sessions       []SessionStats `json:"sessions,omitempty"`
}

// SessionStats contains per-session statistics.
type SessionStats struct {
	SessionKey string    `json:"session_key"`
	Messages   int       `json:"messages"`
	Tokens     int       `json:"tokens"`
	Summaries  int       `json:"summaries"`
	OldestAt   time.Time `json:"oldest_at,omitempty"`
	NewestAt   time.Time `json:"newest_at,omitempty"`
}

// EngineHealth contains database health check results.
type EngineHealth struct {
	OK          bool     `json:"ok"`
	Integrity   string   `json:"integrity"`
	JournalMode string   `json:"journal_mode"`
	WALPages    int      `json:"wal_pages"`
	DBSizeBytes int64    `json:"db_size_bytes"`
	FTS5OK      bool     `json:"fts5_ok"`
	LatencyMs   int64    `json:"latency_ms"`
	Issues      []string `json:"issues"`
}

// GetStats returns aggregate memory statistics for all sessions.
// If includeSessions is true, per-session breakdowns are included.
func (e *Engine) GetStats(ctx context.Context, includeSessions bool) (*EngineStats, error) {
	statuses, err := e.store.GetAllSessionStatuses(ctx)
	if err != nil {
		return nil, err
	}

	stats := &EngineStats{
		TotalSessions: len(statuses),
	}

	for _, s := range statuses {
		stats.TotalMessages += s.Messages
		stats.TotalTokens += s.TotalTokens
		stats.TotalSummaries += s.Summaries

		if includeSessions {
			stats.Sessions = append(stats.Sessions, SessionStats{
				SessionKey: s.SessionKey,
				Messages:   s.Messages,
				Tokens:     s.TotalTokens,
				Summaries:  s.Summaries,
				OldestAt:   s.OldestAt,
				NewestAt:   s.NewestAt,
			})
		}
	}

	stats.DBSizeBytes = e.getDBSize()
	return stats, nil
}

// GetSessionStats returns statistics for a specific session.
func (e *Engine) GetSessionStats(ctx context.Context, sessionKey string) (*SessionStats, error) {
	status, err := e.store.GetSessionStatus(ctx, sessionKey)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, nil
	}

	return &SessionStats{
		SessionKey: status.SessionKey,
		Messages:   status.Messages,
		Tokens:     status.TotalTokens,
		Summaries:  status.Summaries,
		OldestAt:   status.OldestAt,
		NewestAt:   status.NewestAt,
	}, nil
}

// GetHealth runs diagnostic checks on the memory database.
func (e *Engine) GetHealth(ctx context.Context) (*EngineHealth, error) {
	start := time.Now()
	result := &EngineHealth{OK: true}

	db := e.store.db
	if db == nil {
		return &EngineHealth{OK: false, Issues: []string{"database not initialized"}}, nil
	}

	// 1. Integrity check
	var integrity string
	if err := db.QueryRowContext(ctx, "PRAGMA quick_check").Scan(&integrity); err != nil {
		result.Issues = append(result.Issues, "integrity check failed: "+err.Error())
		result.OK = false
		integrity = "error"
	}
	result.Integrity = integrity
	if integrity != "ok" {
		result.OK = false
	}

	// 2. Journal mode
	var journalMode string
	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode); err == nil {
		result.JournalMode = journalMode
		if journalMode != "wal" {
			result.Issues = append(result.Issues, "expected WAL mode, got: "+journalMode)
		}
	}

	// 3. WAL file size (read-only)
	walSize := e.getWALSize()
	result.WALPages = int(walSize / 4096)

	// 4. DB file size
	result.DBSizeBytes = e.getDBSize()

	// 5. FTS5 presence check (read-only)
	fts5OK := true
	var ftsCount int
	if err := db.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name IN ('summaries_fts','messages_fts')",
	).Scan(&ftsCount); err != nil {
		fts5OK = false
		result.Issues = append(result.Issues, "FTS5 check failed: "+err.Error())
	} else if ftsCount != 2 {
		fts5OK = false
		result.Issues = append(result.Issues, "FTS5 tables missing from schema")
	}
	result.FTS5OK = fts5OK
	if !fts5OK {
		result.OK = false
	}

	if walSize > 50*1024*1024 {
		result.Issues = append(result.Issues, "WAL file is large, consider checkpoint")
	}

	result.LatencyMs = time.Since(start).Milliseconds()
	return result, nil
}

func (e *Engine) getDBSize() int64 {
	return getFileSize(e.config.DBPath)
}

func (e *Engine) getWALSize() int64 {
	return getFileSize(e.config.DBPath + "-wal")
}

func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
