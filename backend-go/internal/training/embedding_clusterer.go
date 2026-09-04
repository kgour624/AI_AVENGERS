package training

import (
	"context"
	"math"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/ml"
)

// EmbeddingClusterer groups chunks by semantic similarity.
// Used BEFORE LLM topic extraction to reduce LLM calls by 60%.
//
// WHY embedding clustering (Byte by Byte AI course):
// Course taught: similar texts end up nearby in embedding space.
// If chunks are semantically similar, they likely share the same topic.
// We can cluster them WITHOUT an LLM call, then only use LLM
// for ambiguous cases where similarity is low.
//
// Cost impact:
// Without clustering: 1000 chunks / 10 per batch = 100 LLM calls
// With clustering:    1000 chunks -> 50 clusters -> 50 LLM calls (50% reduction)
type EmbeddingClusterer struct {
	ml              *ml.SidecarClient
	similarityThreshold float32
	logger          *zap.Logger
}

// ChunkCluster groups chunks that share a topic.
type ChunkCluster struct {
	Chunks     []TextChunk
	Indices    []int    // original indices in chunks slice
	Centroid   []float32
	Topic      string   // filled after LLM extraction
}

// NewEmbeddingClusterer creates a new clusterer.
// threshold: cosine similarity above which chunks are in same cluster (0.75 recommended).
func NewEmbeddingClusterer(mlClient *ml.SidecarClient, threshold float32, logger *zap.Logger) *EmbeddingClusterer {
	return &EmbeddingClusterer{
		ml:                  mlClient,
		similarityThreshold: threshold,
		logger:              logger,
	}
}

// Cluster groups chunks by semantic similarity.
// Returns clusters where each cluster's chunks share a topic.
//
// Algorithm (greedy nearest-neighbor):
// 1. Embed all chunks
// 2. For each chunk, find if it belongs to an existing cluster
//    (cosine similarity > threshold with cluster centroid)
// 3. If yes: add to cluster, update centroid
// 4. If no: create new cluster
//
// Mental execution:
// chunks = ["sharding intro", "consistent hashing", "kafka basics", "kafka consumers"]
// embeddings generated for all 4
// chunk 0 -> new cluster A (sharding)
// chunk 1 -> similarity with A = 0.82 > 0.75 -> add to A
// chunk 2 -> similarity with A = 0.31 < 0.75 -> new cluster B (kafka)
// chunk 3 -> similarity with B = 0.88 > 0.75 -> add to B
// Result: 2 clusters instead of 4 LLM calls
func (c *EmbeddingClusterer) Cluster(ctx context.Context, chunks []TextChunk) ([]ChunkCluster, [][]float32, error) {
	if len(chunks) == 0 {
		return nil, nil, nil
	}

	// Embed all chunks
	texts := make([]string, len(chunks))
	for i, ch := range chunks {
		texts[i] = ch.Text
	}

	embeddings, err := c.ml.Embed(ctx, texts)
	if err != nil {
		return nil, nil, err
	}

	// Greedy clustering
	var clusters []ChunkCluster

	for i, chunk := range chunks {
		embedding := embeddings[i]
		bestCluster := -1
		bestSim := float32(0)

		// Find best matching cluster
		for ci, cluster := range clusters {
			sim := cosineSimilarity(embedding, cluster.Centroid)
			if sim > c.similarityThreshold && sim > bestSim {
				bestSim = sim
			bestCluster = ci
			}
		}

		if bestCluster >= 0 {
			// Add to existing cluster
			clusters[bestCluster].Chunks = append(clusters[bestCluster].Chunks, chunk)
			clusters[bestCluster].Indices = append(clusters[bestCluster].Indices, i)
			// Update centroid (running average)
			clusters[bestCluster].Centroid = updateCentroid(
				clusters[bestCluster].Centroid,
				embedding,
				len(clusters[bestCluster].Chunks),
			)
		} else {
			// New cluster
			clusters = append(clusters, ChunkCluster{
				Chunks:   []TextChunk{chunk},
				Indices:  []int{i},
				Centroid: embedding,
			})
		}
	}

	c.logger.Info("embedding clustering complete",
		zap.Int("chunks", len(chunks)),
		zap.Int("clusters", len(clusters)),
		zap.Float64("reduction_pct", float64(len(chunks)-len(clusters))/float64(len(chunks))*100),
	)

	return clusters, embeddings, nil
}

// cosineSimilarity computes cosine similarity between two vectors.
// Returns value in [-1, 1]. Higher = more similar.
// WHY cosine (Byte by Byte AI course):
// Course showed embedding space where similar texts are nearby.
// Cosine similarity measures angle between vectors, not magnitude.
// bge-base-en-v1.5 embeddings are L2-normalized, so cosine = dot product.
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(normA) * math.Sqrt(normB)))
}

// updateCentroid updates cluster centroid with a new vector (running average).
func updateCentroid(centroid, newVec []float32, n int) []float32 {
	if len(centroid) != len(newVec) {
		return centroid
	}
	updated := make([]float32, len(centroid))
	for i := range centroid {
		// Running average: new_centroid = (old_centroid * (n-1) + new_vec) / n
		updated[i] = (centroid[i]*float32(n-1) + newVec[i]) / float32(n)
	}
	return updated
}
