// Package mcpv2 — Isolated Hexagonal Module (R2). First file = package name (RULE 8-A:30).
// No import from internal/mcp (old). Only shared DB pool via port interface.
package mcpv2

import (
	"context"
	"go.uber.org/zap"
)

// TxStarter is App-layer transaction begin (Ultimate Go §7 — BEGIN at App, tx in ctx).
type TxStarter interface {
	Begin(ctx context.Context) (context.Context, error)
}

// Deps holds constructor injected dependencies (RULE 8-B:31 — DI mandatory).
// TxStarter optional — if nil, writes are non-transactional (fallback for tests).
type Deps struct {
	DB        DBPort
	Redis     RedisPort
	Gateway   GatewayPort
	Decision  DecisionPort
	Vector    VectorPort
	TxStarter TxStarter
	Logger    *zap.Logger
}

// Service is the App layer orchestrator. Thin, delegates to Business + Storage.
type Service struct {
	deps Deps
}

// NewService creates Service with DI — no globals (RULE 8-B:31).
func NewService(d Deps) *Service {
	return &Service{deps: d}
}