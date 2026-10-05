package orchestrator

import (
	"container/heap"
	"sort"

	"github.com/google/uuid"
)

// ScoredChunk is a normalized retrieval unit for RRF merging.
// Text is the parent page_text (Phase 2) or chunk_text (legacy fallback).
// Score is rerank or ANN distance score; RRFScore is computed by mergeRRF.
type ScoredChunk struct {
	ID       uuid.UUID `json:"id"`
	Text     string    `json:"text"`
	Topic    string    `json:"topic"`
	Score    float64   `json:"score"`
	RRFScore float64   `json:"rrf_score"`
	Rank     int       `json:"rank"`
	Source   string    `json:"source"` // "parent-child" | "hybrid" | "graph"
}

// mergeRRF merges multiple ranked lists using Reciprocal Rank Fusion k≈60.
// Formula: RRFScore(d) = sum( 1 / (k + rank) ) across lists where d appears.
// k=60 per retrieval_config.RRFK (spec). Returns merged list sorted by RRFScore desc.
// Dedupes by chunk ID (uuid) — same parent appearing in multiple lists boosts score.
func mergeRRF(lists [][]ScoredChunk, k int) []ScoredChunk {
	if k <= 0 {
		k = 60
	}
	scores := make(map[uuid.UUID]*ScoredChunk, 64)
	ranks := make(map[uuid.UUID]float64, 64)
	for _, list := range lists {
		for rank, ch := range list {
			rrf := 1.0 / float64(k+rank+1) // rank 0 => 1/(k+1)
			if prev, ok := ranks[ch.ID]; ok {
				ranks[ch.ID] = prev + rrf
				// keep best Text/Score for final display (highest original Score wins)
				if ch.Score > scores[ch.ID].Score {
					scores[ch.ID].Text = ch.Text
					scores[ch.ID].Score = ch.Score
					scores[ch.ID].Topic = ch.Topic
				}
			} else {
				ranks[ch.ID] = rrf
				cp := ch
				cp.RRFScore = rrf
				cp.Rank = rank
				scores[ch.ID] = &cp
			}
		}
	}
	merged := make([]ScoredChunk, 0, len(scores))
	for id, ch := range scores {
		ch.RRFScore = ranks[id]
		merged = append(merged, *ch)
	}
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].RRFScore > merged[j].RRFScore
	})
	return merged
}

// pruneTopK returns top K by RRFScore using min-heap (O(n log K)).
// K=3 per spec (heap prune Top 3) — bounded loop, p95 safe.
func pruneTopK(chunks []ScoredChunk, k int) []ScoredChunk {
	if k <= 0 {
		k = 3
	}
	if len(chunks) <= k {
		sort.Slice(chunks, func(i, j int) bool { return chunks[i].RRFScore > chunks[j].RRFScore })
		return chunks
	}
	h := &minHeap{}
	heap.Init(h)
	for _, ch := range chunks {
		if h.Len() < k {
			heap.Push(h, ch)
			continue
		}
		if ch.RRFScore > (*h)[0].RRFScore {
			heap.Pop(h)
			heap.Push(h, ch)
		}
	}
	out := make([]ScoredChunk, h.Len())
	for i := len(out) - 1; i >= 0; i-- {
		out[i] = heap.Pop(h).(ScoredChunk)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RRFScore > out[j].RRFScore })
	return out
}

// minHeap implements heap.Interface for ScoredChunk by RRFScore ascending (min at top).
type minHeap []ScoredChunk

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].RRFScore < h[j].RRFScore }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{}) { *h = append(*h, x.(ScoredChunk)) }
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
