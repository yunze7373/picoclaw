package seahorse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sipeed/picoclaw/pkg/tools"
)

// StatsTool exposes memory statistics to the LLM agent.
type StatsTool struct {
	engine *Engine
}

// NewStatsTool creates a new memory statistics tool.
func NewStatsTool(engine *Engine) *StatsTool {
	return &StatsTool{engine: engine}
}

func (t *StatsTool) Name() string {
	return "memory_stats"
}

func (t *StatsTool) Description() string {
	return `Retrieve memory statistics for the current or all sessions.

Returns message counts, token usage, summary counts, compression ratios,
and time ranges for stored conversations.

Parameters:
- session_key: (optional) Specific session to query. If omitted, returns aggregate stats for all sessions.
- include_sessions: (optional, default false) If true, includes per-session breakdown in the response.

Returns:
{
  "total_sessions": 5,
  "total_messages": 1200,
  "total_tokens": 450000,
  "total_summaries": 85,
  "db_size_bytes": 2097152,
  "sessions": [...]  // only if include_sessions=true
}`
}

func (t *StatsTool) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"session_key": {
				"type": "string",
				"description": "Specific session key to query. Omit for aggregate stats."
			},
			"include_sessions": {
				"type": "boolean",
				"description": "If true, include per-session breakdown.",
				"default": false
			}
		}
	}`)
}

type statsResult struct {
	TotalSessions int              `json:"total_sessions"`
	TotalMessages int              `json:"total_messages"`
	TotalTokens   int              `json:"total_tokens"`
	TotalSummaries int             `json:"total_summaries"`
	DBSizeBytes   int64            `json:"db_size_bytes"`
	Sessions      []sessionStats   `json:"sessions,omitempty"`
}

type sessionStats struct {
	SessionKey  string `json:"session_key"`
	Messages    int    `json:"messages"`
	Tokens      int    `json:"tokens"`
	Summaries   int    `json:"summaries"`
	OldestAt    string `json:"oldest_at,omitempty"`
	NewestAt    string `json:"newest_at,omitempty"`
}

func (t *StatsTool) Execute(ctx context.Context, params map[string]any) tools.ToolResult {
	sessionKey, _ := params["session_key"].(string)
	includeSessions, _ := params["include_sessions"].(bool)

	// Single session query
	if sessionKey != "" {
		ss, err := t.engine.GetSessionStats(ctx, sessionKey)
		if err != nil {
			return tools.ToolResult{IsError: true, ForLLM: fmt.Sprintf("query session: %v", err)}
		}
		if ss == nil {
			return tools.ToolResult{IsError: true, ForLLM: fmt.Sprintf("session %q not found", sessionKey)}
		}

		result := statsResult{
			TotalSessions:  1,
			TotalMessages:  ss.Messages,
			TotalTokens:    ss.Tokens,
			TotalSummaries: ss.Summaries,
			Sessions: []sessionStats{{
				SessionKey: ss.SessionKey,
				Messages:   ss.Messages,
				Tokens:     ss.Tokens,
				Summaries:  ss.Summaries,
				OldestAt:   formatTime(ss.OldestAt),
				NewestAt:   formatTime(ss.NewestAt),
			}},
		}
		return marshalResult(result)
	}

	// All sessions query — delegate to Engine.GetStats
	es, err := t.engine.GetStats(ctx, includeSessions)
	if err != nil {
		return tools.ToolResult{IsError: true, ForLLM: fmt.Sprintf("query all sessions: %v", err)}
	}

	result := statsResult{
		TotalSessions:  es.TotalSessions,
		TotalMessages:  es.TotalMessages,
		TotalTokens:    es.TotalTokens,
		TotalSummaries: es.TotalSummaries,
		DBSizeBytes:    es.DBSizeBytes,
	}

	for _, s := range es.Sessions {
		result.Sessions = append(result.Sessions, sessionStats{
			SessionKey: s.SessionKey,
			Messages:   s.Messages,
			Tokens:     s.Tokens,
			Summaries:  s.Summaries,
			OldestAt:   formatTime(s.OldestAt),
			NewestAt:   formatTime(s.NewestAt),
		})
	}

	return marshalResult(result)
}

func marshalResult(v any) tools.ToolResult {
	data, err := json.Marshal(v)
	if err != nil {
		return tools.ToolResult{IsError: true, ForLLM: fmt.Sprintf("marshal result: %v", err)}
	}
	return tools.ToolResult{ForLLM: string(data)}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
