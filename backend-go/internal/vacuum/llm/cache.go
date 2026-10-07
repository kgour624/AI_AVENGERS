package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"ai_avengers/backend/internal/observability"
)

// Cache — in-memory TTL cache for LLM Kachra spans (cost saver).
// I: chunkText sha256 -> P: lookup -> O: cached spans or miss.
// No DB, no mutation, pure perf. Size-capped 2000 entries, TTL 10m.
type Cache struct {
	mu   sync.RWMutex
	data map[string]cacheEntry
	ttl  time.Duration
}

type cacheEntry struct {
	spans []KachraSpan
	exp   time.Time
}

func NewCache(ttl time.Duration) *Cache {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &Cache{data: make(map[string]cacheEntry), ttl: ttl}
}

func cacheKey(chunkText string) string {
	h := sha256.Sum256([]byte(chunkText))
	return hex.EncodeToString(h[:])
}

func (c *Cache) Get(chunkText string) ([]KachraSpan, bool) {
	k := cacheKey(chunkText)
	c.mu.RLock()
	e, ok := c.data[k]
	c.mu.RUnlock()
	if !ok {
		observability.Global.IncCacheMiss()
		return nil, false
	}
	if time.Now().After(e.exp) {
		c.mu.Lock()
		delete(c.data, k)
		c.mu.Unlock()
		observability.Global.IncCacheMiss()
		return nil, false
	}
	observability.Global.IncCacheHit()
	out := make([]KachraSpan, len(e.spans))
	copy(out, e.spans)
	return out, true
}

func (c *Cache) Set(chunkText string, spans []KachraSpan) {
	k := cacheKey(chunkText)
	cp := make([]KachraSpan, len(spans))
	copy(cp, spans)
	c.mu.Lock()
	c.data[k] = cacheEntry{spans: cp, exp: time.Now().Add(c.ttl)}
	if len(c.data) > 2000 {
		now := time.Now()
		for kk, v := range c.data {
			if now.After(v.exp) {
				delete(c.data, kk)
			}
		}
		if len(c.data) > 2000 {
			n := 0
			for kk := range c.data {
				delete(c.data, kk)
				n++
				if n > 200 {
					break
				}
			}
		}
	}
	c.mu.Unlock()
}

// CachedKachraDetector decorates any KachraDetector with read-through cache.
type CachedKachraDetector struct {
	inner KachraDetector
	cache *Cache
}

func NewCachedKachraDetector(inner KachraDetector, cache *Cache) *CachedKachraDetector {
	if cache == nil {
		cache = NewCache(10 * time.Minute)
	}
	return &CachedKachraDetector{inner: inner, cache: cache}
}

func (d *CachedKachraDetector) Detect(ctx context.Context, chunkText string) ([]KachraSpan, error) {
	if d.cache != nil {
		if cached, ok := d.cache.Get(chunkText); ok {
			return cached, nil
		}
	}
	if d.inner == nil {
		return nil, nil
	}
	spans, err := d.inner.Detect(ctx, chunkText)
	if err != nil {
		return nil, err
	}
	if d.cache != nil {
		d.cache.Set(chunkText, spans)
	}
	return spans, nil
}

// BatchDetector — cost saver: merges tiny chunks (<500 chars) into batches for single LLM call.
// I: []chunkText -> P: batch small chunks, split large -> O: []KachraSpan per batch (caller demuxes).
// Minimal version: just passthrough with cache; true batching is caller-side via BatchedDetect.
type BatchDetector struct {
	inner KachraDetector
	cache *Cache
}

func NewBatchDetector(inner KachraDetector, cache *Cache) *BatchDetector {
	if cache == nil {
		cache = NewCache(10 * time.Minute)
	}
	return &BatchDetector{inner: inner, cache: cache}
}

// BatchedDetect groups chunkTexts by size, calls LLM per group, returns spans per group index.
func (b *BatchDetector) BatchedDetect(ctx context.Context, chunkTexts []string) ([][]KachraSpan, error) {
	out := make([][]KachraSpan, len(chunkTexts))
	// small-chunk batching: single call for all tiny chunks joined
	var tinyIdx []int
	var tinyJoined string
	for i, t := range chunkTexts {
		if len(t) < 500 {
			tinyIdx = append(tinyIdx, i)
			tinyJoined += "\n---CHUNK_BOUNDARY---\n" + t
		}
	}
	if len(tinyIdx) > 0 && b.inner != nil {
		if cached, ok := b.cache.Get(tinyJoined); ok {
			// best-effort: assign cached to first tiny chunk, rest empty (heuristic fallback)
			if len(cached) > 0 && len(tinyIdx) > 0 {
				out[tinyIdx[0]] = cached
			}
		} else {
			spans, err := b.inner.Detect(ctx, tinyJoined)
			if err == nil {
				b.cache.Set(tinyJoined, spans)
				if len(tinyIdx) > 0 {
					out[tinyIdx[0]] = spans
				}
			}
		}
	}
	// large chunks individual (with cache)
	for i, t := range chunkTexts {
		if len(t) >= 500 {
			if cached, ok := b.cache.Get(t); ok {
				out[i] = cached
				continue
			}
			if b.inner != nil {
				spans, err := b.inner.Detect(ctx, t)
				if err == nil {
					b.cache.Set(t, spans)
					out[i] = spans
				}
			}
		}
	}
	return out, nil
}

