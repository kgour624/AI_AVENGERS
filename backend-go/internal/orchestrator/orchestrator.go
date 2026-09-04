package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	appcontext "ai_avengers/backend/internal/context"
	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/decision"
	"ai_avengers/backend/internal/memory"
)

// OrchestratorRequest is the input to the orchestrator.
type OrchestratorRequest struct {
	ProjectID   uuid.UUID
	ClientID    uuid.UUID
	ChatID      uuid.UUID
	Message     string
	ExpertIDs   []uuid.UUID
	TurnNumber  int
}

// OrchestratorResponse is the full output including all expert responses.
type OrchestratorResponse struct {
	ExpertResponses []ExpertResponse
	Synthesis       *SynthesisResult
	TurnNumber      int
	TotalTokens     int
	DurationMs      int64
}

// ExpertResponse is one expert's response.
type ExpertResponse struct {
	ExpertID    uuid.UUID            `json:"expert_id"`
	ExpertName  string               `json:"expert_name"`
	Domain      string               `json:"domain"`
	Mode        decision.ResponseMode `json:"mode"`
	Content     string               `json:"content"`
	Citations   []chinawall.Citation `json:"citations"`
	Confidence  float64              `json:"confidence"`
	GateStopped int                  `json:"gate_stopped"`
	Warning     string               `json:"warning,omitempty"`
	Questions   []string             `json:"questions,omitempty"`
	Error       string               `json:"error,omitempty"`
}

// SynthesisResult holds the combined view when multiple experts respond.
type SynthesisResult struct {
	Agreements     []string       `json:"agreements"`
	Contradictions []Contradiction `json:"contradictions"`
	Summary        string         `json:"summary"`
}

// Contradiction is a point where two experts disagree.
type Contradiction struct {
	Topic     string `json:"topic"`
	ExpertA   string `json:"expert_a"`
	PositionA string `json:"position_a"`
	ExpertB   string `json:"expert_b"`
	PositionB string `json:"position_b"`
}

// expertRecord holds DB data for an expert.
type expertRecord struct {
	ID                   uuid.UUID
	Name                 string
	Domain               string
	ReasoningCharter     string
	ClarificationCharter map[string][]string
}

// Orchestrator coordinates multiple domain experts for a single request.
//
// WHY parallel goroutines:
// 3 experts sequential = 3x latency (15+ seconds).
// 3 experts parallel = same latency as 1 (5 seconds).
// Go goroutines make this trivial — this is exactly what they're for.
type Orchestrator struct {
	db          *pgxpool.Pool
	assembler   *appcontext.Assembler
	decisionEng *decision.Engine
	memManager  *memory.Manager
	logger      *zap.Logger
}

// NewOrchestrator creates a new orchestrator.
func NewOrchestrator(
	db *pgxpool.Pool,
	assembler *appcontext.Assembler,
	decisionEng *decision.Engine,
	memManager *memory.Manager,
	logger *zap.Logger,
) *Orchestrator {
	return &Orchestrator{
		db:          db,
		assembler:   assembler,
		decisionEng: decisionEng,
		memManager:  memManager,
		logger:      logger,
	}
}

// Process handles a client message through all selected experts.
//
// Mental execution:
// Client: "How should I design the user table?"
// Selected experts: [DB Expert, System Design Expert]
//
// Step 1: Load both experts from DB
// Step 2: Launch 2 goroutines simultaneously
//   Goroutine 1: DB Expert processes question
//   Goroutine 2: SD Expert processes question
// Step 3: Collect results (timeout: 30s)
// Step 4: If 2+ experts responded → synthesize
// Step 5: Update memory async
// Step 6: Return combined response
func (o *Orchestrator) Process(ctx context.Context, req OrchestratorRequest) (*OrchestratorResponse, error) {
	start := time.Now()

	if len(req.ExpertIDs) == 0 {
		return nil, fmt.Errorf("no experts selected")
	}

	// Load expert records
	experts, err := o.loadExperts(ctx, req.ExpertIDs)
	if err != nil {
		return nil, fmt.Errorf("load experts failed: %w", err)
	}
	if len(experts) == 0 {
		return nil, fmt.Errorf("no active experts found")
	}

	// Run experts in parallel
	resultCh := make(chan ExpertResponse, len(experts))
	var wg sync.WaitGroup

	for _, exp := range experts {
		wg.Add(1)
		go func(expert expertRecord) {
			defer wg.Done()
			result := o.processWithExpert(ctx, req, expert)
			resultCh <- result
		}(exp)
	}

	// Close channel when all goroutines finish
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results with timeout
	// WHY 30s timeout: LLM calls can be slow. Don't wait forever.
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var expertResponses []ExpertResponse
	for {
		select {
		case result, ok := <-resultCh:
			if !ok {
				goto collected
			}
			expertResponses = append(expertResponses, result)
		case <-timeoutCtx.Done():
			o.logger.Warn("orchestrator timeout", zap.Int("collected", len(expertResponses)))
			goto collected
		}
	}
collected:

	if len(expertResponses) == 0 {
		return nil, fmt.Errorf("all experts failed or timed out")
	}

	// Synthesize if multiple experts
	var synthesis *SynthesisResult
	if len(expertResponses) > 1 {
		synthesis = o.synthesize(expertResponses)
	}

	// Update memory async (non-blocking)
	go o.updateMemory(context.Background(), req, expertResponses)

	return &OrchestratorResponse{
		ExpertResponses: expertResponses,
		Synthesis:       synthesis,
		TurnNumber:      req.TurnNumber,
		DurationMs:      time.Since(start).Milliseconds(),
	}, nil
}

