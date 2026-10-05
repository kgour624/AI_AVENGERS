package training

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RetrievalConfig is the Phase 1-3 RAG runtime config (stored in system_settings).
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

// LoadRetrievalConfig reads retrieval_config from system_settings (best-effort, returns defaults on miss).
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
	return cfg
}

// LoadChunkerConfig reads chunker_config from system_settings (best-effort).
// ChunkerConfig HEAD has only TargetSize/MinSize/MaxSize/Overlap — map soft/hard aliases onto them.
func LoadChunkerConfig(ctx context.Context, db *pgxpool.Pool) ChunkerConfig {
	cfg := DefaultChunkerConfig()
	if db == nil {
		return cfg
	}
	var raw []byte
	if err := db.QueryRow(ctx, `SELECT value FROM system_settings WHERE key='chunker_config'`).Scan(&raw); err != nil {
		return cfg
	}
	var m map[string]interface{}
	if json.Unmarshal(raw, &m) != nil {
		return cfg
	}
	if v, ok := m["soft_limit"].(float64); ok {
		cfg.TargetSize = int(v)
	}
	if v, ok := m["target_size"].(float64); ok {
		cfg.TargetSize = int(v)
	}
	if v, ok := m["hard_limit"].(float64); ok {
		cfg.MaxSize = int(v)
	}
	if v, ok := m["max_size"].(float64); ok {
		cfg.MaxSize = int(v)
	}
	if v, ok := m["overlap_tokens"].(float64); ok {
		cfg.Overlap = int(v)
	}
	if v, ok := m["overlap"].(float64); ok {
		cfg.Overlap = int(v)
	}
	if v, ok := m["min_size"].(float64); ok {
		cfg.MinSize = int(v)
	}
	return cfg
}
