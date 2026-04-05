# Phase 5 Code Review — AUDIT_REPORT_005

**Verdict: NEEDS_WORK**  
**Reviewer:** staff-code-reviewer (automated Phase R)  
**Scope:** Phase 5 changes — CacheStats, Vertex AI bearer auth, TextType config, OnEmbedError hook

---

## P0 Findings (Blocking)

### P0-1 — Google AI Studio API key leaks into log output on network errors
`pkg/memory/embedding/google.go` — `p.client.Do(req)` failures return a `*url.Error` whose `Error()` string includes the full request URL including `?key=AIza...`. This is wrapped and propagated to `OnEmbedError`, which logs it via `logger.WarnCF`. API key appears in production logs on every network hiccup.

**Fix:** Catch `*url.Error`, extract `Op` and `Err` without the URL.

### P0-2 — Wrong comment causes Vertex AI URL double-construction
`pkg/memory/embedding/google.go:94–97` — Comment says baseURL "already contains model path", but code appends `/models/%s:batchEmbedContents`. Developer following the comment would pass a full endpoint URL and get guaranteed 404.

**Fix:** Correct the comment to describe what `baseURL` should be (publisher root, not full path).

---

## P1 Findings (Significant)

### P1-1 — `onEmbedError` not called in `SimilaritySearch`
`pkg/memory/cloud/supabase.go` — Embed failure in `SimilaritySearch` silently falls back to text search without invoking `onEmbedError`. Broken provider degrades silently with no alerts.

### P1-2 — `CachedProvider` returns direct references to cached vectors
`pkg/memory/embedding/cache.go` — Cache hits return `el.Value.(*cacheEntry).vec` directly. Any caller that modifies the returned slice corrupts the cache entry. Cache poisoning bug.

**Fix:** `copy()` on hit return.

### P1-3 — `TextType` silently ignored for Google provider
`pkg/memory/embedding/google.go` — `Config.TextType` is accepted but never used. Google's `text-embedding-004` supports `taskType`. Silent quality degradation with no error.

**Fix:** Add `taskType` field to `googleEmbedContentRequest`, map `TextType` values to Google enum values.

### P1-4 — Vertex AI tokens expire silently; no operational guidance
Vertex AI `apiKey` is a short-lived OAuth2 token (~1h TTL). Expiry causes all embeds to fail silently (logged as WARN). No documentation about token rotation.

**Fix:** Add warning comment in code + docs section on token management.

### P1-5 — `Stats()` snapshot inconsistency
`pkg/memory/embedding/cache.go` — `Size` read under lock, then lock released before reading atomic counters. Snapshot can show mismatched state.

**Fix:** Read all counters under the same lock.

### P1-6 — `SyncManager` started with `context.Background()`
`pkg/agent/cloud_memory.go` — `stack.sync.Start(context.Background())` prevents context-based shutdown. Goroutine leaks if `Close()` is not reached.

**Fix:** Thread agent lifecycle context into `initCloudMemory` and `Start()`.

---

## P2 Findings (Minor)

- `TestCachedProvider_LRUEviction` makes no assertions
- No concurrent stress test for `CachedProvider` despite "Thread-safe" doc claim
- `time.Now()` side effect inside `memoryToRow` (not testable without clock mock)
- `hasEmbedder()` uses fragile string matching (`Model() == "none"`) — prefer type assertion
- Memory IDs use millisecond timestamps (collision risk under high load)
- `CacheSize = -1` behavior differs between `New()` and `NewCachedProvider()`
- Floating godoc comment on `CachedProvider` not attached to type
- `NoopProvider.Embed` returns non-nil empty slice for empty input (inconsistent with other providers)
- HTTP response body not size-limited in Google provider success path

---

## Fix Plan

Commits to produce:
1. `fix(phase-r5): P0 — redact API key from Google error messages`
2. `fix(phase-r5): P0 — fix Vertex AI URL comment`
3. `fix(phase-r5): P1 — call onEmbedError in SimilaritySearch`
4. `fix(phase-r5): P1 — clone cached vectors on hit return`
5. `fix(phase-r5): P1 — implement task_type for Google provider`
6. `fix(phase-r5): P1 — Stats() consistent snapshot under lock`
7. `fix(phase-r5): P1/P2 — misc fixes (NoopProvider, comment, hasEmbedder)`
