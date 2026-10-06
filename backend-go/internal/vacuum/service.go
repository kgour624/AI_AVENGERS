package vacuum

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/vacuum/brain"
	"ai_avengers/backend/internal/vacuum/engine"
	"ai_avengers/backend/internal/vacuum/llm"
)

// Service is the 3-firewall wiring: Business holds Brain Store + Engine + DB.
// API layer calls Service methods; Foundation is pgxpool + zap.
type Service struct {
	db              *pgxpool.Pool
	brain           *brain.Store
	engine          *engine.Engine
	classifier      llm.Classifier
	headingGen      llm.HeadingGenerator
	preservationGuard llm.PreservationGuard
	logger          *zap.Logger
}

func NewService(db *pgxpool.Pool, logger *zap.Logger) *Service {
	bs := brain.NewStore(db, logger)
	return &Service{db: db, brain: bs, engine: engine.New(bs), logger: logger}
}

func (s *Service) Brain() *brain.Store   { return s.brain }
func (s *Service) Engine() *engine.Engine { return s.engine }
func (s *Service) DB() *pgxpool.Pool      { return s.db }
func (s *Service) Classifier() llm.Classifier { return s.classifier }
func (s *Service) HeadingGen() llm.HeadingGenerator { return s.headingGen }
func (s *Service) Guard() llm.PreservationGuard { return s.preservationGuard }

// SetLLM wires Phase 5 LLM components. Nil args leave existing wiring.
func (s *Service) SetLLM(c llm.Classifier, hg llm.HeadingGenerator, g llm.PreservationGuard) {
	if c != nil {
		s.classifier = c
	}
	if hg != nil {
		s.headingGen = hg
	}
	if g != nil {
		s.preservationGuard = g
	}
}

// Start launches hot-reload ticker (30s) with context lifecycle.
func (s *Service) Start(ctx context.Context) { s.brain.Start(ctx) }

// KachraPattern CRUD helpers used by admin handler (App layer)
type KachraPattern struct {
	ID          string `json:"id"`
	Pattern     string `json:"pattern"`
	PatternType string `json:"pattern_type"`
	Category    string `json:"category"`
	IsActive    bool   `json:"is_active"`
	HitCount    int64  `json:"hit_count"`
	Version     int64  `json:"version"`
}

type CandidateRow struct {
	ID         string  `json:"id"`
	Pattern    string  `json:"pattern"`
	Category   string  `json:"category"`
	Status     string  `json:"status"`
	HitCount   int64   `json:"hit_count"`
	Confidence float64 `json:"confidence"`
	Context    string  `json:"context_snippet"`
}
