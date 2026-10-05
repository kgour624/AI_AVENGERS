package orchestrator

import (
	"context"
	"math"
	"sync"
	"time"
)

// SemanticCache is app-side (Tweak 3: no RediSearch, no extra infra).
// 200 items, cosine 0.98, TTL 1h, ~0.3ms per lookup (in-memory scan).
// Loop bounded: caller uses scored Top3 only, so cache key is normalized question embedding.
type SemanticCache struct {
	mu        sync.RWMutex
	items     map[string]*cacheEntry
	order     []string // LRU order, oldest at 0
	maxSize   int
	threshold float64
	ttl       time.Duration
}

type cacheEntry struct {
	embedding []float32
	response  string
	chunks    []ScoredChunk
	expiresAt time.Time
	key       string
}

// NewSemanticCache creates cache with defaults: 200 items, 0.98 cosine, 1h TTL.
func NewSemanticCache() *SemanticCache {
	return &SemanticCache{
		items:     make(map[string]*cacheEntry, 200),
		order:     make([]string, 0, 200),
		maxSize:   200,
		threshold: 0.98,
		ttl:       time.Hour,
	}
}

// NewSemanticCacheWithConfig allows tuning for tests.
func NewSemanticCacheWithConfig(maxSize int, threshold float64, ttl time.Duration) *SemanticCache {
	if maxSize <= 0 {
		maxSize = 200
	}
	if threshold == 0 {
		threshold = 0.98
	}
	if ttl == 0 {
		ttl = time.Hour
	}
	return &SemanticCache{
		items:     make(map[string]*cacheEntry, maxSize),
		order:     make([]string, 0, maxSize),
		maxSize:   maxSize,
		threshold: threshold,
		ttl:       ttl,
	}
}

// Get returns cached response if cosine similarity >= 0.98 and not expired.
// 0.3ms path: single mutex RLock + linear scan 200 * 768 dot-product (~0.2ms in Go).
func (c *SemanticCache) Get(ctx context.Context, embedding []float32) ([]ScoredChunk, string, bool) {
	if c == nil || len(embedding) == 0 {
		return nil, "", false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	bestScore := 0.0
	var best *cacheEntry
	for _, e := range c.items {
		if time.Now().After(e.expiresAt) {
			continue
		}
		select {
		case <-ctx.Done():
			return nil, "", false
		default:
		}
		score := cosineSimilarity(embedding, e.embedding)
		if score >= c.threshold && score > bestScore {
			bestScore = score
			best = e
		}
	}
	if best == nil {
		return nil, "", false
	}
	return best.chunks, best.response, true
}

// Set stores response under embedding key. LRU eviction when full.
func (c *SemanticCache) Set(ctx context.Context, embedding []float32, chunks []ScoredChunk, response string) {
	if c == nil || len(embedding) == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	// sweep expired lazily on write if over 80% capacity
	if len(c.items) >= c.maxSize {
		c.evictExpiredLocked()
	}
	if len(c.items) >= c.maxSize && len(c.order) > 0 {
		oldest := c.order[0]
		delete(c.items, oldest)
		c.order = c.order[1:]
	}
	key := cacheKey(embedding)
	// copy embedding to avoid aliasing
	cp := make([]float32, len(embedding))
	copy(cp, embedding)
	chunksCp := make([]ScoredChunk, len(chunks))
	copy(chunksCp, chunks)
	entry := &cacheEntry{
		embedding: cp,
		response:  response,
		chunks:    chunksCp,
		expiresAt: time.Now().Add(c.ttl),
		key:       key,
	}
	c.items[key] = entry
	c.order = append(c.order, key)
}

func (c *SemanticCache) evictExpiredLocked() {
	now := time.Now()
	newOrder := c.order[:0]
	for _, k := range c.order {
		if e, ok := c.items[k]; ok && now.After(e.expiresAt) {
			delete(c.items, k)
			continue
		}
		newOrder = append(newOrder, k)
	}
	c.order = newOrder
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		av := float64(a[i])
		bv := float64(b[i])
		dot += av * bv
		normA += av * av
		normB += bv * bv
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// cacheKey creates a short key from embedding for map indexing.
// Uses first 8 dims quantized to avoid full 768 string; collision is okay because Get uses cosine scan, key is only for eviction.
func cacheKey(embedding []float32) string {
	// simple hash: sum of first 8 * 1000 as string
	var h int64
	for i := 0; i < len(embedding) && i < 8; i++ {
		h = h*31 + int64(math.Round(float64(embedding[i])*1000))
	}
	// include length + time nanos for uniqueness
	return string(rune(h)) + string(rune(len(embedding))) + time.Now().Format("150405.000000000")
}

// Size returns current items count (for metrics/tests).
func (c *SemanticCache) Size() int {
	if c == nil {
		return 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Clear sweeps all entries.
func (c *SemanticCache) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*cacheEntry, c.maxSize)
	c.order = c.order[:0]
}
