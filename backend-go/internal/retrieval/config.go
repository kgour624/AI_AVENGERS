package retrieval

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RetrievalConfig is the single source of truth for retrieval tuning.
// Mirrors training.RetrievalConfig / knowledge.RetrievalConfig /
// context assembler inline struct — identical JSON keys so all packages
// decode the same system_settings.retrieval_config row.
type RetrievalConfig struct {
	EnableParentChild bool `json:"enable_parent_child"`
	ChildSoftLimit    int  `json:"child_soft_limit"`
	ChildHardLimit    int  `json:"child_hard_limit"`
	ParentSoftLimit   int  `json:"parent_soft_limit"`
	ParentHardLimit   int  `json:"parent_hard_limit"`
	OverlapTokens     int  `json:"overlap_tokens"`
	TopKChildren      int  `json:"top_k_children"`
	TopKParents       int  `json:"top_k_parents"`
	RerankTopN        int  `json:"rerank_top_n"`
	MaxHops           int  `json:"max_hops"`
	RRFK              int  `json:"rrf_k"`
	AtomicCodeFence   bool `json:"atomic_code_fence"`
}

func DefaultRetrievalConfig() RetrievalConfig {
	return RetrievalConfig{
		EnableParentChild: false,
		ChildSoftLimit: 150, ChildHardLimit: 220,
		ParentSoftLimit: 1200, ParentHardLimit: 1500,
		OverlapTokens: 20, TopKChildren: 50, TopKParents: 3,
		RerankTopN: 10, MaxHops: 3, RRFK: 60, AtomicCodeFence: true,
	}
}

// LoadRetrievalConfig best-effort from system_settings. Returns defaults on miss
// so old DB keeps working (same contract as knowledge/ training loaders).
func LoadRetrievalConfig(ctx context.Context, db *pgxpool.Pool) RetrievalConfig {
	cfg := DefaultRetrievalConfig()
	if db == nil {
		return cfg
	}
	var raw []byte
	if err := db.QueryRow(ctx, `SELECT value FROM system_settings WHERE key='retrieval_config'`).Scan(&raw); err != nil {
		return cfg
	}
	_ = json.Unmarshal(raw, &cfg)
	if cfg.TopKChildren == 0 {
		cfg.TopKChildren = 50
	}
	if cfg.TopKParents == 0 {
		cfg.TopKParents = 3
	}
	if cfg.RerankTopN == 0 {
		cfg.RerankTopN = 10
	}
	if cfg.RRFK == 0 {
		cfg.RRFK = 60
	}
	return cfg
}
