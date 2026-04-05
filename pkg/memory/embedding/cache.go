package embedding

import (
"container/list"
"context"
"crypto/sha256"
"encoding/hex"
"sync"
)

const defaultCacheSize = 10000

// CachedProvider wraps a Provider with an in-memory LRU cache.
// Identical texts (same hash) are embedded only once per cache lifetime.
// Thread-safe for concurrent use.
var _ Provider = (*CachedProvider)(nil)

type cacheEntry struct {
key string
vec []float32
}

// CachedProvider wraps an inner Provider with an O(1) LRU cache.
type CachedProvider struct {
inner   Provider
mu      sync.Mutex
cache   map[string]*list.Element // key -> doubly-linked-list node
order   *list.List               // front=LRU (oldest), back=MRU (newest)
maxSize int
}

// NewCachedProvider wraps p with an LRU cache of maxSize entries.
// If maxSize <= 0, defaultCacheSize (10000) is used.
func NewCachedProvider(p Provider, maxSize int) *CachedProvider {
if maxSize <= 0 {
maxSize = defaultCacheSize
}
return &CachedProvider{
inner:   p,
cache:   make(map[string]*list.Element, maxSize),
order:   list.New(),
maxSize: maxSize,
}
}

// Model delegates to the inner provider.
func (c *CachedProvider) Model() string { return c.inner.Model() }

// Dims delegates to the inner provider.
func (c *CachedProvider) Dims() int { return c.inner.Dims() }

// Embed returns cached embeddings where available and calls the inner
// provider only for cache misses. The mutex is released before the network
// call to avoid blocking concurrent callers during latency.
func (c *CachedProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
if len(texts) == 0 {
return nil, nil
}

keys := make([]string, len(texts))
for i, t := range texts {
keys[i] = hashText(t)
}

c.mu.Lock()
result := make([][]float32, len(texts))
missIdxs := make([]int, 0)
missTexts := make([]string, 0)

for i, k := range keys {
if el, ok := c.cache[k]; ok {
result[i] = el.Value.(*cacheEntry).vec
c.order.MoveToBack(el) // O(1) touch
} else {
missIdxs = append(missIdxs, i)
missTexts = append(missTexts, texts[i])
}
}
c.mu.Unlock()

if len(missIdxs) == 0 {
return result, nil
}

vectors, err := c.inner.Embed(ctx, missTexts)
if err != nil {
return nil, err
}

c.mu.Lock()
for j, idx := range missIdxs {
v := vectors[j]
result[idx] = v
c.store(keys[idx], v)
}
c.mu.Unlock()

return result, nil
}

// store inserts or updates a cache entry in O(1). Must be called with c.mu held.
func (c *CachedProvider) store(key string, v []float32) {
if el, ok := c.cache[key]; ok {
el.Value.(*cacheEntry).vec = v
c.order.MoveToBack(el)
return
}
if c.order.Len() >= c.maxSize {
oldest := c.order.Front()
if oldest != nil {
c.order.Remove(oldest)
delete(c.cache, oldest.Value.(*cacheEntry).key)
}
}
el := c.order.PushBack(&cacheEntry{key: key, vec: v})
c.cache[key] = el
}

// hashText returns a 128-bit truncation of SHA-256 as a hex string.
// 128 bits provides ample collision resistance for a bounded cache.
func hashText(s string) string {
h := sha256.Sum256([]byte(s))
return hex.EncodeToString(h[:16]) // intentional 128-bit truncation
}