package mcpv2

import (
	"context"
	"fmt"

	"ai_avengers/backend/internal/ml"
)

type RealVectorAdapter struct{ embedder ml.Embedder }

func NewRealVectorAdapter(e ml.Embedder) *RealVectorAdapter { return &RealVectorAdapter{embedder: e} }

func (r *RealVectorAdapter) Embed(ctx context.Context, text string) ([]float32, error) {
	if r.embedder == nil {
		return nil, fmt.Errorf("embedder not wired")
	}
	vecs, err := r.embedder.Embed(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 || len(vecs[0]) == 0 {
		return nil, fmt.Errorf("empty embedding")
	}
	if len(vecs[0]) != 768 {
		return nil, fmt.Errorf("unexpected dim %d", len(vecs[0]))
	}
	return vecs[0], nil
}