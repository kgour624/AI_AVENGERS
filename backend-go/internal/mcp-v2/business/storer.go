// Package business — Business defines storage interface it needs (Ultimate Go §7).
// App will import Business (down), Business never imports App.
package business

import "context"

// Storer is the storage contract Business needs. Implemented by business/storage/postgres.go
// Discover interfaces, don't design — small, defined where needed ( §5 ).
type Storer interface {
	ListExperts(ctx context.Context) ([]Expert, error)
	GetExpert(ctx context.Context, expertID string) (Expert, error)
	SearchChunks(ctx context.Context, expertID string, embedding []float32, limit int, cursor string) ([]Chunk, string, int, error)
	SearchRepoChunks(ctx context.Context, expertID string, embedding []float32, limit int) ([]RepoChunk, error)
	InsertUsageLog(ctx context.Context, row UsageLog) error
	GetExpertLimits(ctx context.Context, expertID, platform string) (ExpertLimits, error)
	IncLimits(ctx context.Context, expertID, platform string, tokens int) error
}

// TxBeginner + CommitRollbacker for App-layer transactions ( §7 — BEGIN at App, tx in ctx )
type TxBeginner interface {
	Begin(ctx context.Context) (context.Context, error)
}
type CommitRollbacker interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}