// processWithExpert runs one expert through the full pipeline.
func (o *Orchestrator) processWithExpert(ctx context.Context, req OrchestratorRequest, expert expertRecord) ExpertResponse {
	// Assemble context
	assembledCtx, err := o.assembler.Assemble(
		ctx, req.ChatID, req.ProjectID, expert.ID, req.Message, req.TurnNumber,
	)
	if err != nil {
		o.logger.Warn("context assembly failed", zap.String("expert", expert.Name), zap.Error(err))
		return ExpertResponse{
			ExpertID: expert.ID, ExpertName: expert.Name, Domain: expert.Domain,
			Mode: decision.ModeREFUSE, Content: "Context assembly failed", Error: err.Error(),
		}
	}

	// Get project summary for Gate 4
	projectSummary := ""
	if assembledCtx.RollingSummary != "" {
		projectSummary = assembledCtx.RollingSummary
	}

	// Run decision engine
	result, err := o.decisionEng.Process(
		ctx,
		req.Message,
		decision.Expert{
			ID:                   expert.ID,
			Name:                 expert.Name,
			Domain:               expert.Domain,
			ReasoningCharter:     expert.ReasoningCharter,
			ClarificationCharter: expert.ClarificationCharter,
		},
		assembledCtx.CourseChunks,
		projectSummary,
		1,
	)
	if err != nil {
		o.logger.Error("decision engine failed", zap.String("expert", expert.Name), zap.Error(err))
		return ExpertResponse{
			ExpertID: expert.ID, ExpertName: expert.Name, Domain: expert.Domain,
			Mode: decision.ModeREFUSE, Content: "Processing failed", Error: err.Error(),
		}
	}

	return ExpertResponse{
		ExpertID:    expert.ID,
		ExpertName:  expert.Name,
		Domain:      expert.Domain,
		Mode:        result.Mode,
		Content:     result.Content,
		Citations:   result.Citations,
		Confidence:  result.Confidence,
		GateStopped: result.GateStopped,
		Warning:     result.Warning,
		Questions:   result.Questions,
	}
}

// synthesize finds agreements and contradictions between expert responses.
func (o *Orchestrator) synthesize(responses []ExpertResponse) *SynthesisResult {
	// Simple synthesis: find ADVISE responses and note any WARN/REFUSE
	var advising []ExpertResponse
	var warnings []ExpertResponse

	for _, r := range responses {
		switch r.Mode {
		case decision.ModeADVISE:
			advising = append(advising, r)
		case decision.ModeWARN, decision.ModePUSHBACK:
			warnings = append(warnings, r)
		}
	}

	result := &SynthesisResult{}

	if len(advising) > 1 {
		result.Agreements = []string{"Multiple experts have relevant knowledge on this topic"}
	}

	for _, w := range warnings {
		for _, a := range advising {
			result.Contradictions = append(result.Contradictions, Contradiction{
				Topic:     "approach",
				ExpertA:   a.ExpertName,
				PositionA: "Proceed with implementation",
				ExpertB:   w.ExpertName,
				PositionB: w.Warning,
			})
		}
	}

	result.Summary = fmt.Sprintf("%d expert(s) responded. Review each response carefully.", len(responses))
	return result
}

// updateMemory records the turn in all memory levels.
func (o *Orchestrator) updateMemory(ctx context.Context, req OrchestratorRequest, responses []ExpertResponse) {
	for _, resp := range responses {
		if resp.Error != "" {
			continue
		}
		importance := 3
		if resp.Mode == decision.ModeADVISE {
			importance = 4
		}
		o.memManager.RecordTurn(
			ctx,
			req.ProjectID, resp.ExpertID, req.ClientID,
			req.ChatID, uuid.New(),
			req.TurnNumber,
			req.Message, resp.Content,
			string(resp.Mode), importance,
		)
	}
}

// loadExperts fetches expert records from DB.
func (o *Orchestrator) loadExperts(ctx context.Context, expertIDs []uuid.UUID) ([]expertRecord, error) {
	var experts []expertRecord
	for _, id := range expertIDs {
		var e expertRecord
		var clarJSON []byte
		err := o.db.QueryRow(ctx,
			`SELECT id, name, domain, COALESCE(reasoning_charter,''), clarification_charter
			 FROM experts
			 WHERE id=$1 AND is_active=TRUE AND is_training=FALSE AND deleted_at IS NULL`,
			id,
		).Scan(&e.ID, &e.Name, &e.Domain, &e.ReasoningCharter, &clarJSON)
		if err != nil {
			o.logger.Warn("expert not found or inactive", zap.String("id", id.String()))
			continue
		}
		// Parse clarification charter
		if len(clarJSON) > 0 {
			_ = parseJSON(clarJSON, &e.ClarificationCharter)
		}
		experts = append(experts, e)
	}
	return experts, nil
}

// parseJSON unmarshals JSON bytes into v.
func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
