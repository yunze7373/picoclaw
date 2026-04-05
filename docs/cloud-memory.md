# Cloud Memory Module

> Optional cloud-backed memory for PicoClaw with vector similarity search.

## Overview

PicoClaw stores conversation memory locally in SQLite (via the Seahorse engine). The Cloud Memory module adds **optional** sync to a cloud backend for:

- **Cross-device memory**: Access conversation history from any device
- **Similarity search**: Find relevant past conversations using vector embeddings (pgvector)
- **Backup**: Cloud-side persistence as a safety net for local data

**Default: disabled.** When off, a zero-overhead NoopStore is used.

---

## Architecture

```
┌─────────────────┐     ┌──────────────┐     ┌───────────────────┐
│  Agent Loop      │────▶│  SyncManager │────▶│  CloudMemoryStore │
│  (EventBus)      │     │  (batched)   │     │  (Supabase/Noop)  │
└─────────────────┘     └──────────────┘     └───────────────────┘
         │                                             │
         ▼                                             ▼
┌─────────────────┐                          ┌───────────────────┐
│  Seahorse SQLite │                          │  Supabase+pgvector│
│  (local, primary)│                          │  (cloud, optional)│
└─────────────────┘                          └───────────────────┘
```

- **Local-first**: Seahorse SQLite remains the primary store
- **Async sync**: SyncManager batches and debounces uploads
- **Graceful degradation**: Network failures don't affect local operation

---

## Configuration

Add to your PicoClaw config file:

```json
{
  "cloud_memory": {
    "enabled": true,
    "backend": "supabase",
    "base_url": "https://your-project.supabase.co",
    "api_key": "your-anon-or-service-role-key",
    "table_name": "memories",
    "sync_interval_seconds": 300,
    "max_memories": 10000
  }
}
```

Or via environment variables:

```bash
export PICOCLAW_CLOUD_MEMORY_ENABLED=true
export PICOCLAW_CLOUD_MEMORY_BACKEND=supabase
export PICOCLAW_CLOUD_MEMORY_BASE_URL=https://your-project.supabase.co
export PICOCLAW_CLOUD_MEMORY_API_KEY=your-key
```

### Configuration Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Enable cloud memory sync |
| `backend` | string | `"none"` | Backend: `"supabase"` or `"none"` |
| `base_url` | string | - | Supabase project URL |
| `api_key` | string | - | Supabase API key (anon or service role) |
| `table_name` | string | `"memories"` | Database table name |
| `sync_interval_seconds` | int | `300` | Seconds between auto-sync (0 = manual only) |
| `max_memories` | int | `10000` | Max memories per session (0 = unlimited) |

---

## Supabase Setup

### 1. Create a Supabase Project

Go to [supabase.com](https://supabase.com) and create a new project.

### 2. Enable pgvector Extension

In the Supabase SQL Editor:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

### 3. Create the Memories Table

```sql
CREATE TABLE memories (
    id TEXT PRIMARY KEY,
    session_key TEXT NOT NULL,
    content TEXT NOT NULL,
    embedding vector(384),  -- adjust dimension to match your embedding model
    kind TEXT NOT NULL DEFAULT 'message',
    token_count INTEGER NOT NULL DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for fast vector similarity search
CREATE INDEX ON memories USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Index for session-based queries
CREATE INDEX idx_memories_session ON memories(session_key);
CREATE INDEX idx_memories_kind ON memories(kind);
```

### 4. Create the Similarity Search Function

```sql
CREATE OR REPLACE FUNCTION match_memories(
    query_text TEXT,
    match_count INT DEFAULT 10,
    min_similarity FLOAT DEFAULT 0.5
)
RETURNS TABLE (
    id TEXT,
    content TEXT,
    session_key TEXT,
    kind TEXT,
    similarity FLOAT
)
LANGUAGE plpgsql
AS $$
BEGIN
    -- Note: This is a text-based fallback. For vector similarity,
    -- replace with embedding-based search using your chosen model.
    RETURN QUERY
    SELECT
        m.id,
        m.content,
        m.session_key,
        m.kind,
        ts_rank(to_tsvector('english', m.content), plainto_tsquery('english', query_text))::FLOAT AS similarity
    FROM memories m
    WHERE to_tsvector('english', m.content) @@ plainto_tsquery('english', query_text)
    ORDER BY similarity DESC
    LIMIT match_count;
END;
$$;
```

### 5. Configure Row Level Security (Optional)

```sql
ALTER TABLE memories ENABLE ROW LEVEL SECURITY;

-- Allow the service role full access
CREATE POLICY "Service role access" ON memories
    FOR ALL
    USING (auth.role() = 'service_role');

-- Or allow anon key read-only access
CREATE POLICY "Anon read access" ON memories
    FOR SELECT
    USING (true);
```

---

## Memory Analysis Tools

Two optional tools for inspecting memory state:

### `memory_stats`

Returns memory statistics (message counts, token usage, summaries).

Enable in config:
```json
{
  "tools": {
    "memory_stats": { "enabled": true }
  }
}
```

Example output:
```json
{
  "total_sessions": 5,
  "total_messages": 1200,
  "total_tokens": 450000,
  "total_summaries": 85,
  "db_size_bytes": 2097152
}
```

### `memory_health`

Checks SQLite database health (integrity, WAL status, FTS5 indexes).

Enable in config:
```json
{
  "tools": {
    "memory_health": { "enabled": true }
  }
}
```

Example output:
```json
{
  "ok": true,
  "integrity": "ok",
  "journal_mode": "wal",
  "wal_pages": 42,
  "db_size_bytes": 2097152,
  "fts5_ok": true,
  "latency_ms": 12,
  "issues": []
}
```

---

## Sync Behavior

- **Timer-based**: Auto-flushes every `sync_interval_seconds`
- **Batch-based**: Flushes when batch reaches 100 memories
- **Shutdown flush**: Pending memories are synced before shutdown (10s deadline)
- **Non-blocking**: Queue overflow silently drops memories (agent loop never blocks)
- **Idempotent**: Upsert semantics — safe to sync the same memory multiple times

---

## Privacy Considerations

- **API keys**: Store in environment variables, never in config files committed to git
- **Content**: All conversation content is sent to the cloud backend when sync is enabled
- **Network**: Uses HTTPS for all Supabase communication
- **Disable**: Set `enabled: false` to completely disable cloud sync (zero network calls)

---

## Troubleshooting

### Sync not working

1. Check config: `"cloud_memory": { "enabled": true }`
2. Verify Supabase URL and API key
3. Use `memory_health` tool to check local DB status
4. Check Supabase dashboard for table existence and RLS policies

### High memory usage

- Reduce `max_memories` to limit stored entries
- Increase `sync_interval_seconds` to reduce sync frequency
- Check `memory_stats` for total token counts

### Connection errors

- The SyncManager gracefully handles network failures
- Local operations are never affected by cloud sync errors
- Check Supabase project status and quotas
