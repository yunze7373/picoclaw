package embedding

import (
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

// CachedProvider wraps an inner Provider with LRU caching.
type CachedProvider struct {
	inner    Provider
	mu       sync.Mutex
	cache    map[string][]float32
	order    []string // LRU order: front = oldest
	maxSize  int
}

// NewCachedProvider wraps p with an LRU cache of maxSize entries.
// If maxSize <= 0, defaultCacheSize (10000) is used.
func NewCachedProvider(p Provider, maxSize int) *CachedProvider {
	if maxSize <= 0 {
		maxSize = defaultCacheSize
	}
	return &CachedProvider{
		inner:   p,
		cache:   make(map[string][]float32, maxSize),
		order:   make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

// Model delegates to the inner provider.
func (c *CachedProvider) Model() string { return c.inner.Model() }

// Dims delegates to the inner provider.
func (c *CachedProvider) Dims() int { return c.inner.Dims() }

// Embed returns cached embeddings where available and calls the inner
// provider only for cache misses. Results are stored back into the cache.
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
	misses := make([]int, 0)
	missTexts := make([]string, 0)

	for i, k := range keys {
		if v, ok := c.cache[k]; ok {
			result[i] = v
			c.touchLRU(k)
		} else {
			misses = append(misses, i)
			missTexts = append(missTexts, texts[i])
		}
	}
	c.mu.Unlock()

	if len(misses) == 0 {
		return result, nil
	}

	// Call inner provider only for cache misses
	vectors, err := c.inner.Embed(ctx, missTexts)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	for j, idx := range misses {
		v := vectors[j]
		result[idx] = v
		c.store(keys[idx], v)
	}
	c.mu.Unlock()

	return result, nil
}

// store adds or updates a key in the LRU cache. Must be called with c.mu held.
func (c *CachedProvider) store(key string, v []float32) {
	if _, exists := c.cache[key]; !exists {
		if len(c.cache) >= c.maxSize {
			// Evict oldest
			oldest := c.order[0]
			c.order = c.order[1:]
			delete(c.cache, oldest)
		}
		c.order = append(c.order, key)
	}
	c.cache[key] = v
}

// touchLRU moves key to the back (most recently used). Must be called with c.mu held.
func (c *CachedProvider) touchLRU(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
	c.order = append(c.order, key)
}

// hashText returns a short stable hash for cache keying.
func hashText(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:16])
}
