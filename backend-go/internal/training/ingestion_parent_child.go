package training

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// BuildParentChild decides chunking strategy based on RetrievalConfig flag.
// Additive: when flag false, behaves exactly like old Chunk() path.
// When true, uses ChunkMarkdown -> parents 1200-1500 tok + children 150-220 tok with atomic fence.
func BuildParentChild(ctx context.Context, chunker *TextChunker, raw string, rcfg RetrievalConfig) ([]ParentChunk, []TextChunk) {
	if !rcfg.EnableParentChild {
		return nil, chunker.Chunk(raw)
	}
	parentCfg := ChunkConfig{TargetSize: rcfg.ParentSoftLimit, MinSize: 900, MaxSize: rcfg.ParentHardLimit, Overlap: rcfg.OverlapTokens, SoftLimit: rcfg.ParentSoftLimit, HardLimit: rcfg.ParentHardLimit, OverlapTokens: rcfg.OverlapTokens, AtomicCodeFence: rcfg.AtomicCodeFence, WordCountEstimate: true}
	childCfg := ChunkConfig{TargetSize: rcfg.ChildSoftLimit, MinSize: 100, MaxSize: rcfg.ChildHardLimit, Overlap: rcfg.OverlapTokens, SoftLimit: rcfg.ChildSoftLimit, HardLimit: rcfg.ChildHardLimit, OverlapTokens: rcfg.OverlapTokens, AtomicCodeFence: rcfg.AtomicCodeFence, WordCountEstimate: true}
	if parentCfg.SoftLimit == 0 {
		parentCfg = DefaultParentConfig()
	}
	if childCfg.SoftLimit == 0 {
		childCfg = DefaultChildConfig()
	}
	parents, children := chunker.ChunkMarkdown(ctx, raw, parentCfg, childCfg)
	return parents, children
}

// AssignParentIDs maps each child to its parent page index and fills ParentIndex/ParentID.
// Parents are built by grouping consecutive children up to ParentSoftLimit, so child's parent
// is determined by cumulative token walk identical to buildParentsFromChildren.
func AssignParentIDs(children []TextChunk, parents []ParentChunk, parentIDByIndex map[int]uuid.UUID, rcfg RetrievalConfig) []TextChunk {
	if len(parents) == 0 || len(parentIDByIndex) == 0 {
		for i := range children {
			children[i].ParentIndex = -1
		}
		return children
	}
	// Build lookup: child index -> parent index by replaying same grouping with rcfg limits (FIX4)
	parentCfg := ChunkConfig{TargetSize: rcfg.ParentSoftLimit, MinSize: 900, MaxSize: rcfg.ParentHardLimit, SoftLimit: rcfg.ParentSoftLimit, HardLimit: rcfg.ParentHardLimit, OverlapTokens: rcfg.OverlapTokens, AtomicCodeFence: rcfg.AtomicCodeFence}
	if parentCfg.SoftLimit == 0 {
		parentCfg = DefaultParentConfig()
	}
	pHard := parentCfg.effectiveHard()
	pSoft := parentCfg.effectiveSoft()
	curTokens := 0
	curParent := 0
	for i := range children {
		ct := children[i].TokenCount
		if ct == 0 {
			ct = estimateTokensWithConfig(children[i].Text, true)
		}
		if curTokens+ct > pHard && curTokens > 0 {
			curParent++
			curTokens = 0
		}
		children[i].ParentIndex = curParent
		if id, ok := parentIDByIndex[curParent]; ok {
			children[i].ParentID = id.String()
		}
		curTokens += ct
		if curTokens >= pSoft {
			curParent++
			curTokens = 0
		}
		if curParent >= len(parents) {
			curParent = len(parents) - 1
		}
	}
	return children
}

// EnsureParentsStored stores parents if flag enabled, best-effort (never fails ingestion).
func EnsureParentsStored(ctx context.Context, db *pgxpool.Pool, logger *zap.Logger, expertID uuid.UUID, sourceFile string, parents []ParentChunk) map[int]uuid.UUID {
	if len(parents) == 0 {
		return map[int]uuid.UUID{}
	}
	ids, err := StoreParents(ctx, db, expertID, sourceFile, parents)
	if err != nil {
		if logger != nil {
			logger.Warn("parent store failed (non-fatal, children still stored)", zap.Error(err))
		}
		return map[int]uuid.UUID{}
	}
	return ids
}
