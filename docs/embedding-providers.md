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
| `PICOCLAW_EMBEDDING_TEXT_TYPE` | Role hint for asymmetric models: `document` (default) or `query` |

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
Google AI Studio (Gemini) embeddings. For production, Vertex AI is recommended (see below).

| Field | Default |
|---|---|
| `model` | `text-embedding-004` |
| `base_url` | `https://generativelanguage.googleapis.com/v1beta` |
| `api_key` | **Required** (Google AI Studio key) |

```json
{ "backend": "google", "api_key": "AIza...", "model": "text-embedding-004" }
```

> **Warning:** AI Studio embeds the API key in the request URL (`?key=...`), which may appear in proxy and access logs. Use Vertex AI for production.

#### Vertex AI

Set `base_url` to your Vertex AI endpoint. Bearer token auth is used automatically when the URL contains `aiplatform.googleapis.com`. Pass an access token (e.g. from `gcloud auth print-access-token`) as `api_key`.

```json
{
  "backend": "google",
  "api_key": "ya29.ACCESS_TOKEN",
  "base_url": "https://us-central1-aiplatform.googleapis.com/v1/projects/MY_PROJECT/locations/us-central1/publishers/google",
  "model": "text-embedding-004"
}
```

The token is sent as `Authorization: Bearer <token>` — no key in the URL.

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

Cache hit/miss/eviction counters are available via `(*CachedProvider).Stats()` and can be exposed to metrics systems.

## Asymmetric Embeddings (`text_type`)

Some providers (currently: **Aliyun DashScope** `text-embedding-v3`) distinguish between corpus documents and search queries at embed time. The `text_type` field controls this:

| Value | Meaning |
|---|---|
| `document` (default) | Text being stored in the vector database |
| `query` | Text used as a search query |

```json
{ "backend": "aliyun", "api_key": "sk-...", "text_type": "query" }
```

Set `text_type: "query"` when you are embedding a query at search time (i.e., in client code that calls `SimilaritySearch`).  
Leave it at the default `"document"` for the store that indexes memories.

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
