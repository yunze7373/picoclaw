# Embedding Providers

PicoClaw supports pluggable vector embedding backends for semantic memory search.
By default, embeddings are **disabled** — all operations have zero overhead and fall back to text-based search.

## Configuration

Add the `embedding` block inside `cloud_memory` in your config:

```json
{
  "cloud_memory": {
    "enabled": true,
    "backend": "supabase",
    "base_url": "https://xxx.supabase.co",
    "api_key": "YOUR_SUPABASE_KEY",
    "embedding": {
      "backend": "openai",
      "api_key": "sk-...",
      "model": "text-embedding-3-small",
      "cache_size": 10000
    }
  }
}
```

Or via environment variables:

| Variable | Description |
|---|---|
| `PICOCLAW_EMBEDDING_BACKEND` | Backend name (see below) |
| `PICOCLAW_EMBEDDING_MODEL` | Model override |
| `PICOCLAW_EMBEDDING_API_KEY` | API key for cloud providers |
| `PICOCLAW_EMBEDDING_BASE_URL` | Endpoint override |
| `PICOCLAW_EMBEDDING_CACHE_SIZE` | LRU cache entries (default: 10000) |

## Backends

### `none` (default)
No embeddings. Zero network calls. Supabase falls back to text-based `match_memories` RPC.

### `openai`
OpenAI text-embedding API. Also compatible with **Azure OpenAI** (set `base_url`).

| Field | Default |
|---|---|
| `model` | `text-embedding-3-small` |
| `base_url` | `https://api.openai.com/v1` |
| `api_key` | **Required** |

```json
{ "backend": "openai", "api_key": "sk-...", "model": "text-embedding-3-large" }
```

### `ollama`
Local [Ollama](https://ollama.com/) instance. No API key required. Ideal for **offline** and **Termux/Android** deployments.

| Field | Default |
|---|---|
| `model` | `nomic-embed-text` |
| `base_url` | `http://localhost:11434` |

```json
{ "backend": "ollama", "model": "nomic-embed-text" }
```

Pull the model first: `ollama pull nomic-embed-text`

### `google`
Google AI Studio (Gemini) embeddings. Set `base_url` for Vertex AI.

| Field | Default |
|---|---|
| `model` | `text-embedding-004` |
| `base_url` | `https://generativelanguage.googleapis.com/v1beta` |
| `api_key` | **Required** (Google AI Studio key) |

```json
{ "backend": "google", "api_key": "AIza...", "model": "text-embedding-004" }
```

### `aliyun`
Alibaba Cloud DashScope (Bailian) text-embedding API.

| Field | Default |
|---|---|
| `model` | `text-embedding-v3` |
| `base_url` | `https://dashscope.aliyuncs.com/api/v1` |
| `api_key` | **Required** |

```json
{ "backend": "aliyun", "api_key": "sk-...", "model": "text-embedding-v3" }
```

### `deepseek`
DeepSeek embeddings (OpenAI-compatible API).

| Field | Default |
|---|---|
| `model` | `deepseek-embedding` |
| `base_url` | `https://api.deepseek.com/v1` |
| `api_key` | **Required** |

```json
{ "backend": "deepseek", "api_key": "sk-...", "model": "deepseek-embedding" }
```

## Caching

All providers are automatically wrapped with an **LRU in-memory cache** when `cache_size > 0` (the default).  
Identical texts are embedded only once per process lifetime. Set `cache_size: 0` to disable.

## Supabase Vector Search

When an embedding provider is configured, `SimilaritySearch` calls `match_memories_vector` instead of `match_memories`. Create this function in Supabase:

```sql
CREATE OR REPLACE FUNCTION match_memories_vector(
  query_embedding vector,
  match_count     INT,
  min_similarity  FLOAT
)
RETURNS TABLE (id TEXT, content TEXT, session_key TEXT, kind TEXT, similarity FLOAT)
LANGUAGE plpgsql AS $$
BEGIN
  RETURN QUERY
  SELECT m.id, m.content, m.session_key, m.kind,
         1 - (m.embedding <=> query_embedding) AS similarity
  FROM memories m
  WHERE m.embedding IS NOT NULL
    AND 1 - (m.embedding <=> query_embedding) >= min_similarity
  ORDER BY m.embedding <=> query_embedding
  LIMIT match_count;
END;
$$;
```

The text-based `match_memories` fallback (used when no embedding provider is set):

```sql
CREATE OR REPLACE FUNCTION match_memories(
  query_text      TEXT,
  match_count     INT,
  min_similarity  FLOAT
)
RETURNS TABLE (id TEXT, content TEXT, session_key TEXT, kind TEXT, similarity FLOAT)
LANGUAGE plpgsql AS $$
BEGIN
  RETURN QUERY
  SELECT m.id, m.content, m.session_key, m.kind,
         ts_rank(to_tsvector('simple', m.content), plainto_tsquery('simple', query_text)) AS similarity
  FROM memories m
  WHERE to_tsvector('simple', m.content) @@ plainto_tsquery('simple', query_text)
  ORDER BY similarity DESC
  LIMIT match_count;
END;
$$;
```
