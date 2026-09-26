package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/byoexpert"
	"ai_avengers/backend/internal/category"
	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/docextract"
	"ai_avengers/backend/internal/eval"
	"ai_avengers/backend/internal/expertversion"
	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/jobevents"
	"ai_avengers/backend/internal/knowledge"
	"ai_avengers/backend/internal/ml"
	"ai_avengers/backend/internal/response"
	"ai_avengers/backend/internal/tenant"
	"ai_avengers/backend/internal/training"
	"ai_avengers/backend/internal/usage"
	"ai_avengers/backend/internal/workflow"
)

// AdminHandler handles all admin panel HTTP requests.
// All routes require admin role (enforced by AdminMiddleware).
type AdminHandler struct {
	db          *pgxpool.Pool
	gateway     *gateway.ModelGateway
	embedder    ml.Embedder // ml.Embedder interface: sidecar or CodeCraftAPI, resolved at call time
	ingestion   *training.IngestionPipeline
	// reconcile audits an expert's corpus against the run ledger and repairs the
	// derived state (Phase D). Built here from the same pool + embedder the
	// pipeline uses, so a repair cannot embed with a different model than
	// ingestion used.
	reconcile *training.Reconciler
	// capabilityEval (I2) measures what an expert can actually answer. Nil-safe:
	// when unwired the endpoints report that measurement is unavailable rather
	// than pretending a score exists.
	capEval *training.CapabilityEvaluator
	// concepts (I4) stores the typed relationships between an expert's topics. Nil-safe:
	// when unwired the concept endpoints report that the graph is unavailable.
	concepts *training.ConceptGraph
	// depthLayers (I5) classifies chunks by content kind and reports coverage. Nil-safe:
	// when unwired the depth endpoints report that it is unavailable.
	depthLayers *training.DepthClassifier
	categoryReg *category.Registry
	// domainReg backs the domain-profile admin endpoints (max tokens,
	// coverage/citation/strip modes, etc — see ListDomainProfiles /
	// GetDomainProfile / UpdateDomainProfile below). SAME registry
	// instance wired at startup in cmd/server/main.go and used by
	// chinawall.Enforcer — writes here take effect for the very next
	// question with zero redeploy, exactly like categoryReg above.
	domainReg *chinawall.DomainRegistry
	// versions (C2): expert versioning + capability drift. Nil-safe — when
	// unset, ingestion still runs, only snapshots/drift are skipped.
	versions *expertversion.Service
	// evals (C3): golden-set run store. Nil-safe — view endpoints return empty.
	evals *eval.Store
	// tenants (C4): enterprise isolation controls. Nil-safe — list returns empty.
	tenants *tenant.Service
	// usage (C5): cost/usage analytics + budgets. Nil-safe — returns empty.
	usage *usage.Service
	// freshness (C6): knowledge staleness/refresh tasks. Nil-safe.
	freshness *knowledge.Freshness
	// byo (C8): tenant self-service expert entitlement + audit. Nil-safe.
	byo    *byoexpert.Service
	// events (T1): durable ingestion timeline. Powers the live SSE stream
	// (true push, not a 1s snapshot poll) and the history endpoint. Nil-safe:
	// when unwired, StreamIngestionJob falls back to snapshot-only updates.
	events *jobevents.Store
	// extractor (D3): converts an uploaded document to text (stage 0) before
	// the pipeline runs. Nil-safe: disables office/PDF formats, keeps .txt/.md.
	extractor *docextract.Extractor
	logger    *zap.Logger
}

// NewAdminHandler creates a new admin handler.
//
// embedder satisfies ml.Embedder — either *ml.SidecarClient (default) or
// *ml.DynamicEmbedder (when CodeCraftAPI embeddings are enabled).
// Passed to IngestionPipeline so transcript ingestion uses the same
// embedding source as the rest of the system.
//
// categoryReg is used by the expert-category CRUD handlers (CT-A3) and
// by CreateExpert/UpdateExpert's category_id validation (CT-A4). It is
// the SAME registry instance wired at startup in cmd/server/main.go —
// writes here must call categoryReg.Reload(ctx) afterward so the cache
// used elsewhere (e.g. future chinawall/decision-engine lookups) never
// goes stale relative to what admin just wrote.
func NewAdminHandler(
	db *pgxpool.Pool,
	gw *gateway.ModelGateway,
	mlClient *ml.SidecarClient,
	embedder ml.Embedder,
	categoryReg *category.Registry,
	domainReg *chinawall.DomainRegistry,
	versions *expertversion.Service,
	evals *eval.Store,
	tenants *tenant.Service,
	usageSvc *usage.Service,
	freshness *knowledge.Freshness,
	byoSvc *byoexpert.Service,
	events *jobevents.Store,
	extractor *docextract.Extractor,
	logger *zap.Logger,
) *AdminHandler {
	return &AdminHandler{
		db:          db,
		gateway:     gw,
		embedder:    embedder,
		ingestion:   training.NewIngestionPipeline(db, embedder, mlClient, gw, events, extractor, logger),
		reconcile:   training.NewReconciler(db, embedder, logger),
		categoryReg: categoryReg,
		domainReg:   domainReg,
		versions:    versions,
		evals:       evals,
		tenants:     tenants,
		usage:       usageSvc,
		freshness:   freshness,
		byo:         byoSvc,
		events:      events,
		extractor:   extractor,
		logger:      logger,
	}
}

// SetCapabilityEvaluator wires capability measurement (I2).
//
// WHY a setter and not a constructor parameter: the evaluator needs the context
// assembler, which is built after this handler in cmd/server/main.go. A setter
// keeps the dependency explicit at the wiring site instead of adding a sixteenth
// positional argument that every caller has to keep in order.
func (h *AdminHandler) SetCapabilityEvaluator(e *training.CapabilityEvaluator) {
	h.capEval = e

	// I3: the ingest gate reads a MEASURED capability, so the pipeline needs a way to
	// trigger a pass. It is handed in as a single function rather than as the
	// evaluator itself, so the pipeline never learns about the context assembler the
	// measurement depends on. Not wiring it is a supported state: the gate then
	// reports the capability as unmeasured and the expert stays in draft.
	if h.ingestion != nil {
		h.ingestion.SetCapabilityMeasurer(func(ctx context.Context, expertID uuid.UUID, topics int) error {
			_, err := e.RunEval(ctx, expertID, training.CapabilityEvalRequest{Topics: topics})
			return err
		})
	}
}

// SetConceptGraph wires the concept-relationship store (I4).
func (h *AdminHandler) SetConceptGraph(g *training.ConceptGraph) {
	h.concepts = g
}

// SetDepthClassifier wires the content-kind classifier (I5).
func (h *AdminHandler) SetDepthClassifier(c *training.DepthClassifier) {
	h.depthLayers = c
}

// ============================================================
// EXPERT MANAGEMENT
// ============================================================

// adminExpertRow is the admin-facing view of an expert.
// Includes all config fields from migration 006 that are admin-only.
// WHY separate from ExpertPublic in expert/handler.go:
//   Clients must NOT see model_tier, temperature, loop_pattern, etc.
//   These are internal config. Separation enforces this at the type level.
type adminExpertRow struct {
	ID                 uuid.UUID       `json:"id"`
	Name               string          `json:"name"`
	Slug               string          `json:"slug"`
	Domain             string          `json:"domain"`
	Description        string          `json:"description"`
	// ReasoningCharter fix (2026-09-08): previously never SELECTed here,
	// so the admin "Edit Charter" modal always opened with an empty
	// textarea even though the charter WAS saved correctly in the DB by
	// the ingestion pipeline (charter_extractor.go -> IngestTranscript's
	// Step 7 UPDATE). The gap was purely on the read side: this list
	// endpoint is the only place AdminExperts.tsx sources expert data
	// from, and it silently omitted the column.
	ReasoningCharter   string          `json:"reasoning_charter"`
	TotalChunks        int             `json:"total_chunks"`
	TotalTopics        int             `json:"total_topics"`
	AvgDepth           float64         `json:"avg_depth_level"`
	AvgRating          float64         `json:"avg_rating"`
	TotalRatings       int             `json:"total_ratings"`
	IsActive           bool            `json:"is_active"`
	IsTraining         bool            `json:"is_training"`
	// Migration 006 fields
	ModelTier          string          `json:"model_tier"`
	Temperature        float64         `json:"temperature"`
	TopP               float64         `json:"top_p"`
	LoopPattern        string          `json:"loop_pattern"`
	MaxLoopIterations  int             `json:"max_loop_iterations"`
	AllowedTools       json.RawMessage `json:"allowed_tools"`
	TrainingStatus     string          `json:"training_status"`
	// CategoryID (migration 010, CT-A4): nullable. FIX (2026-09-08): this
	// field existed on the DB row and was already PATCH-able via
	// UpdateExpert (category_id below), but was never SELECTed/returned
	// here — the admin UI had no way to SEE or edit which category an
	// expert is in, so a mismatch introduced by direct-SQL retrofitting
	// an existing expert into a category (done once, manually, before
	// this admin UI existed) was invisible and uncorrectable from the
	// panel. Read-side gap only, same class of bug as ReasoningCharter
	// above.
	CategoryID         *uuid.UUID      `json:"category_id"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// ListExperts GET /admin/experts
// Returns all experts (including draft/deprecated) with full config.
func (h *AdminHandler) ListExperts(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT id, name, slug, domain, COALESCE(description,''),
		       COALESCE(reasoning_charter,''),
		       total_chunks, total_topics, COALESCE(avg_depth_level,0),
		       COALESCE(avg_rating,0), total_ratings,
		       is_active, is_training,
		       model_tier, temperature, top_p,
		       loop_pattern, max_loop_iterations, allowed_tools,
		       training_status, category_id,
		       created_at, updated_at
		FROM experts
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC`)
	if err != nil {
		h.logger.Error("list experts failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	defer rows.Close()

	var experts []adminExpertRow
	for rows.Next() {
		var e adminExpertRow
		if err := rows.Scan(
			&e.ID, &e.Name, &e.Slug, &e.Domain, &e.Description,
			&e.ReasoningCharter,
			&e.TotalChunks, &e.TotalTopics, &e.AvgDepth,
			&e.AvgRating, &e.TotalRatings,
			&e.IsActive, &e.IsTraining,
			&e.ModelTier, &e.Temperature, &e.TopP,
			&e.LoopPattern, &e.MaxLoopIterations, &e.AllowedTools,
			&e.TrainingStatus, &e.CategoryID,
			&e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			h.logger.Warn("scan expert row failed", zap.Error(err))
			continue
		}
		experts = append(experts, e)
	}
	if experts == nil {
		experts = []adminExpertRow{}
	}
	response.OK(c, experts)
}

// validModelTiers is the set of allowed model_tier values.
// Must match the CHECK constraint in migration 006.
var validModelTiers = map[string]bool{"cheap": true, "strong": true, "fast": true}

// validLoopPatterns is the set of allowed loop_pattern values.
// Must match the CHECK constraint in migration 006.
var validLoopPatterns = map[string]bool{"ota": true, "react": true, "plan_execute": true}

// CreateExpert POST /admin/experts
// Required: name, slug, domain.
// Optional (migration 006 config fields): model_tier, temperature, top_p,
// loop_pattern, max_loop_iterations, allowed_tools.
// Optional (migration 010, CT-A4): category_id. NULLABLE at the DB level
// per CATEGORY_TEMPLATE_HANDOFF.md CT-L2 — omitting it is valid and keeps
// the expert on flat-text behavior. If provided, it must reference an
// existing expert_categories row or this returns 400.
// Omitting optional fields uses the DB column defaults.
func (h *AdminHandler) CreateExpert(c *gin.Context) {
	var req struct {
		// Required
		Name        string `json:"name" binding:"required"`
		Slug        string `json:"slug" binding:"required"`
		Domain      string `json:"domain" binding:"required"`
		Description string `json:"description"`
		// Optional — migration 006 config fields
		ModelTier         *string  `json:"model_tier"`
		Temperature       *float64 `json:"temperature"`
		TopP              *float64 `json:"top_p"`
		LoopPattern       *string  `json:"loop_pattern"`
		MaxLoopIterations *int     `json:"max_loop_iterations"`
		AllowedTools      []string `json:"allowed_tools"`
		// Optional — migration 010 field (CT-A4)
		CategoryID *uuid.UUID `json:"category_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	// Validate optional enum/range fields before hitting the DB.
	// WHY validate here and not rely on DB CHECK:
	//   DB CHECK gives a cryptic pgx error. We return a clear 400.
	if req.ModelTier != nil && !validModelTiers[*req.ModelTier] {
		response.BadRequest(c, "INVALID_MODEL_TIER", "model_tier must be cheap, strong, or fast")
		return
	}
	if req.Temperature != nil && (*req.Temperature < 0.0 || *req.Temperature > 2.0) {
		response.BadRequest(c, "INVALID_TEMPERATURE", "temperature must be between 0.0 and 2.0")
		return
	}
	if req.TopP != nil && (*req.TopP < 0.0 || *req.TopP > 1.0) {
		response.BadRequest(c, "INVALID_TOP_P", "top_p must be between 0.0 and 1.0")
		return
	}
	if req.LoopPattern != nil && !validLoopPatterns[*req.LoopPattern] {
		response.BadRequest(c, "INVALID_LOOP_PATTERN", "loop_pattern must be ota, react, or plan_execute")
		return
	}
	if req.MaxLoopIterations != nil && (*req.MaxLoopIterations < 1 || *req.MaxLoopIterations > 50) {
		response.BadRequest(c, "INVALID_MAX_LOOP_ITERATIONS", "max_loop_iterations must be between 1 and 50")
		return
	}
	// WHY check registry cache, not a fresh DB query: category_id
	// validity check happens on every expert create — using the
	// already-loaded in-memory cache (category.Registry.Get) avoids an
	// extra round trip and matches how chinawall.DomainRegistry.Get is
	// used as the hot-path lookup elsewhere in this codebase.
	if req.CategoryID != nil && h.categoryReg.Get(*req.CategoryID) == nil {
		response.BadRequest(c, "INVALID_CATEGORY_ID", "category_id does not reference an existing category")
		return
	}

	// Serialize allowed_tools to JSON.
	// WHY: Postgres JSONB column expects a JSON string, not a Go slice.
	allowedToolsJSON := []byte("[]")
	if len(req.AllowedTools) > 0 {
		var err error
		allowedToolsJSON, err = json.Marshal(req.AllowedTools)
		if err != nil {
			response.BadRequest(c, "INVALID_ALLOWED_TOOLS", "allowed_tools must be a valid JSON array")
			return
		}
	}

	// Apply defaults for omitted optional fields.
	modelTier := "strong"
	if req.ModelTier != nil {
		modelTier = *req.ModelTier
	}
	temperature := 0.30
	if req.Temperature != nil {
		temperature = *req.Temperature
	}
	topP := 0.50
	if req.TopP != nil {
		topP = *req.TopP
	}
	loopPattern := "react"
	if req.LoopPattern != nil {
		loopPattern = *req.LoopPattern
	}
	maxLoopIterations := 5
	if req.MaxLoopIterations != nil {
		maxLoopIterations = *req.MaxLoopIterations
	}

	var id uuid.UUID
	err := h.db.QueryRow(c.Request.Context(),
		`INSERT INTO experts
		  (name, slug, domain, description,
		   model_tier, temperature, top_p,
		   loop_pattern, max_loop_iterations, allowed_tools, category_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id`,
		req.Name, req.Slug, req.Domain, req.Description,
		modelTier, temperature, topP,
		loopPattern, maxLoopIterations, string(allowedToolsJSON), req.CategoryID,
	).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			response.Conflict(c, "slug already exists")
			return
		}
		h.logger.Error("create expert failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.Created(c, map[string]interface{}{"id": id, "slug": req.Slug})
}

// validTrainingStatuses is the set of allowed training_status values.
// Must match the CHECK constraint in migration 006.
var validTrainingStatuses = map[string]bool{
	"draft": true, "ingesting": true, "trained": true, "deprecated": true,
}

// UpdateExpert PATCH /admin/experts/:id
// All fields are optional. Only provided fields are updated.
// Existing fields: name, description, is_active, reasoning_charter.
// New fields (migration 006): model_tier, temperature, top_p,
// loop_pattern, max_loop_iterations, allowed_tools, training_status.
func (h *AdminHandler) UpdateExpert(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	var req struct {
		// Existing fields
		Name             *string `json:"name"`
		Description      *string `json:"description"`
		IsActive         *bool   `json:"is_active"`
		ReasoningCharter *string `json:"reasoning_charter"`
		// Migration 006 config fields
		ModelTier         *string  `json:"model_tier"`
		Temperature       *float64 `json:"temperature"`
		TopP              *float64 `json:"top_p"`
		LoopPattern       *string  `json:"loop_pattern"`
		MaxLoopIterations *int     `json:"max_loop_iterations"`
		AllowedTools      []string `json:"allowed_tools"`
		TrainingStatus    *string  `json:"training_status"`
		// Migration 010 field (CT-A4). Pointer-to-pointer would be needed
		// to distinguish "omit" from "explicitly clear to NULL" — but this
		// handler's existing convention (see every other *T field above)
		// only supports set-if-present, never explicit-clear, so category_id
		// follows that same convention for consistency. Clearing a category
		// assignment is not a requirement from CATEGORY_TEMPLATE_HANDOFF.md
		// and can be added later as its own explicit endpoint if needed.
		CategoryID *uuid.UUID `json:"category_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if req.CategoryID != nil && h.categoryReg.Get(*req.CategoryID) == nil {
		response.BadRequest(c, "INVALID_CATEGORY_ID", "category_id does not reference an existing category")
		return
	}

	// Validate new fields before any DB writes.
	if req.ModelTier != nil && !validModelTiers[*req.ModelTier] {
		response.BadRequest(c, "INVALID_MODEL_TIER", "model_tier must be cheap, strong, or fast")
		return
	}
	if req.Temperature != nil && (*req.Temperature < 0.0 || *req.Temperature > 2.0) {
		response.BadRequest(c, "INVALID_TEMPERATURE", "temperature must be between 0.0 and 2.0")
		return
	}
	if req.TopP != nil && (*req.TopP < 0.0 || *req.TopP > 1.0) {
		response.BadRequest(c, "INVALID_TOP_P", "top_p must be between 0.0 and 1.0")
		return
	}
	if req.LoopPattern != nil && !validLoopPatterns[*req.LoopPattern] {
		response.BadRequest(c, "INVALID_LOOP_PATTERN", "loop_pattern must be ota, react, or plan_execute")
		return
	}
	if req.MaxLoopIterations != nil && (*req.MaxLoopIterations < 1 || *req.MaxLoopIterations > 50) {
		response.BadRequest(c, "INVALID_MAX_LOOP_ITERATIONS", "max_loop_iterations must be between 1 and 50")
		return
	}
	if req.TrainingStatus != nil && !validTrainingStatuses[*req.TrainingStatus] {
		response.BadRequest(c, "INVALID_TRAINING_STATUS", "training_status must be draft, ingesting, trained, or deprecated")
		return
	}

	ctx := c.Request.Context()

	// Existing fields
	if req.Name != nil {
		if _, err := h.db.Exec(ctx,
			`UPDATE experts SET name=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
			*req.Name, id); err != nil {
			h.logger.Error("update expert name failed", zap.Error(err))
		}
	}
	if req.Description != nil {
		if _, err := h.db.Exec(ctx,
			`UPDATE experts SET description=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
			*req.Description, id); err != nil {
			h.logger.Error("update expert description failed", zap.Error(err))
		}
	}
	if req.IsActive != nil {
		if _, err := h.db.Exec(ctx,
			`UPDATE experts SET is_active=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
			*req.IsActive, id); err != nil {
			h.logger.Error("update expert is_active failed", zap.Error(err))
		}
	}
	if req.ReasoningCharter != nil {
		if _, err := h.db.Exec(ctx,
			`UPDATE experts SET reasoning_charter=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
			*req.ReasoningCharter, id); err != nil {
			h.logger.Error("update expert reasoning_charter failed", zap.Error(err))
		}
	}

	// Migration 006 config fields
	if req.ModelTier != nil {
		if _, err := h.db.Exec(ctx,
			`UPDATE experts SET model_tier=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
			*req.ModelTier, id); err != nil {
			h.logger.Error("update expert model_tier failed", zap.Error(err))
		}
	}
	if req.Temperature != nil {
		if _, err := h.db.Exec(ctx,
			`UPDATE experts SET temperature=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
			*req.Temperature, id); err != nil {
			h.logger.Error("update expert temperature failed", zap.Error(err))
		}
	}
	if req.TopP != nil {
		if _, err := h.db.Exec(ctx,
			`UPDATE experts SET top_p=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
			*req.TopP, id); err != nil {
			h.logger.Error("update expert top_p failed", zap.Error(err))
		}
	}
	if req.LoopPattern != nil {
		if _, err := h.db.Exec(ctx,
			`UPDATE experts SET loop_pattern=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
			*req.LoopPattern, id); err != nil {
			h.logger.Error("update expert loop_pattern failed", zap.Error(err))
		}
	}
	if req.MaxLoopIterations != nil {
		if _, err := h.db.Exec(ctx,
			`UPDATE experts SET max_loop_iterations=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
			*req.MaxLoopIterations, id); err != nil {
			h.logger.Error("update expert max_loop_iterations failed", zap.Error(err))
		}
	}
	if req.AllowedTools != nil {
		allowedToolsJSON, err := json.Marshal(req.AllowedTools)
		if err != nil {
			response.BadRequest(c, "INVALID_ALLOWED_TOOLS", "allowed_tools must be a valid JSON array")
			return
		}
		if _, err := h.db.Exec(ctx,
			`UPDATE experts SET allowed_tools=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
			string(allowedToolsJSON), id); err != nil {
			h.logger.Error("update expert allowed_tools failed", zap.Error(err))
		}
	}
	if req.TrainingStatus != nil {
		// WHY also sync is_training:
		//   is_training=TRUE means ingestion is in progress.
		//   When admin manually sets training_status='trained' or 'deprecated',
		//   is_training must be FALSE so the expert appears in public listings.
		isTraining := *req.TrainingStatus == "ingesting"
		if _, err := h.db.Exec(ctx,
			`UPDATE experts SET training_status=$1, is_training=$2, updated_at=NOW()
			 WHERE id=$3 AND deleted_at IS NULL`,
			*req.TrainingStatus, isTraining, id); err != nil {
			h.logger.Error("update expert training_status failed", zap.Error(err))
		}
	}
	if req.CategoryID != nil {
		if _, err := h.db.Exec(ctx,
			`UPDATE experts SET category_id=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
			*req.CategoryID, id); err != nil {
			h.logger.Error("update expert category_id failed", zap.Error(err))
		}
	}

	response.OK(c, map[string]string{"status": "updated"})
}

// ============================================================
// EXPERT CATEGORIES (migration 010, CT-A3)
// ============================================================

// categoryRow is the wire shape for a single expert_categories row.
// WHY a distinct struct from category.Category: the registry's
// in-memory type is optimized for lookup (unexported cache fields
// live alongside it in registry.go); this is the explicit JSON
// response contract for the admin API, matching the same separation
// already used elsewhere in this file (adminExpertRow vs the DB-layer
// expertRecord type in orchestrator.go).
type categoryRow struct {
	ID                     uuid.UUID       `json:"id"`
	Name                   string          `json:"name"`
	Slug                   string          `json:"slug"`
	Description            string          `json:"description"`
	TemplateSchema         json.RawMessage `json:"template_schema"`
	DefaultLanguage        string          `json:"default_language"`
	AskStructurePermission bool            `json:"ask_structure_permission"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
}

// ListExpertCategories GET /admin/expert-categories
func (h *AdminHandler) ListExpertCategories(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT id, name, slug, COALESCE(description,''), template_schema,
		       default_language, ask_structure_permission, created_at, updated_at
		FROM expert_categories
		ORDER BY created_at DESC`)
	if err != nil {
		h.logger.Error("list expert categories failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	defer rows.Close()

	var categories []categoryRow
	for rows.Next() {
		var cat categoryRow
		if err := rows.Scan(
			&cat.ID, &cat.Name, &cat.Slug, &cat.Description, &cat.TemplateSchema,
			&cat.DefaultLanguage, &cat.AskStructurePermission, &cat.CreatedAt, &cat.UpdatedAt,
		); err != nil {
			h.logger.Warn("scan category row failed", zap.Error(err))
			continue
		}
		categories = append(categories, cat)
	}
	if categories == nil {
		categories = []categoryRow{}
	}
	response.OK(c, categories)
}

// GetExpertCategory GET /admin/expert-categories/:id
func (h *AdminHandler) GetExpertCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid category ID")
		return
	}
	var cat categoryRow
	err = h.db.QueryRow(c.Request.Context(), `
		SELECT id, name, slug, COALESCE(description,''), template_schema,
		       default_language, ask_structure_permission, created_at, updated_at
		FROM expert_categories WHERE id=$1`, id,
	).Scan(
		&cat.ID, &cat.Name, &cat.Slug, &cat.Description, &cat.TemplateSchema,
		&cat.DefaultLanguage, &cat.AskStructurePermission, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		response.NotFound(c, "category")
		return
	}
	response.OK(c, cat)
}

// templateSchemaInput is the request shape for template_schema on
// create/update. Validated against category.ValidSectionTypes before
// any DB write — a category with an invalid section type would let an
// expert silently fall back to flat-text generation later (CT-B scope)
// with no clear error at the point the mistake was actually made.
type templateSchemaInput struct {
	Sections []struct {
		Key         string `json:"key"`
		Label       string `json:"label"`
		Type        string `json:"type"`
		Required    bool   `json:"required"`
		// Description (2026-09-08 RCA fix): optional. See
		// category.TemplateSection.Description's doc comment for the
		// full root-cause explanation - this is the admin-facing input
		// wire shape for that same field.
		Description string `json:"description"`
	} `json:"sections"`
}

// validateTemplateSchema checks every section's "type" against
// category.ValidSectionTypes (prose | code | test_cases) and that key
// values are non-empty and unique. Returns a human-readable error on
// the first problem found, or nil if the schema is well-formed.
// An empty/omitted sections list is valid — it means "no structured
// template", the documented flat-text fallback (CT-L2).
func validateTemplateSchema(schema templateSchemaInput) error {
	seenKeys := make(map[string]bool)
	for _, s := range schema.Sections {
		if s.Key == "" {
			return fmt.Errorf("every section must have a non-empty key")
		}
		if seenKeys[s.Key] {
			return fmt.Errorf("duplicate section key: %s", s.Key)
		}
		seenKeys[s.Key] = true
		if !category.ValidSectionTypes[category.SectionType(s.Type)] {
			return fmt.Errorf("section %q has invalid type %q — must be prose, code, or test_cases", s.Key, s.Type)
		}
	}
	return nil
}

// CreateExpertCategory POST /admin/expert-categories
func (h *AdminHandler) CreateExpertCategory(c *gin.Context) {
	var req struct {
		Name                   string              `json:"name" binding:"required"`
		Slug                   string              `json:"slug" binding:"required"`
		Description            string              `json:"description"`
		TemplateSchema         templateSchemaInput `json:"template_schema"`
		DefaultLanguage        *string             `json:"default_language"`
		AskStructurePermission bool                `json:"ask_structure_permission"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if err := validateTemplateSchema(req.TemplateSchema); err != nil {
		response.BadRequest(c, "INVALID_TEMPLATE_SCHEMA", err.Error())
		return
	}

	// WHY default "java" here at the code level, not required from the
	// admin request: CATEGORY_TEMPLATE_HANDOFF.md CT-L5 — "default_language
	// ... set at the code level (not admin-required input), overridable
	// per category". Matches the DB column default exactly so behavior is
	// identical whether or not this field is included in the request body.
	defaultLanguage := "java"
	if req.DefaultLanguage != nil && *req.DefaultLanguage != "" {
		defaultLanguage = *req.DefaultLanguage
	}

	schemaJSON, err := json.Marshal(req.TemplateSchema)
	if err != nil {
		response.BadRequest(c, "INVALID_TEMPLATE_SCHEMA", "template_schema must be valid JSON")
		return
	}

	adminID := c.MustGet("user_id").(uuid.UUID)

	var id uuid.UUID
	err = h.db.QueryRow(c.Request.Context(),
		`INSERT INTO expert_categories
			(name, slug, description, template_schema, default_language, ask_structure_permission, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		req.Name, req.Slug, req.Description, string(schemaJSON),
		defaultLanguage, req.AskStructurePermission, adminID,
	).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			response.Conflict(c, "slug already exists")
			return
		}
		h.logger.Error("create expert category failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	// Refresh the shared registry cache so the newly created category is
	// immediately visible to Get()/GetBySlug() callers without a restart.
	// Non-fatal if this fails — the row is already durably in Postgres;
	// the cache will self-heal on the next natural Reload() or restart.
	if err := h.categoryReg.Reload(c.Request.Context()); err != nil {
		h.logger.Warn("category registry reload after create failed", zap.Error(err))
	}

	response.Created(c, map[string]interface{}{"id": id, "slug": req.Slug})
}

// UpdateExpertCategory PATCH /admin/expert-categories/:id
// All fields optional. Only provided fields are updated.
func (h *AdminHandler) UpdateExpertCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid category ID")
		return
	}
	var req struct {
		Name                   *string              `json:"name"`
		Description            *string              `json:"description"`
		TemplateSchema         *templateSchemaInput `json:"template_schema"`
		DefaultLanguage        *string              `json:"default_language"`
		AskStructurePermission *bool                `json:"ask_structure_permission"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if req.TemplateSchema != nil {
		if err := validateTemplateSchema(*req.TemplateSchema); err != nil {
			response.BadRequest(c, "INVALID_TEMPLATE_SCHEMA", err.Error())
			return
		}
	}

	ctx := c.Request.Context()

	if req.Name != nil {
		if _, err := h.db.Exec(ctx,
			`UPDATE expert_categories SET name=$1, updated_at=NOW() WHERE id=$2`,
			*req.Name, id); err != nil {
			h.logger.Error("update category name failed", zap.Error(err))
		}
	}
	if req.Description != nil {
		if _, err := h.db.Exec(ctx,
			`UPDATE expert_categories SET description=$1, updated_at=NOW() WHERE id=$2`,
			*req.Description, id); err != nil {
			h.logger.Error("update category description failed", zap.Error(err))
		}
	}
	if req.TemplateSchema != nil {
		schemaJSON, err := json.Marshal(*req.TemplateSchema)
		if err != nil {
			response.BadRequest(c, "INVALID_TEMPLATE_SCHEMA", "template_schema must be valid JSON")
			return
		}
		if _, err := h.db.Exec(ctx,
			`UPDATE expert_categories SET template_schema=$1, updated_at=NOW() WHERE id=$2`,
			string(schemaJSON), id); err != nil {
			h.logger.Error("update category template_schema failed", zap.Error(err))
		}
	}
	if req.DefaultLanguage != nil {
		if _, err := h.db.Exec(ctx,
			`UPDATE expert_categories SET default_language=$1, updated_at=NOW() WHERE id=$2`,
			*req.DefaultLanguage, id); err != nil {
			h.logger.Error("update category default_language failed", zap.Error(err))
		}
	}
	if req.AskStructurePermission != nil {
		if _, err := h.db.Exec(ctx,
			`UPDATE expert_categories SET ask_structure_permission=$1, updated_at=NOW() WHERE id=$2`,
			*req.AskStructurePermission, id); err != nil {
			h.logger.Error("update category ask_structure_permission failed", zap.Error(err))
		}
	}

	if err := h.categoryReg.Reload(ctx); err != nil {
		h.logger.Warn("category registry reload after update failed", zap.Error(err))
	}

	response.OK(c, map[string]string{"status": "updated"})
}

// ============================================================
// DOMAIN PROFILES (China Wall per-domain config, admin-configurable)
// ============================================================
//
// WHY this exists (2026-09-08): MaxTokensFlat/MaxTokensStructured were
// added to chinawall.DomainProfile so a truncated-JSON incident for one
// domain (see HANDOFF.md's 2026-09-08 round-3 RCA) could be tuned without
// a backend redeploy. Scoped to the WHOLE DomainProfile (not just the two
// token fields) per admin decision: every field on this struct is already
// admin-panel-editable in spirit (DomainRegistry.Upsert exists precisely
// for admin overrides) but had no HTTP surface yet — this closes that gap
// once, generically, instead of bolting on a narrow max-tokens-only route
// now and a second broader route later.

// domainProfileRow is the wire shape for a chinawall.DomainProfile.
// Mirrors the struct field-for-field (domain_profile.go) rather than
// reusing chinawall.DomainProfile directly as the JSON contract — same
// separation-of-concerns rationale as categoryRow/adminExpertRow above:
// the admin HTTP contract is explicit and stable even if the internal
// struct's Go-side shape changes.
type domainProfileRow struct {
	Domain              string                    `json:"domain"`
	Gate1Skip           bool                      `json:"gate1_skip"`
	CoverageMode        chinawall.CoverageMode    `json:"coverage_mode"`
	CitationMode        chinawall.CitationMode    `json:"citation_mode"`
	StripMode           chinawall.StripMode       `json:"strip_mode"`
	SystemPromptExt     string                    `json:"system_prompt_ext"`
	DomainKeywords      []string                  `json:"domain_keywords"`
	CustomRules         []chinawall.DomainRule    `json:"custom_rules"`
	MaxTokensFlat       int                       `json:"max_tokens_flat"`
	MaxTokensStructured int                       `json:"max_tokens_structured"`
}

func toDomainProfileRow(p *chinawall.DomainProfile) domainProfileRow {
	row := domainProfileRow{
		Domain:              p.Domain,
		Gate1Skip:           p.Gate1Skip,
		CoverageMode:        p.CoverageMode,
		CitationMode:        p.CitationMode,
		StripMode:           p.StripMode,
		SystemPromptExt:     p.SystemPromptExt,
		DomainKeywords:      p.DomainKeywords,
		CustomRules:         p.CustomRules,
		MaxTokensFlat:       p.MaxTokensFlat,
		MaxTokensStructured: p.MaxTokensStructured,
	}
	if row.DomainKeywords == nil {
		row.DomainKeywords = []string{}
	}
	if row.CustomRules == nil {
		row.CustomRules = []chinawall.DomainRule{}
	}
	return row
}

// validCoverageModes / validCitationModes / validStripModes mirror the
// CoverageMode/CitationMode/StripMode enums in domain_profile.go.
// Validated here (400 on bad input) rather than left to fail silently
// downstream — an invalid mode would not error anywhere else, it would
// just make the China Wall behave unpredictably for that domain.
var validCoverageModes = map[string]bool{
	string(chinawall.CoverageModeApplyPrinciples): true,
	string(chinawall.CoverageModeLiteralMatch):    true,
}
var validCitationModes = map[string]bool{
	string(chinawall.CitationModeLoose):  true,
	string(chinawall.CitationModeStrict): true,
}
var validStripModes = map[string]bool{
	string(chinawall.StripModeCodeExempt): true,
	string(chinawall.StripModeFull):       true,
}

// ListDomainProfiles GET /admin/domain-profiles
// Returns every domain profile currently cached in the registry
// (DB-backed, DomainRegistry.List — not BaseProfile, which is the
// unknown-domain fallback, not a stored/editable row).
func (h *AdminHandler) ListDomainProfiles(c *gin.Context) {
	profiles := h.domainReg.List()
	rows := make([]domainProfileRow, 0, len(profiles))
	for _, p := range profiles {
		rows = append(rows, toDomainProfileRow(p))
	}
	response.OK(c, rows)
}

// GetDomainProfile GET /admin/domain-profiles/:domain
func (h *AdminHandler) GetDomainProfile(c *gin.Context) {
	domain := c.Param("domain")
	profile := h.domainReg.GetExact(domain)
	if profile == nil {
		response.NotFound(c, "domain profile")
		return
	}
	response.OK(c, toDomainProfileRow(profile))
}

// UpdateDomainProfile PATCH /admin/domain-profiles/:domain
// All fields optional — only provided fields are changed. Starts from
// the domain's EXISTING profile if one is cached (GetExact), or from
// chinawall.BaseProfile's defaults if this domain has no stored profile
// yet (first-time creation via PATCH, upsert semantics — matches
// DomainRegistry.Upsert's own doc comment: "Admin panel when updating
// domain config"). :domain in the URL always wins over any "domain"
// field in the body, so the path is the single source of truth for
// which row is being written — the body cannot be used to silently
// retarget a different domain's row.
func (h *AdminHandler) UpdateDomainProfile(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		response.BadRequest(c, "INVALID_DOMAIN", "domain is required")
		return
	}

	var req struct {
		Gate1Skip           *bool     `json:"gate1_skip"`
		CoverageMode        *string   `json:"coverage_mode"`
		CitationMode        *string   `json:"citation_mode"`
		StripMode           *string   `json:"strip_mode"`
		SystemPromptExt     *string   `json:"system_prompt_ext"`
		DomainKeywords      *[]string `json:"domain_keywords"`
		MaxTokensFlat       *int      `json:"max_tokens_flat"`
		MaxTokensStructured *int      `json:"max_tokens_structured"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	if req.CoverageMode != nil && !validCoverageModes[*req.CoverageMode] {
		response.BadRequest(c, "INVALID_COVERAGE_MODE", "coverage_mode must be APPLY_PRINCIPLES or LITERAL_MATCH")
		return
	}
	if req.CitationMode != nil && !validCitationModes[*req.CitationMode] {
		response.BadRequest(c, "INVALID_CITATION_MODE", "citation_mode must be LOOSE or STRICT")
		return
	}
	if req.StripMode != nil && !validStripModes[*req.StripMode] {
		response.BadRequest(c, "INVALID_STRIP_MODE", "strip_mode must be CODE_EXEMPT or FULL_STRIP")
		return
	}
	// 0 is the documented "use default" sentinel (DefaultMaxTokensFlat/
	// DefaultMaxTokensStructured) — negative values are never valid.
	if req.MaxTokensFlat != nil && *req.MaxTokensFlat < 0 {
		response.BadRequest(c, "INVALID_MAX_TOKENS_FLAT", "max_tokens_flat must be >= 0 (0 = use default)")
		return
	}
	if req.MaxTokensStructured != nil && *req.MaxTokensStructured < 0 {
		response.BadRequest(c, "INVALID_MAX_TOKENS_STRUCTURED", "max_tokens_structured must be >= 0 (0 = use default)")
		return
	}

	// Start from the existing stored profile, or BaseProfile's defaults
	// if this domain has never been saved before. Copy the struct value
	// (not the pointer) so mutating fields below never corrupts the
	// registry's live cache before Upsert — Upsert is what commits the
	// change, not this local copy.
	existing := h.domainReg.GetExact(domain)
	var profile chinawall.DomainProfile
	if existing != nil {
		profile = *existing
	} else {
		profile = *chinawall.BaseProfile
	}
	profile.Domain = domain

	if req.Gate1Skip != nil {
		profile.Gate1Skip = *req.Gate1Skip
	}
	if req.CoverageMode != nil {
		profile.CoverageMode = chinawall.CoverageMode(*req.CoverageMode)
	}
	if req.CitationMode != nil {
		profile.CitationMode = chinawall.CitationMode(*req.CitationMode)
	}
	if req.StripMode != nil {
		profile.StripMode = chinawall.StripMode(*req.StripMode)
	}
	if req.SystemPromptExt != nil {
		profile.SystemPromptExt = *req.SystemPromptExt
	}
	if req.DomainKeywords != nil {
		profile.DomainKeywords = *req.DomainKeywords
	}
	if req.MaxTokensFlat != nil {
		profile.MaxTokensFlat = *req.MaxTokensFlat
	}
	if req.MaxTokensStructured != nil {
		profile.MaxTokensStructured = *req.MaxTokensStructured
	}

	if err := h.domainReg.Upsert(c.Request.Context(), &profile); err != nil {
		h.logger.Error("update domain profile failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.OK(c, toDomainProfileRow(&profile))
}

// IngestTranscript POST /admin/experts/:id/ingest
// Accepts multipart form with "transcript" file.
// Starts background ingestion job.
func (h *AdminHandler) IngestTranscript(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}

	// Get expert name
	var expertName string
	h.db.QueryRow(c.Request.Context(),
		`SELECT name FROM experts WHERE id=$1 AND deleted_at IS NULL`, expertID,
	).Scan(&expertName)
	if expertName == "" {
		response.NotFound(c, "expert")
		return
	}

	// Parse file
	file, header, err := c.Request.FormFile("transcript")
	if err != nil {
		response.BadRequest(c, "FILE_REQUIRED", "transcript file is required")
		return
	}
	defer file.Close()

	// Validate
	if header.Size > 50*1024*1024 {
		response.BadRequest(c, "FILE_TOO_LARGE", "file must be under 50MB")
		return
	}

	// D3: reject unsupported formats BEFORE creating a job.
	// WHY before: an unusable upload must not leave an orphan job row in the
	// ingestion list, and the admin gets the supported list straight away
	// instead of having to diagnose a failed job.
	if !docextract.IsSupported(header.Filename) {
		response.BadRequest(c, "UNSUPPORTED_FORMAT",
			fmt.Sprintf(".%s is not a supported format. Supported: %s",
				docextract.Extension(header.Filename), docextract.SupportedList()))
		return
	}

	content := make([]byte, header.Size)
	if _, err := io.ReadFull(file, content); err != nil {
		response.InternalError(c)
		return
	}

	// Create ingestion job
	var jobID uuid.UUID
	err = h.db.QueryRow(c.Request.Context(),
		`INSERT INTO ingestion_jobs (expert_id, job_type, status, source_path)
		 VALUES ($1, 'transcript', 'pending', $2)
		 RETURNING id`,
		expertID, header.Filename,
	).Scan(&jobID)
	if err != nil {
		h.logger.Error("create ingestion job failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	// Store transcript content for resume support (migration 009)
	// WHY: If job fails during chunking, resume needs the original text.
	// Non-fatal if this fails — resume will require re-upload in that case.
	//
	// NOTE (D3): for non-text uploads the stored transcript is now the EXTRACTED
	// text, written by PrepareTranscript in the goroutine below (it cannot be
	// written here — the bytes are a PDF here, not text). For .txt/.md the
	// extraction is a decode, so this column ends up holding the same content
	// one step later than before.

	// replace_existing=true replaces the expert's whole corpus instead of adding
	// to it. Read from the multipart form (the file upload already parsed it).
	// A malformed value is treated as false — the safe default.
	replaceExisting := false
	if v := strings.TrimSpace(c.PostForm("replace_existing")); v != "" {
		replaceExisting, _ = strconv.ParseBool(v)
	}

	// Mark expert as training
	_, _ = h.db.Exec(c.Request.Context(),
		`UPDATE experts SET is_training=TRUE, updated_at=NOW() WHERE id=$1`, expertID)

	// Start background ingestion
	// WHY goroutine: Ingestion takes minutes. Client gets job ID immediately.
	// replaceExisting is false (append) unless the request asks to replace.
	//
	// WHY the choice has to be reachable now: append mode dedups on chunk_hash, so
	// re-ingesting through a CHANGED chunker does not replace anything — it adds a
	// second copy of the same course with different hashes, and the corpus gets
	// worse (duplicate retrieval hits) while looking like a successful re-ingest.
	// Replacing is the only way to move an expert onto new chunking, so it is a
	// request field: replace_existing=true deletes the expert's chunks before
	// storing, and the admin screen warns before using it.
	// The default stays append: DOMAIN_EXPERT_COLLABORATION_DESIGN.md §5.4.
	// Admin uploading a new transcript adds to the corpus; full-retrain (true) is a
	// separate explicit operation reserved for Phase B.
	//
	// WHY 2-hour timeout (not context.Background()):
	//   context.Background() has NO deadline. If any step hangs (LLM timeout,
	//   DB deadlock, network partition), the goroutine hangs forever.
	//   Observed bug: CodeCraft LLM hung, HTTP timeout (120s) fired, but
	//   pauseOnLLMFailure() DB update blocked forever on a slow connection.
	//   Job stayed 'running' for 30+ minutes with no way to recover.
	//   2-hour timeout is generous (covers large transcripts with 60min embedding)
	//   but prevents infinite hangs. If timeout fires, job fails with clear error.
	//   Checkpoint system preserves progress; admin can retry from last checkpoint.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer cancel() // prevent context leak

		// Stage 0 (D3): convert the upload to text. PDFs, office documents and
		// spreadsheets are parsed by the ML sidecar; .txt/.md are decoded here.
		// WHY inside the goroutine: a 40MB PDF can take seconds, and the admin
		// should get the job id immediately. On failure the job is already
		// marked 'failed' with an admin-readable reason on the timeline, so
		// there is nothing left to do but log and stop.
		transcript, prepErr := h.ingestion.PrepareTranscript(
			ctx, jobID, expertID, header.Filename, content,
		)
		if prepErr != nil {
			h.logger.Warn("ingestion: transcript preparation failed",
				zap.String("job_id", jobID.String()),
				zap.String("expert_id", expertID.String()),
				zap.String("filename", header.Filename),
				zap.Error(prepErr),
			)
			return
		}

		_, err := h.ingestion.IngestTranscript(
			ctx,
			jobID, expertID, expertName,
			transcript,
			header.Filename,
			replaceExisting, // from the request; false = append (the default)
		)
		if err != nil {
			if errors.Is(err, training.ErrJobPaused) {
				// Not a crash — pipeline paused intentionally at charter
				// extraction. Job status is already 'paused' in DB.
				// SSE stream will emit llm_failure_decision_required event.
				// Admin must click Retry Now or leave paused (auto-fails 24h).
				h.logger.Info("ingestion paused: charter LLM failure — awaiting admin action",
					zap.String("job_id", jobID.String()),
					zap.String("expert_id", expertID.String()),
				)
				return
			}
			// Check if timeout fired (context deadline exceeded)
			if errors.Is(err, context.DeadlineExceeded) {
				h.logger.Error("ingestion timeout: exceeded 2-hour deadline",
					zap.String("job_id", jobID.String()),
					zap.String("expert_id", expertID.String()),
					zap.Error(err),
				)
				// Job should be marked 'failed' by ingestion pipeline.
				// If not, the 24h auto-fail checker will catch it.
				return
			}
			h.logger.Error("ingestion failed",
				zap.String("job_id", jobID.String()),
				zap.Error(err),
			)
		}
		// C2: on success, snapshot the new corpus/charter state and detect
		// drift vs the prior active version. Best-effort — a versioning
		// failure must never undo a successful ingestion.
		if err == nil && h.versions != nil {
			if _, sErr := h.versions.Snapshot(ctx, expertID, "ingest", header.Filename); sErr != nil {
				h.logger.Warn("expert version snapshot failed",
					zap.String("expert_id", expertID.String()),
					zap.Error(sErr),
				)
			}
		}
		// C6: on success, refresh the expert's freshness/refresh tasks so a
		// fresh ingest clears stale/re-embed signals. Best-effort.
		if err == nil && h.freshness != nil && h.freshness.Enabled() {
			if _, fErr := h.freshness.ScanExpert(ctx, expertID); fErr != nil {
				h.logger.Warn("freshness scan after ingest failed",
					zap.String("expert_id", expertID.String()),
					zap.Error(fErr),
				)
			}
		}
	}()

	response.Created(c, map[string]interface{}{
		"job_id":    jobID,
		"status":    "pending",
		"message":   "ingestion started in background",
		"expert_id": expertID,
	})
}

// ============================================================
// EXPERT VERSIONING & DRIFT (C2)
// ============================================================

// ListExpertVersions GET /admin/experts/:id/versions
func (h *AdminHandler) ListExpertVersions(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	if h.versions == nil {
		response.OK(c, []expertversion.Version{})
		return
	}
	versions, err := h.versions.ListVersions(c.Request.Context(), expertID, 20)
	if err != nil {
		h.logger.Error("list expert versions failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, versions)
}

// SnapshotExpertVersion POST /admin/experts/:id/versions/snapshot
// Body (optional): {"source": "manual", "notes": "..."}
func (h *AdminHandler) SnapshotExpertVersion(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	if h.versions == nil {
		response.InternalError(c)
		return
	}
	var body struct {
		Source string `json:"source"`
		Notes  string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&body) // body optional
	if body.Source == "" {
		body.Source = "manual"
	}
	v, err := h.versions.Snapshot(c.Request.Context(), expertID, body.Source, body.Notes)
	if err != nil {
		h.logger.Error("snapshot expert version failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	if v == nil {
		response.NotFound(c, "expert")
		return
	}
	response.Created(c, v)
}

// PinExpertVersion POST /admin/experts/:id/versions/:versionId/pin
// Marks the version canonical and rolls the expert's charter back to it.
func (h *AdminHandler) PinExpertVersion(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	versionID, err := uuid.Parse(c.Param("versionId"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid version ID")
		return
	}
	if h.versions == nil {
		response.InternalError(c)
		return
	}
	v, err := h.versions.Pin(c.Request.Context(), expertID, versionID)
	if err != nil {
		h.logger.Error("pin expert version failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	if v == nil {
		response.NotFound(c, "expert version")
		return
	}
	response.OK(c, v)
}

// ListExpertDrift GET /admin/experts/:id/drift
func (h *AdminHandler) ListExpertDrift(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	if h.versions == nil {
		response.OK(c, []expertversion.DriftEvent{})
		return
	}
	events, err := h.versions.ListDrift(c.Request.Context(), expertID, 20)
	if err != nil {
		h.logger.Error("list expert drift failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, events)
}

// AcknowledgeExpertDrift POST /admin/experts/:id/drift/:driftId/ack
func (h *AdminHandler) AcknowledgeExpertDrift(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	driftID, err := uuid.Parse(c.Param("driftId"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid drift ID")
		return
	}
	if h.versions == nil {
		response.InternalError(c)
		return
	}
	if err := h.versions.AcknowledgeDrift(c.Request.Context(), expertID, driftID); err != nil {
		h.logger.Error("acknowledge drift failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"acknowledged": true})
}

// ============================================================
// EVAL HARNESS (C3)
// ============================================================

// evalSuiteView is the admin payload for one suite: latest run, baseline,
// and the score delta (current - baseline). Pure assembly; scoring lives
// in internal/eval.
type evalSuiteView struct {
	Suite    string           `json:"suite"`
	Latest   *eval.RunSummary `json:"latest,omitempty"`
	Baseline *eval.RunSummary `json:"baseline,omitempty"`
	Delta    float64          `json:"delta"`
	Suites   []string         `json:"suites,omitempty"`
}

// GetEvalRuns GET /admin/evals/runs?suite=<name>
// Without suite: lists known suites. With suite: latest + baseline + delta.
func (h *AdminHandler) GetEvalRuns(c *gin.Context) {
	if h.evals == nil {
		response.OK(c, evalSuiteView{Suites: []string{}})
		return
	}
	suite := strings.TrimSpace(c.Query("suite"))
	if suite == "" {
		names, err := h.evals.ListSuites(c.Request.Context())
		if err != nil {
			h.logger.Error("list eval suites failed", zap.Error(err))
			response.InternalError(c)
			return
		}
		response.OK(c, evalSuiteView{Suites: names})
		return
	}
	latest, err := h.evals.LatestRun(c.Request.Context(), suite)
	if err != nil {
		h.logger.Error("latest eval run failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	baseline, err := h.evals.LatestBaseline(c.Request.Context(), suite)
	if err != nil {
		h.logger.Error("latest eval baseline failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, evalSuiteView{
		Suite:    suite,
		Latest:   latest,
		Baseline: baseline,
		Delta:    eval.ScoreDelta(baseline, latest),
	})
}

// PromoteEvalBaseline POST /admin/evals/runs/:id/baseline
// Body: {"suite":"chat"} — promotes the given run id as the suite baseline.
func (h *AdminHandler) PromoteEvalBaseline(c *gin.Context) {
	if h.evals == nil {
		response.InternalError(c)
		return
	}
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid run ID")
		return
	}
	var body struct {
		Suite string `json:"suite"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Suite) == "" {
		response.BadRequest(c, "INVALID_INPUT", "suite is required")
		return
	}
	if err := h.evals.PromoteBaseline(c.Request.Context(), body.Suite, runID); err != nil {
		h.logger.Error("promote eval baseline failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"suite": body.Suite, "run_id": runID, "is_baseline": true})
}

// ============================================================
// TENANT ISOLATION & ENTERPRISE CONTROLS (C4)
// ============================================================

// ListTenants GET /admin/tenants
func (h *AdminHandler) ListTenants(c *gin.Context) {
	if h.tenants == nil {
		response.OK(c, []tenant.Tenant{})
		return
	}
	tenants, err := h.tenants.List(c.Request.Context())
	if err != nil {
		h.logger.Error("list tenants failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, tenants)
}

// CreateTenant POST /admin/tenants
// Body: {"name":"Acme Corp","slug":"acme"} (slug optional → derived from name)
func (h *AdminHandler) CreateTenant(c *gin.Context) {
	if h.tenants == nil {
		response.InternalError(c)
		return
	}
	var body struct {
		Name string `json:"name" binding:"required"`
		Slug string `json:"slug"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	t, err := h.tenants.Create(c.Request.Context(), body.Name, body.Slug)
	if err != nil {
		h.logger.Warn("create tenant failed", zap.Error(err))
		response.BadRequest(c, "CREATE_FAILED", err.Error())
		return
	}
	response.Created(c, t)
}

// AssignTenantUser POST /admin/tenants/:id/users
// Body: {"user_id":"<uuid>"} — moves the account into the tenant.
func (h *AdminHandler) AssignTenantUser(c *gin.Context) {
	if h.tenants == nil {
		response.InternalError(c)
		return
	}
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid tenant ID")
		return
	}
	var body struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	userID, err := uuid.Parse(body.UserID)
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid user ID")
		return
	}
	if err := h.tenants.SetUserTenant(c.Request.Context(), userID, tenantID); err != nil {
		h.logger.Warn("assign tenant user failed", zap.Error(err))
		response.BadRequest(c, "ASSIGN_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"user_id": userID, "tenant_id": tenantID})
}

// AssignTenantExpert POST /admin/tenants/:id/experts
// Body: {"expert_id":"<uuid>","global":false}
// Homes an expert into this tenant (tenant-private), or back to platform
// (global=true → tenant_id NULL, visible to every tenant).
func (h *AdminHandler) AssignTenantExpert(c *gin.Context) {
	if h.tenants == nil {
		response.InternalError(c)
		return
	}
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid tenant ID")
		return
	}
	var body struct {
		ExpertID string `json:"expert_id" binding:"required"`
		Global   bool   `json:"global"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	expertID, err := uuid.Parse(body.ExpertID)
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	var target *uuid.UUID
	if !body.Global {
		target = &tenantID
	}
	if err := h.tenants.SetExpertTenant(c.Request.Context(), expertID, target); err != nil {
		h.logger.Warn("assign tenant expert failed", zap.Error(err))
		response.BadRequest(c, "ASSIGN_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"expert_id": expertID, "tenant_id": target, "global": body.Global})
}

// ============================================================
// COST & USAGE ANALYTICS (C5)
// ============================================================

// GetUsage GET /admin/usage?group_by=tenant|project|expert|model|use_case
//              &from=RFC3339&to=RFC3339&tenant_id=&project_id=&expert_id=&use_case=
func (h *AdminHandler) GetUsage(c *gin.Context) {
	if h.usage == nil {
		response.OK(c, gin.H{"group_by": "tenant", "rows": []usage.Row{}})
		return
	}
	f := usage.Filter{
		GroupBy: c.Query("group_by"),
		UseCase: strings.TrimSpace(c.Query("use_case")),
	}
	if v := c.Query("from"); v != "" {
		if ts, err := time.Parse(time.RFC3339, v); err == nil {
			f.From = ts
		} else {
			response.BadRequest(c, "INVALID_INPUT", "from must be RFC3339")
			return
		}
	}
	if v := c.Query("to"); v != "" {
		if ts, err := time.Parse(time.RFC3339, v); err == nil {
			f.To = ts
		} else {
			response.BadRequest(c, "INVALID_INPUT", "to must be RFC3339")
			return
		}
	}
	var err error
	if f.TenantID, err = parseOptionalUUID(c.Query("tenant_id")); err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid tenant_id")
		return
	}
	if f.ProjectID, err = parseOptionalUUID(c.Query("project_id")); err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid project_id")
		return
	}
	if f.ExpertID, err = parseOptionalUUID(c.Query("expert_id")); err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert_id")
		return
	}
	rows, err := h.usage.Summary(c.Request.Context(), f)
	if err != nil {
		h.logger.Error("usage summary failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"group_by": f.GroupBy, "rows": rows})
}

// GetUsageBudgets GET /admin/usage/budgets
// Returns the global default budget + every per-tenant budget, resolved
// against this month's spend.
func (h *AdminHandler) GetUsageBudgets(c *gin.Context) {
	if h.usage == nil {
		response.OK(c, gin.H{"global": nil, "tenants": []usage.BudgetStatus{}})
		return
	}
	ctx := c.Request.Context()
	global, err := h.usage.BudgetStatus(ctx, nil)
	if err != nil {
		h.logger.Error("usage global budget failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	alerts, err := h.usage.Alerts(ctx)
	if err != nil {
		h.logger.Error("usage budgets failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"global": global, "tenants": alerts})
}

// SetUsageBudget PUT /admin/usage/budgets
// Body: {"tenant_id":"<uuid>"?, "monthly_limit_usd":100, "alert_threshold":0.8}
// Omit tenant_id to set the platform default.
func (h *AdminHandler) SetUsageBudget(c *gin.Context) {
	if h.usage == nil {
		response.InternalError(c)
		return
	}
	var body struct {
		TenantID        string  `json:"tenant_id"`
		MonthlyLimitUSD float64 `json:"monthly_limit_usd"`
		AlertThreshold  float64 `json:"alert_threshold"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	tenantID, err := parseOptionalUUID(body.TenantID)
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid tenant_id")
		return
	}
	if body.AlertThreshold == 0 {
		body.AlertThreshold = 0.8
	}
	if err := h.usage.SetBudget(c.Request.Context(), tenantID, body.MonthlyLimitUSD, body.AlertThreshold); err != nil {
		h.logger.Warn("set usage budget failed", zap.Error(err))
		response.BadRequest(c, "SET_FAILED", err.Error())
		return
	}
	st, err := h.usage.BudgetStatus(c.Request.Context(), tenantID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, st)
}

// GetUsageAlerts GET /admin/usage/alerts
// Budgets whose spend has reached the alert threshold or the limit.
func (h *AdminHandler) GetUsageAlerts(c *gin.Context) {
	if h.usage == nil {
		response.OK(c, []usage.BudgetStatus{})
		return
	}
	alerts, err := h.usage.Alerts(c.Request.Context())
	if err != nil {
		h.logger.Error("usage alerts failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, alerts)
}

// parseOptionalUUID returns nil for an empty string, a parsed UUID otherwise.
func parseOptionalUUID(s string) (*uuid.UUID, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// ============================================================
// KNOWLEDGE FRESHNESS (C6)
// ============================================================

// GetExpertFreshness GET /admin/experts/:id/freshness
// Read-only expert freshness summary (no tasks written).
func (h *AdminHandler) GetExpertFreshness(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	if h.freshness == nil || !h.freshness.Enabled() {
		response.OK(c, map[string]interface{}{"expert_id": expertID, "freshness_enabled": false})
		return
	}
	ef, err := h.freshness.GetExpert(c.Request.Context(), expertID)
	if err != nil {
		h.logger.Error("get expert freshness failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	tasks, err := h.freshness.ListTasks(c.Request.Context(), "", &expertID, 100)
	if err != nil {
		h.logger.Warn("list expert freshness tasks failed", zap.Error(err))
		tasks = nil
	}
	response.OK(c, gin.H{
		"freshness_enabled": true,
		"max_age_days":      h.freshness.MaxAgeDays(),
		"status":            ef,
		"tasks":             tasks,
	})
}

// ScanExpertFreshness POST /admin/experts/:id/freshness/scan
// Recomputes signals and upserts refresh tasks for this expert.
func (h *AdminHandler) ScanExpertFreshness(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	if h.freshness == nil || !h.freshness.Enabled() {
		response.ServiceUnavailable(c, "knowledge freshness is disabled")
		return
	}
	ef, err := h.freshness.ScanExpert(c.Request.Context(), expertID)
	if err != nil {
		h.logger.Error("scan expert freshness failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, ef)
}

// ScanAllFreshness POST /admin/freshness/scan
// Scans every expert. Returns the per-expert summaries.
func (h *AdminHandler) ScanAllFreshness(c *gin.Context) {
	if h.freshness == nil || !h.freshness.Enabled() {
		response.ServiceUnavailable(c, "knowledge freshness is disabled")
		return
	}
	results, err := h.freshness.ScanAll(c.Request.Context())
	if err != nil {
		h.logger.Error("scan all freshness failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"scanned": len(results), "experts": results})
}

// ListFreshnessTasks GET /admin/freshness/tasks?status=&expert_id=
func (h *AdminHandler) ListFreshnessTasks(c *gin.Context) {
	if h.freshness == nil || !h.freshness.Enabled() {
		response.OK(c, []knowledge.Task{})
		return
	}
	expertID, err := parseOptionalUUID(c.Query("expert_id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert_id")
		return
	}
	tasks, err := h.freshness.ListTasks(c.Request.Context(),
		strings.TrimSpace(c.Query("status")), expertID, 200)
	if err != nil {
		h.logger.Error("list freshness tasks failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, tasks)
}

// AcknowledgeFreshnessTask POST /admin/freshness/tasks/:taskId/ack
func (h *AdminHandler) AcknowledgeFreshnessTask(c *gin.Context) {
	if h.freshness == nil || !h.freshness.Enabled() {
		response.ServiceUnavailable(c, "knowledge freshness is disabled")
		return
	}
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid task ID")
		return
	}
	if err := h.freshness.AcknowledgeTask(c.Request.Context(), taskID); err != nil {
		h.logger.Warn("acknowledge freshness task failed", zap.Error(err))
		response.BadRequest(c, "ACK_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"task_id": taskID, "status": knowledge.StatusAcknowledged})
}

// ResolveFreshnessTask POST /admin/freshness/tasks/:taskId/resolve
func (h *AdminHandler) ResolveFreshnessTask(c *gin.Context) {
	if h.freshness == nil || !h.freshness.Enabled() {
		response.ServiceUnavailable(c, "knowledge freshness is disabled")
		return
	}
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid task ID")
		return
	}
	if err := h.freshness.ResolveTask(c.Request.Context(), taskID); err != nil {
		h.logger.Warn("resolve freshness task failed", zap.Error(err))
		response.BadRequest(c, "RESOLVE_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"task_id": taskID, "status": knowledge.StatusResolved})
}

// ============================================================
// BRING-YOUR-OWN EXPERT (C8)
// ============================================================

// SetTenantEntitlement POST /admin/tenants/:id/entitlement
// Body: {"allow_byo_expert":true,"max_experts":5}
// Grants/revokes tenant self-service experts (C8). Merges into tenants.settings.
func (h *AdminHandler) SetTenantEntitlement(c *gin.Context) {
	if h.byo == nil {
		response.ServiceUnavailable(c, "byo experts are not configured")
		return
	}
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid tenant ID")
		return
	}
	var body struct {
		AllowByoExpert bool `json:"allow_byo_expert"`
		MaxExperts     int  `json:"max_experts"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if body.MaxExperts < 0 {
		response.BadRequest(c, "INVALID_INPUT", "max_experts must be >= 0")
		return
	}
	if err := h.byo.SetEntitlement(c.Request.Context(), tenantID, body.AllowByoExpert, body.MaxExperts); err != nil {
		h.logger.Warn("set tenant entitlement failed", zap.Error(err))
		response.BadRequest(c, "SET_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"tenant_id": tenantID, "allow_byo_expert": body.AllowByoExpert, "max_experts": body.MaxExperts})
}

// ListByoEvents GET /admin/byo/events?tenant_id=&limit=
// Append-only audit of tenant BYO activity (register/ingest/denials).
func (h *AdminHandler) ListByoEvents(c *gin.Context) {
	if h.byo == nil {
		response.OK(c, []byoexpert.Event{})
		return
	}
	tenantID, err := parseOptionalUUID(c.Query("tenant_id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid tenant_id")
		return
	}
	limit := 100
	if l := strings.TrimSpace(c.Query("limit")); l != "" {
		if n, convErr := strconv.Atoi(l); convErr == nil {
			limit = n
		}
	}
	events, err := h.byo.ListEvents(c.Request.Context(), tenantID, limit)
	if err != nil {
		h.logger.Error("list byo events failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, events)
}

// ============================================================
// INGESTION TRANSPARENCY (T1)
// ============================================================

// ingestionJobSnapshot is the mutable "hot" row the UI renders (progress bar,
// stage cards, ETA, cost, resume button). The fine-grained "what happened"
// detail lives in ingestion_job_events and is streamed alongside it.
type ingestionJobSnapshot struct {
	ID                        uuid.UUID  `json:"id"`
	Status                    string     `json:"status"`
	SourcePath                string     `json:"source_path"`
	TotalChunks               int        `json:"total_chunks"`
	ProcessedChunks           int        `json:"processed_chunks"`
	ErrorMessage              string     `json:"error_message"`
	StartedAt                 *time.Time `json:"started_at"`
	CompletedAt               *time.Time `json:"completed_at"`
	CreatedAt                 time.Time  `json:"created_at"`
	CurrentStage              string     `json:"current_stage"`
	StageDetail               string     `json:"stage_detail"`
	CostUsd                   float64    `json:"cost_usd"`
	EstimatedSecondsRemaining *int       `json:"estimated_seconds_remaining"`
	ResumedFromCheckpoint     bool       `json:"resumed_from_checkpoint"`
}

// loadLatestJobSnapshot reads the most recent ingestion job for an expert.
func (h *AdminHandler) loadLatestJobSnapshot(ctx context.Context, expertID uuid.UUID) (*ingestionJobSnapshot, error) {
	var job ingestionJobSnapshot
	err := h.db.QueryRow(ctx, `
		SELECT id, status, COALESCE(source_path,''),
		       total_chunks, processed_chunks,
		       COALESCE(error_message,''), started_at, completed_at, created_at,
		       COALESCE(current_stage,'pending'), COALESCE(stage_detail,''),
		       COALESCE(cost_usd,0), estimated_seconds_remaining,
		       COALESCE(resumed_from_checkpoint,false)
		FROM ingestion_jobs
		WHERE expert_id=$1
		ORDER BY created_at DESC LIMIT 1`,
		expertID,
	).Scan(
		&job.ID, &job.Status, &job.SourcePath,
		&job.TotalChunks, &job.ProcessedChunks,
		&job.ErrorMessage, &job.StartedAt, &job.CompletedAt, &job.CreatedAt,
		&job.CurrentStage, &job.StageDetail,
		&job.CostUsd, &job.EstimatedSecondsRemaining,
		&job.ResumedFromCheckpoint,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// replayEvents pages the durable timeline from `since` and writes each row as an
// SSE `event` frame, advancing *since.
//
// WHY read the DB instead of trusting the live stream: events are append-only,
// so a viewer that refreshes, reconnects, or joins mid-run reconstructs the
// identical timeline. That is what makes the log survive a page reload — the
// previous implementation kept it in browser memory only.
func (h *AdminHandler) replayEvents(ctx context.Context, jobID uuid.UUID, since *int64, send func(string, interface{})) {
	if !h.events.Enabled() || jobID == uuid.Nil {
		return
	}
	for {
		events, err := h.events.GetSince(ctx, jobID, *since, 200)
		if err != nil {
			h.logger.Warn("ingestion stream: timeline replay failed",
				zap.String("job_id", jobID.String()), zap.Error(err))
			return
		}
		for _, ev := range events {
			send("event", ev)
			*since = ev.SequenceNumber
		}
		if len(events) < 200 {
			return
		}
	}
}

// StreamIngestionJob GET /admin/experts/:id/jobs/stream
//
// SSE endpoint with two cooperating feeds:
//  1. `event` frames — the durable timeline (ingestion_job_events), pushed as
//     soon as the pipeline writes it (Redis pub/sub is the wake-up; Postgres is
//     the source of truth). This is the real-time transparency feed.
//  2. `update` / `complete` / `failed` / `llm_failure_decision_required` frames
//     — the job row, re-read every 2s. WHY still needed: events carry history,
//     the snapshot carries current ETA/cost and reconciles anything that could
//     not be written as an event. (It is deliberately slower than the events so
//     the events, not the poll, are what the admin reads.)
//
// Closes on a terminal state after flushing the tail of the timeline, so the
// final events (verified / complete / paused) are never cut off by the close.
func (h *AdminHandler) StreamIngestionJob(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // disable nginx buffering

	ctx := c.Request.Context()

	sendEvent := func(eventType string, payload interface{}) {
		envelope := map[string]interface{}{
			"type": eventType,
			"ts":   time.Now().UTC().Format(time.RFC3339Nano),
		}
		// The frontend has read {type, job} since the first version — keep that
		// shape for snapshots, and add {type:"event", event:{...}} for timeline
		// rows so old clients simply ignore the new frames.
		if eventType == "event" {
			envelope["event"] = payload
		} else {
			envelope["job"] = payload
		}
		data, _ := json.Marshal(envelope)
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		c.Writer.Flush()
	}

	// Force headers out immediately (Go writes them only on first write) so the
	// browser leaves CONNECTING even when nothing happens for a while.
	fmt.Fprint(c.Writer, ": connected\n\n")
	c.Writer.Flush()

	var (
		lastSeq       int64
		eventCh       <-chan jobevents.Event
		subscribedFor uuid.UUID
	)

	// subscribe attaches a live timeline feed for a job (idempotent).
	subscribe := func(jobID uuid.UUID) {
		if !h.events.Enabled() || jobID == uuid.Nil || jobID == subscribedFor {
			return
		}
		subscribedFor = jobID
		lastSeq = 0
		h.replayEvents(ctx, jobID, &lastSeq, sendEvent)

		ch, errCh := h.events.Subscriber(h.logger).Subscribe(ctx, jobID, lastSeq)
		eventCh = ch
		// Drain the error channel: the subscriber reports non-fatal problems
		// there (initial catch-up failure) and keeps streaming. Not draining
		// it would be a silent drop; the timeline is what the admin trusts.
		go func() {
			for subErr := range errCh {
				if subErr != nil {
					h.logger.Warn("ingestion stream: subscriber error (non-fatal)",
						zap.String("job_id", jobID.String()),
						zap.Error(subErr),
					)
				}
			}
		}()
	}

	// A job can appear shortly AFTER the client connects (upload is async), so
	// a missing snapshot is not an error — the ticker picks it up below.
	if snapshot, snapErr := h.loadLatestJobSnapshot(ctx, expertID); snapErr == nil {
		sendEvent("hello", snapshot)
		subscribe(snapshot.ID)
	}

	ticker := time.NewTicker(2 * time.Second)
	heartbeat := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-heartbeat.C:
			// Keep the connection alive through proxies (nginx idle timeout).
			fmt.Fprintf(c.Writer, "data: {\"type\":\"heartbeat\",\"ts\":\"%s\"}\n\n",
				time.Now().UTC().Format(time.RFC3339Nano))
			c.Writer.Flush()

		case ev, ok := <-eventCh:
			if !ok {
				// Feed closed (context cancelled). Stop selecting on it rather
				// than spinning on a closed channel.
				eventCh = nil
				continue
			}
			sendEvent("event", ev)
			if ev.SequenceNumber > lastSeq {
				lastSeq = ev.SequenceNumber
			}

		case <-ticker.C:
			job, loadErr := h.loadLatestJobSnapshot(ctx, expertID)
			if loadErr != nil {
				fmt.Fprintf(c.Writer, "data: {\"type\":\"waiting\",\"ts\":\"%s\"}\n\n",
					time.Now().UTC().Format(time.RFC3339Nano))
				c.Writer.Flush()
				continue
			}
			// Late-arriving job: attach the timeline feed now.
			subscribe(job.ID)

			eventType := "update"
			terminal := false
			switch job.Status {
			case "complete":
				eventType, terminal = "complete", true
			case "complete_with_warnings":
				// The run finished but the corpus is not fully usable (storage
				// verification failed, or the smoke test did not pass). It is
				// terminal — the pipeline is done and will not write again — so
				// the stream must close, or the modal sits on "running" forever
				// waiting for events that will never come.
				eventType, terminal = "complete_with_warnings", true
			case "failed":
				eventType, terminal = "failed", true
			case "paused":
				// Pipeline stopped for admin action (charter LLM failure).
				// Frontend shows the Retry button on this event type.
				eventType, terminal = "llm_failure_decision_required", true
			}

			if terminal {
				// Flush the tail of the timeline BEFORE closing, otherwise the
				// final events race the close and can be lost.
				h.replayEvents(ctx, job.ID, &lastSeq, sendEvent)
			}
			sendEvent(eventType, job)

			if terminal {
				return
			}
		}
	}
}

// GetIngestionJobEvents GET /admin/experts/:id/jobs/events?jobId=<uuid>
//
// Returns the durable timeline for one job — used by the modal to render the
// "what happened" log when it opens on an already-finished job (the SSE stream
// closes at terminal state, so a completed run has no live feed).
//
// WHY jobId is a query param and not a path segment: the sibling route
// GET /experts/:id/jobs/stream already occupies that tree position, and mixing
// a static segment with a :jobID wildcard there is exactly the kind of route
// registration that panics at startup. A query param cannot conflict.
//
// Query params: jobId (required), after (sequence_number cursor, default 0),
// limit (default 200, max 1000). Response includes last_sequence so the client
// can continue without re-reading rows it already has.
func (h *AdminHandler) GetIngestionJobEvents(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	rawJobID := c.Query("jobId")
	if rawJobID == "" {
		rawJobID = c.Query("job_id")
	}
	jobID, err := uuid.Parse(rawJobID)
	if err != nil {
		response.BadRequest(c, "INVALID_JOB_ID", "valid jobId query parameter is required")
		return
	}

	if !h.events.Enabled() {
		response.OK(c, map[string]interface{}{
			"events":        []interface{}{},
			"last_sequence": 0,
			"timeline":      false,
		})
		return
	}

	ctx := c.Request.Context()

	// Ownership check: the job must belong to this expert. Prevents an admin
	// from reading another expert's timeline with a guessed job id.
	var ownerID uuid.UUID
	if err := h.db.QueryRow(ctx,
		`SELECT expert_id FROM ingestion_jobs WHERE id=$1`, jobID,
	).Scan(&ownerID); err != nil || ownerID != expertID {
		response.NotFound(c, "ingestion job")
		return
	}

	after := int64(0)
	if raw := c.Query("after"); raw != "" {
		if parsed, parseErr := strconv.ParseInt(raw, 10, 64); parseErr == nil && parsed >= 0 {
			after = parsed
		}
	}
	limit := 200
	if raw := c.Query("limit"); raw != "" {
		if parsed, parseErr := strconv.Atoi(raw); parseErr == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	events, err := h.events.GetSince(ctx, jobID, after, limit)
	if err != nil {
		h.logger.Error("get ingestion job events failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	if events == nil {
		events = []jobevents.Event{}
	}
	lastSeq := after
	if len(events) > 0 {
		lastSeq = events[len(events)-1].SequenceNumber
	}

	response.OK(c, map[string]interface{}{
		"events":        events,
		"last_sequence": lastSeq,
		"timeline":      true,
	})
}

// ============================================================
// INGESTION RECONCILE (Phase D)
// ============================================================

// GetIngestionAudit GET /admin/experts/:id/ingestion/audit
//
// Answers "is this expert's corpus what it claims to be?": the corpus read from
// course_chunks, the totals cached on the experts row, the capability table, and
// the per-run ledger. Read-only — every finding is reported, none is acted on.
func (h *AdminHandler) GetIngestionAudit(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}

	audit, err := h.reconcile.Audit(c.Request.Context(), expertID)
	if err != nil {
		h.logger.Error("ingestion audit failed",
			zap.String("expert_id", expertID.String()), zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, audit)
}

// GetIngestionDiagnostics GET /admin/experts/:id/ingestion/diagnostics?job=<uuid>
//
// Explains one job file by file. WHY a query param: the sibling route
// /experts/:id/jobs/:jobID/... already owns that tree position, and the ingestion
// tree is a separate static prefix — a query param cannot collide with either.
func (h *AdminHandler) GetIngestionDiagnostics(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	rawJobID := c.Query("job")
	if rawJobID == "" {
		rawJobID = c.Query("job_id")
	}
	jobID, err := uuid.Parse(rawJobID)
	if err != nil {
		response.BadRequest(c, "INVALID_JOB_ID", "valid job query parameter is required")
		return
	}

	diagnostics, err := h.reconcile.Diagnose(c.Request.Context(), expertID, jobID)
	if err != nil {
		h.logger.Error("ingestion diagnostics failed",
			zap.String("expert_id", expertID.String()),
			zap.String("job_id", jobID.String()),
			zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, diagnostics)
}

// reconcileRequestDTO is the wire shape. DryRun is a pointer so "absent" is
// distinguishable from "false": an omitted field must not authorise a write, so
// the default below is true and a caller has to say dry_run=false explicitly.
type reconcileRequestDTO struct {
	Actions    []string   `json:"actions"`
	DryRun     *bool      `json:"dry_run"`
	JobID      *uuid.UUID `json:"job_id"`
	SourceFile string     `json:"source_file"`
}

// ReconcileIngestion POST /admin/experts/:id/ingestion/reconcile
//
// Runs the requested repairs. Defaults to a dry run: it reports what it would
// change, and only writes when dry_run is explicitly false. Every action here
// writes derived state that is reconstructible from course_chunks, except
// resolve_run, which records an admin decision and touches no chunk.
//
// It cannot recreate chunks that were never stored — a short store is reported
// (see diagnostics), and the remedy is a re-ingest.
func (h *AdminHandler) ReconcileIngestion(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}

	var dto reconcileRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "INVALID_BODY", "expected {\"actions\": [...], \"dry_run\": true|false}")
		return
	}
	if err := training.ValidateActions(dto.Actions); err != nil {
		response.BadRequest(c, "INVALID_ACTION", err.Error())
		return
	}

	dryRun := true
	if dto.DryRun != nil {
		dryRun = *dto.DryRun
	}

	results, err := h.reconcile.Reconcile(c.Request.Context(), expertID, training.ReconcileRequest{
		Actions:    dto.Actions,
		DryRun:     dryRun,
		JobID:      dto.JobID,
		SourceFile: dto.SourceFile,
	})
	if err != nil {
		h.logger.Error("ingestion reconcile failed",
			zap.String("expert_id", expertID.String()), zap.Error(err))
		response.InternalError(c)
		return
	}

	// applied lets the UI keep the confirmation prompt up until the admin has
	// actually committed, instead of assuming the click wrote anything.
	applied := !dryRun
	response.OK(c, map[string]interface{}{
		"expert_id":         expertID,
		"dry_run":           dryRun,
		"applied":           applied,
		"results":           results,
		"available_actions": training.ReconcileActions(),
	})
}

// ============================================================
// CAPABILITY MEASUREMENT (I2)
// ============================================================

// capabilityEvalRequestDTO is the wire shape for one measurement pass.
type capabilityEvalRequestDTO struct {
	// Topics caps how many topics are measured. Zero = the backend default.
	Topics int `json:"topics"`
	// TopK is the retrieval depth scored against. Zero = the production default.
	TopK int `json:"top_k"`
	// Regenerate replaces the stored question set. Off by default because that
	// set IS the baseline: regenerating it silently would make two runs
	// incomparable while still looking like a fair comparison.
	Regenerate bool `json:"regenerate"`

	// GraphExpansion retrieves with concept-graph expansion. Off by default: the plain
	// path is what production uses until a measurement says otherwise, and a pass
	// records which path it used so the two are never compared against each other.
	GraphExpansion bool `json:"graph_expansion"`

	// LayerPreference retrieves with the ranking nudged toward the depth the
	// question was asked at. Off by default for the same reason: it is measured
	// before it is believed.
	LayerPreference bool `json:"layer_preference"`
}

// MeasureExpertCapability POST /admin/experts/:id/capability-eval
//
// Starts one measurement pass: write questions from the expert's own corpus, ask
// them through the production retrieval path, judge the answers, and record the
// evidence. Returns immediately with the run id; the pass costs one generation
// call per topic plus two calls per case, which is far longer than an HTTP request
// should live.
func (h *AdminHandler) MeasureExpertCapability(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	if h.capEval == nil {
		response.ServiceUnavailable(c, "capability measurement is not wired in this deployment")
		return
	}

	var dto capabilityEvalRequestDTO
	// An empty body is a valid "use the defaults" request, so a bind error is only
	// fatal when there was a body at all.
	if c.Request.ContentLength > 0 {
		if bindErr := c.ShouldBindJSON(&dto); bindErr != nil {
			response.BadRequest(c, "INVALID_BODY", "expected {\"topics\": 10, \"top_k\": 5, \"regenerate\": false}")
			return
		}
	}

	ctx := c.Request.Context()

	// The expert must exist: failing here returns a 404 the admin can act on,
	// whereas letting the background pass fail would report nothing until later.
	var exists bool
	if err := h.db.QueryRow(ctx,
		`SELECT TRUE FROM experts WHERE id = $1`, expertID).Scan(&exists); err != nil {
		response.NotFound(c, "expert")
		return
	}

	// One pass at a time. WHY: two concurrent passes would both write results and
	// both write back the measured capability, so the last writer would silently
	// win, and the cost would be paid twice.
	var running int
	if err := h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM expert_capability_eval_runs WHERE expert_id = $1 AND status = 'running'`,
		expertID).Scan(&running); err != nil {
		h.logger.Error("capability eval: could not check for a running pass", zap.Error(err))
		response.InternalError(c)
		return
	}
	if running > 0 {
		response.Conflict(c, "a capability measurement is already running for this expert")
		return
	}

	request := training.CapabilityEvalRequest{
		Topics:         dto.Topics,
		TopK:           dto.TopK,
		Regenerate:     dto.Regenerate,
		GraphExpansion: dto.GraphExpansion,
		LayerPreference: dto.LayerPreference,
	}

	// Detached context with a deadline: the pass must outlive this request, and a
	// pass that hangs must not hold the "running" gate forever.
	go func() {
		runCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		if _, err := h.capEval.RunEval(runCtx, expertID, request); err != nil {
			h.logger.Error("capability eval failed",
				zap.String("expert_id", expertID.String()),
				zap.Error(err))
		}
	}()

	response.OK(c, map[string]interface{}{
		"expert_id": expertID,
		"status":    "running",
		"topics":    request.Topics,
		"top_k":     request.TopK,
	})
}

// GetExpertCapabilityEval GET /admin/experts/:id/capability-eval
//
// Returns the most recent pass for an expert, or null when it has never been
// measured — "never measured" is a different state from "measured and empty", and
// the screen must be able to say which one it is showing.
func (h *AdminHandler) GetExpertCapabilityEval(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	if h.capEval == nil {
		response.ServiceUnavailable(c, "capability measurement is not wired in this deployment")
		return
	}

	report, err := h.capEval.Report(c.Request.Context(), expertID)
	if err != nil {
		h.logger.Error("capability eval: read report failed",
			zap.String("expert_id", expertID.String()), zap.Error(err))
		response.InternalError(c)
		return
	}
	if report == nil {
		response.OK(c, map[string]interface{}{
			"expert_id": expertID,
			"measured":  false,
		})
		return
	}
	response.OK(c, map[string]interface{}{
		"expert_id": expertID,
		"measured":  true,
		"report":    report,
	})
}

// ============================================================
// CONCEPT RELATIONSHIPS (I4)
// ============================================================

// ExtractExpertConcepts POST /admin/experts/:id/concepts
//
// Asks the model how this expert's topics relate and stores the answer. Synchronous
// and bounded (a handful of calls over the topic list), unlike the capability
// measurement: the topic list is small, and the result is a handful of edges rather
// than per-question evidence.
func (h *AdminHandler) ExtractExpertConcepts(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	if h.concepts == nil {
		response.ServiceUnavailable(c, "concept graph is not wired in this deployment")
		return
	}

	result, err := h.concepts.Extract(c.Request.Context(), expertID)
	if err != nil {
		h.logger.Error("concept extraction failed",
			zap.String("expert_id", expertID.String()), zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, map[string]interface{}{
		"expert_id": expertID,
		"result":    result,
	})
}

// GetExpertConcepts GET /admin/experts/:id/concepts
//
// Reads the stored relationships. Read-only and cheap, so the screen can show the
// structure without triggering an extraction.
func (h *AdminHandler) GetExpertConcepts(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	if h.concepts == nil {
		response.ServiceUnavailable(c, "concept graph is not wired in this deployment")
		return
	}

	edges, err := h.concepts.Load(c.Request.Context(), expertID)
	if err != nil {
		h.logger.Error("concept load failed",
			zap.String("expert_id", expertID.String()), zap.Error(err))
		response.InternalError(c)
		return
	}
	if edges == nil {
		edges = []training.ConceptEdge{}
	}
	response.OK(c, map[string]interface{}{
		"expert_id": expertID,
		"count":     len(edges),
		"edges":     edges,
		"relations": training.ConceptRelations(),
	})
}

// ============================================================
// DEPTH LAYERS (I5)
// ============================================================

// ClassifyExpertDepthLayers POST /admin/experts/:id/depth-layers
//
// Labels a bounded batch of chunks by content kind (what/why, how/trade-offs,
// failure/edge) and reports how many are still unclassified. Deliberately per-call
// bounded and resumable: it costs model calls, so the admin decides how far to go and
// the corpus says how much is left.
func (h *AdminHandler) ClassifyExpertDepthLayers(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	if h.depthLayers == nil {
		response.ServiceUnavailable(c, "depth classification is not wired in this deployment")
		return
	}

	job, err := h.depthLayers.StartBackground(c.Request.Context(), expertID)
	if err != nil {
		if errors.Is(err, training.ErrDepthClassificationRunning) {
			// Return the running job so a second click/tab can follow the same work,
			// rather than launching a duplicate or showing a generic 500.
			job, latestErr := h.depthLayers.LatestJob(c.Request.Context(), expertID)
			if latestErr == nil && job != nil {
				response.OK(c, map[string]interface{}{"expert_id": expertID, "job": job})
				return
			}
			response.Conflict(c, "depth classification is already running")
			return
		}
		h.logger.Error("depth classification failed",
			zap.String("expert_id", expertID.String()), zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, map[string]interface{}{
		"expert_id": expertID,
		"job":       job,
	})
}

// GetExpertDepthLayers GET /admin/experts/:id/depth-layers
//
// Reads the content-kind coverage. This is the answer to "is this expert deep or does
// it only know what things are?", derived from the course rather than declared.
func (h *AdminHandler) GetExpertDepthLayers(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	if h.depthLayers == nil {
		response.ServiceUnavailable(c, "depth classification is not wired in this deployment")
		return
	}

	ctx := c.Request.Context()
	report, err := h.depthLayers.Report(ctx, expertID)
	if err != nil {
		h.logger.Error("depth coverage read failed",
			zap.String("expert_id", expertID.String()), zap.Error(err))
		response.InternalError(c)
		return
	}
	job, err := h.depthLayers.LatestJob(ctx, expertID)
	if err != nil {
		h.logger.Error("depth classification job read failed",
			zap.String("expert_id", expertID.String()), zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, map[string]interface{}{
		"report": report,
		"job":    job,
	})
}

// RegenerateCharter POST /admin/experts/:id/regenerate-charter
// Regenerates ONLY the reasoning charter for an expert whose charter
// is blank (e.g. because the LLM API ran out of credits during ingestion).
//
// WHY this endpoint exists:
//   When charter extraction fails mid-ingestion (402/401/403), the
//   pipeline now pauses. But for experts trained BEFORE that fix
//   (like SCALER-DSA-V5), the job completed with blank reasoning_charter.
//   Re-ingesting wastes 1M+ tokens re-embedding chunks that are already
//   correct. This endpoint regenerates ONLY the charter from the stored
//   transcript, leaving all existing chunks untouched.
//
// Requires: transcript_content stored in the most recent ingestion job.
// If transcript_content is empty, returns 400 — admin must re-upload.
func (h *AdminHandler) RegenerateCharter(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}

	ctx := c.Request.Context()

	// Load expert name
	var expertName string
	err = h.db.QueryRow(ctx,
		`SELECT name FROM experts WHERE id=$1 AND deleted_at IS NULL`,
		expertID,
	).Scan(&expertName)
	if err != nil {
		response.NotFound(c, "expert")
		return
	}

	// Load transcript from most recent ingestion job.
	// WHY most recent: admin may have uploaded multiple transcripts.
	// We use the last one as representative sample for charter extraction.
	// Charter extractor uses only first 8000 chars anyway (see charter_extractor.go).
	var transcriptContent string
	err = h.db.QueryRow(ctx,
		`SELECT COALESCE(transcript_content, '')
		 FROM ingestion_jobs
		 WHERE expert_id = $1
		 ORDER BY created_at DESC
		 LIMIT 1`,
		expertID,
	).Scan(&transcriptContent)
	if err != nil || transcriptContent == "" {
		response.BadRequest(c, "TRANSCRIPT_NOT_STORED",
			"Transcript content not found in ingestion jobs. "+
				"Please re-upload the transcript file and ingest again.")
		return
	}

	// Mark expert as training so it disappears from public listings
	_, _ = h.db.Exec(ctx,
		`UPDATE experts SET is_training=TRUE, updated_at=NOW() WHERE id=$1`,
		expertID,
	)

	// Regenerate charter in background — same goroutine pattern as IngestTranscript
	// WHY 10-minute timeout (not 2 hours):
	//   RegenerateCharter only does charter extraction (single LLM call).
	//   HTTP timeout is 120s. 10min is 5x safety margin.
	//   Charter extraction uses first 8000 chars, takes <2 minutes normally.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		charters := training.NewCharterExtractor(h.gateway, h.logger)

		charter, extractErr := charters.Extract(ctx, transcriptContent, expertName)
		if extractErr != nil {
			if errors.Is(extractErr, context.DeadlineExceeded) {
				h.logger.Error("charter regeneration timeout: exceeded 10-minute deadline",
					zap.String("expert_id", expertID.String()),
					zap.Error(extractErr),
				)
			} else {
				h.logger.Error("charter regeneration failed",
					zap.String("expert_id", expertID.String()),
					zap.Error(extractErr),
				)
			}
			// Reset is_training so expert is not stuck in training state
			// Use fresh context with 5s timeout (original context may be cancelled)
			resetCtx, resetCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer resetCancel()
			_, _ = h.db.Exec(resetCtx,
				`UPDATE experts SET is_training=FALSE, updated_at=NOW() WHERE id=$1`,
				expertID,
			)
			return
		}

		clarificationJSON, _ := json.Marshal(charter.ClarificationCharter)

		// Update charter.
		// reasoning_charter: always overwrite (it was blank, that's why we're here).
		// clarification_charter: preserve existing if new one is empty/null.
		//   WHY ::text cast on both CASE branches:
		//   clarification_charter is JSONB. $2 is text. PostgreSQL CASE requires
		//   both branches to be the same type. Casting ELSE branch to ::text
		//   makes both branches text; PostgreSQL then implicitly casts the
		//   assignment back to JSONB for the column. Without this, PostgreSQL
		//   raises SQLSTATE 42804 "CASE types jsonb and text cannot be matched".
		// training_status: set to 'trained' — chunks are already verified by smoke test.
		_, dbErr := h.db.Exec(ctx,
			`UPDATE experts SET
				reasoning_charter     = $1,
				clarification_charter = CASE
					WHEN $2::text NOT IN ('{}', 'null', '') THEN $2::jsonb
					ELSE clarification_charter
				END,
				training_status = 'trained',
				is_training     = FALSE,
				updated_at      = NOW()
			 WHERE id = $3`,
			charter.ReasoningCharter,
			string(clarificationJSON),
			expertID,
		)
		if dbErr != nil {
			h.logger.Error("charter regeneration DB update failed",
				zap.String("expert_id", expertID.String()),
				zap.Error(dbErr),
			)
			// Reset is_training so expert is not stuck in training state
			// Use fresh context with 5s timeout (original context may be cancelled)
			resetCtx, resetCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer resetCancel()
			_, _ = h.db.Exec(resetCtx,
				`UPDATE experts SET is_training=FALSE, updated_at=NOW() WHERE id=$1`,
				expertID,
			)
			return
		}

		// C2: snapshot the regenerated charter as a new version so the
		// charter change is versioned and drift-detected. Best-effort.
		if h.versions != nil {
			if _, sErr := h.versions.Snapshot(ctx, expertID, "charter_regen", "charter regenerated"); sErr != nil {
				h.logger.Warn("expert version snapshot failed (charter_regen)",
					zap.String("expert_id", expertID.String()),
					zap.Error(sErr),
				)
			}
		}

		h.logger.Info("charter regeneration complete",
			zap.String("expert_id", expertID.String()),
			zap.String("expert_name", expertName),
			zap.Int("charter_length", len(charter.ReasoningCharter)),
		)
	}()

	h.logger.Info("charter regeneration started",
		zap.String("expert_id", expertID.String()),
		zap.String("expert_name", expertName),
	)

	response.OK(c, map[string]interface{}{
		"status":      "regenerating",
		"expert_id":   expertID,
		"expert_name": expertName,
		"message":     "Charter regeneration started. Check expert reasoning_charter in ~30 seconds.",
	})
}

// ResumeIngestionJob POST /admin/experts/:id/jobs/:jobID/resume
// Resumes a failed ingestion job from its last checkpoint.
// Only for status='failed' jobs. For status='paused' use RetryIngestionJob.
// Admin does NOT need to re-upload the transcript.
func (h *AdminHandler) ResumeIngestionJob(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	jobID, err := uuid.Parse(c.Param("jobID"))
	if err != nil {
		response.BadRequest(c, "INVALID_JOB_ID", "invalid job ID")
		return
	}

	ctx := c.Request.Context()

	// Load job — verify it belongs to this expert and is in a resumable state
	var job struct {
		ID                uuid.UUID
		Status            string
		SourcePath        string
		TranscriptContent string // may be empty for old jobs
		CheckpointStage   string
	}
	err = h.db.QueryRow(ctx,
		`SELECT id, status, COALESCE(source_path,''),
		        COALESCE(transcript_content,''),
		        COALESCE(current_stage,'pending')
		 FROM ingestion_jobs
		 WHERE id=$1 AND expert_id=$2`,
		jobID, expertID,
	).Scan(&job.ID, &job.Status, &job.SourcePath, &job.TranscriptContent, &job.CheckpointStage)
	if err != nil {
		response.NotFound(c, "job")
		return
	}

	// P1: resumable from BOTH terminal states:
	//   failed → crashed / errored mid-pipeline
	//   paused → charter LLM failed (see RetryIngestionJob, the explicit path)
	// Previously only 'failed' was accepted, so the UI's "Resume from
	// checkpoint" button on a PAUSED job always returned 400 NOT_RESUMABLE and
	// the admin had to flip the row to 'failed' by hand. Accepting paused here
	// makes both entry points work; /retry stays the explicit paused path.
	if job.Status != "failed" && job.Status != "paused" {
		response.BadRequest(c, "NOT_RESUMABLE",
			fmt.Sprintf("job status is '%s' — only failed or paused jobs can be resumed", job.Status))
		return
	}
	if job.Status == "paused" {
		h.logger.Info("resume called on a paused job — resuming from checkpoint",
			zap.String("job_id", jobID.String()),
		)
	}

	// Stage order, used only to tell whether the chunks are already in the DB.
	// NOTE: 'paused' is intentionally absent → it reads as 0 → chunksInDB=false
	// → transcript required. That is CORRECT: the pipeline stopped BEFORE
	// charter extraction, which needs the transcript text (chunks alone are not
	// enough), and the admin upload path always stores it.
	checkpointStageOrder := map[string]int{
		"pending": 0, "chunking": 1, "topic_extraction": 2,
		"charter_extraction": 3, "embedding": 4, "storing": 5,
		"smoke_test": 6, "complete": 7,
	}

	// Get expert name
	var expertName string
	h.db.QueryRow(ctx, `SELECT name FROM experts WHERE id=$1`, expertID).Scan(&expertName)

	// Are the chunks already in the DB for this job? If yes the pipeline can
	// rebuild from them; if not, the transcript is mandatory to re-chunk.
	chunksInDB := checkpointStageOrder[job.CheckpointStage] >= checkpointStageOrder["topic_extraction"]

	// Prefer the stored transcript WHENEVER we have it. WHY: the step a paused
	// job stopped at (charter extraction) needs the transcript TEXT, not just
	// the chunks — an earlier version passed "" when chunks existed, so charter
	// extraction ran against an empty sample. Passing it costs nothing (the
	// pipeline only reads the first 8000 chars).
	transcriptContent := job.TranscriptContent
	if transcriptContent == "" && !chunksInDB {
		response.BadRequest(c, "TRANSCRIPT_REQUIRED",
			"Job failed before chunking completed and transcript was not stored. "+
				"Please re-upload the transcript file to restart ingestion.")
		return
	}
	// If chunks are in DB and the transcript is missing, IngestTranscript will
	// load them from DB and skip re-chunking.

	// Reset job to resumable state
	_, err = h.db.Exec(ctx,
		`UPDATE ingestion_jobs SET
			status        = 'pending',
			error_message = NULL,
			completed_at  = NULL,
			updated_at    = NOW()
		 WHERE id = $1`,
		jobID,
	)
	if err != nil {
		h.logger.Error("reset job for resume failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	// Mark expert as training again
	_, _ = h.db.Exec(ctx, `UPDATE experts SET is_training=TRUE, updated_at=NOW() WHERE id=$1`, expertID)

	// Resume in background with SAME jobID — LoadCheckpoint() will find the checkpoint
	// WHY 2-hour timeout: same as initial upload. Resume runs the full pipeline.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer cancel()

		_, err := h.ingestion.IngestTranscript(
			ctx,
			jobID, expertID, expertName,
			transcriptContent, // empty if chunks already in DB
			job.SourcePath,
			false,
		)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				h.logger.Error("resume ingestion timeout: exceeded 2-hour deadline",
					zap.String("job_id", jobID.String()),
					zap.Error(err),
				)
				return
			}
			h.logger.Error("resume ingestion failed",
				zap.String("job_id", jobID.String()),
				zap.Error(err),
			)
		}
	}()

	h.logger.Info("ingestion job resumed",
		zap.String("job_id", jobID.String()),
		zap.String("expert_id", expertID.String()),
		zap.String("checkpoint_stage", job.CheckpointStage),
	)

	response.OK(c, map[string]interface{}{
		"job_id":           jobID,
		"status":           "resuming",
		"checkpoint_stage": job.CheckpointStage,
		"message":          fmt.Sprintf("Resuming from stage: %s", job.CheckpointStage),
	})
}

// RetryIngestionJob POST /admin/experts/:id/jobs/:jobID/retry
// Retries a paused ingestion job from its last checkpoint (charter stage).
// Only allowed when job status='paused' — use ResumeIngestionJob for 'failed' jobs.
//
// WHY separate from ResumeIngestionJob:
//   Resume is for crashed/failed jobs (any stage).
//   Retry is specifically for paused jobs (charter LLM failure).
//   Keeping them separate makes the admin UI intent explicit.
func (h *AdminHandler) RetryIngestionJob(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	jobID, err := uuid.Parse(c.Param("jobID"))
	if err != nil {
		response.BadRequest(c, "INVALID_JOB_ID", "invalid job ID")
		return
	}

	ctx := c.Request.Context()

	// Load job — verify it belongs to this expert and is paused.
	var job struct {
		ExpertID          uuid.UUID
		ExpertName        string
		Status            string
		CheckpointStage   string
		TranscriptContent string
		SourcePath        string
	}
	err = h.db.QueryRow(ctx, `
		SELECT ij.expert_id, e.name, ij.status,
		       COALESCE(ij.checkpoint_data->>'stage', 'pending'),
		       COALESCE(ij.transcript_content, ''),
		       COALESCE(ij.source_path, '')
		FROM ingestion_jobs ij
		JOIN experts e ON e.id = ij.expert_id
		WHERE ij.id = $1 AND ij.expert_id = $2`,
		jobID, expertID,
	).Scan(
		&job.ExpertID, &job.ExpertName, &job.Status,
		&job.CheckpointStage, &job.TranscriptContent, &job.SourcePath,
	)
	if err != nil {
		response.NotFound(c, "ingestion job")
		return
	}

	if job.Status != "paused" {
		response.BadRequest(c, "JOB_NOT_PAUSED",
			fmt.Sprintf("job status is '%s', not 'paused' — use /resume for failed jobs", job.Status))
		return
	}

	// Charter extraction needs the transcript text.
	// Chunks are already in DB — IngestTranscript loads them from DB.
	// But charter extraction itself needs the transcript for context.
	if job.TranscriptContent == "" {
		response.BadRequest(c, "TRANSCRIPT_REQUIRED",
			"Transcript content not stored. Please re-upload the transcript file to retry.")
		return
	}

	// Reset job to resumable state.
	// LoadCheckpoint() will find the StagePaused checkpoint and
	// resume from charter_extraction (CharterExtracted=false).
	_, err = h.db.Exec(ctx,
		`UPDATE ingestion_jobs SET
			status        = 'pending',
			error_message = NULL,
			paused_at     = NULL,
			completed_at  = NULL,
			updated_at    = NOW()
		 WHERE id = $1`,
		jobID,
	)
	if err != nil {
		h.logger.Error("reset paused job for retry failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	// Mark expert as training again
	_, _ = h.db.Exec(ctx, `UPDATE experts SET is_training=TRUE, updated_at=NOW() WHERE id=$1`, expertID)

	// Retry in background with SAME jobID — LoadCheckpoint() finds the
	// StagePaused checkpoint and skips chunking + topic extraction.
	// WHY 2-hour timeout: same as initial upload. Retry runs the full pipeline.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer cancel()

		_, err := h.ingestion.IngestTranscript(
			ctx,
			jobID, expertID, job.ExpertName,
			job.TranscriptContent,
			job.SourcePath,
			false,
		)
		if err != nil {
			if errors.Is(err, training.ErrJobPaused) {
				// LLM failed again — job is paused again.
				// Admin will see another llm_failure_decision_required SSE event.
				h.logger.Warn("retry ingestion: charter LLM failed again — job re-paused",
					zap.String("job_id", jobID.String()),
				)
				return
			}
			if errors.Is(err, context.DeadlineExceeded) {
				h.logger.Error("retry ingestion timeout: exceeded 2-hour deadline",
					zap.String("job_id", jobID.String()),
					zap.Error(err),
				)
				return
			}
			h.logger.Error("retry ingestion failed",
				zap.String("job_id", jobID.String()),
				zap.Error(err),
			)
		}
	}()

	h.logger.Info("ingestion job retry started",
		zap.String("job_id", jobID.String()),
		zap.String("expert_id", expertID.String()),
	)

	response.OK(c, map[string]interface{}{
		"job_id":  jobID,
		"status":  "retrying",
		"message": "Retrying charter extraction from checkpoint",
	})
}

// GetIngestionJobs GET /admin/experts/:id/jobs
func (h *AdminHandler) GetIngestionJobs(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT id, status, COALESCE(source_path,''),
		       total_chunks, processed_chunks,
		       COALESCE(error_message,''), started_at, completed_at, created_at,
		       COALESCE(current_stage,'pending'), COALESCE(stage_detail,''),
		       COALESCE(cost_usd,0), estimated_seconds_remaining,
		       COALESCE(resumed_from_checkpoint,false)
		FROM ingestion_jobs
		WHERE expert_id=$1
		ORDER BY created_at DESC LIMIT 20`, expertID)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()

	type jobRow struct {
		ID                      uuid.UUID  `json:"id"`
		Status                  string     `json:"status"`
		SourcePath              string     `json:"source_path"`
		TotalChunks             int        `json:"total_chunks"`
		ProcessedChunks         int        `json:"processed_chunks"`
		ErrorMessage            string     `json:"error_message,omitempty"`
		StartedAt               *time.Time `json:"started_at"`
		CompletedAt             *time.Time `json:"completed_at"`
		CreatedAt               time.Time  `json:"created_at"`
		// Migration 007 fields
		CurrentStage            string     `json:"current_stage"`
		StageDetail             string     `json:"stage_detail"`
		CostUsd                 float64    `json:"cost_usd"`
		EstimatedSecondsRemaining *int     `json:"estimated_seconds_remaining"`
		ResumedFromCheckpoint   bool       `json:"resumed_from_checkpoint"`
	}
	var jobs []jobRow
	for rows.Next() {
		var j jobRow
		if err := rows.Scan(
			&j.ID, &j.Status, &j.SourcePath,
			&j.TotalChunks, &j.ProcessedChunks,
			&j.ErrorMessage, &j.StartedAt, &j.CompletedAt, &j.CreatedAt,
			&j.CurrentStage, &j.StageDetail,
			&j.CostUsd, &j.EstimatedSecondsRemaining,
			&j.ResumedFromCheckpoint,
		); err != nil {
			continue
		}
		jobs = append(jobs, j)
	}
	if jobs == nil {
		jobs = []jobRow{}
	}
	response.OK(c, jobs)
}

// ============================================================
// CLIENT MANAGEMENT
// ============================================================

// ListClients GET /admin/clients
//
// Feature #7 fix (docs bug list): previously returned only
// id/email/full_name/is_active/last_login/created_at - no way to see
// a client's activity at all. Added project_count and message_count
// as correlated subqueries rather than fabricating them on the
// frontend (there was no data to fabricate FROM). message_count only
// counts role='user' rows (messages the client actually sent), not
// assistant responses, so it reads as "how many times has this client
// asked something" rather than double-counting both sides of a turn.
func (h *AdminHandler) ListClients(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT u.id, u.email, u.full_name, u.is_active, u.last_login, u.created_at,
		       COALESCE((
		           SELECT COUNT(*) FROM projects p
		           WHERE p.client_id = u.id AND p.deleted_at IS NULL
		       ), 0) AS project_count,
		       COALESCE((
		           SELECT COUNT(*) FROM messages m
		           JOIN chats c2 ON c2.id = m.chat_id
		           JOIN projects p2 ON p2.id = c2.project_id
		           WHERE p2.client_id = u.id AND m.role = 'user'
		       ), 0) AS message_count
		FROM users u
		WHERE u.role='client' AND u.deleted_at IS NULL
		ORDER BY u.created_at DESC`)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()

	type clientRow struct {
		ID           uuid.UUID  `json:"id"`
		Email        string     `json:"email"`
		FullName     string     `json:"full_name"`
		IsActive     bool       `json:"is_active"`
		LastLogin    *time.Time `json:"last_login"`
		CreatedAt    time.Time  `json:"created_at"`
		ProjectCount int        `json:"project_count"`
		MessageCount int        `json:"message_count"`
	}
	var clients []clientRow
	for rows.Next() {
		var cl clientRow
		if err := rows.Scan(
			&cl.ID, &cl.Email, &cl.FullName, &cl.IsActive, &cl.LastLogin, &cl.CreatedAt,
			&cl.ProjectCount, &cl.MessageCount,
		); err != nil {
			continue
		}
		clients = append(clients, cl)
	}
	if clients == nil {
		clients = []clientRow{}
	}
	response.OK(c, clients)
}

// UpdateClient PATCH /admin/clients/:id
func (h *AdminHandler) UpdateClient(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid client ID")
		return
	}
	var req struct {
		IsActive *bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if req.IsActive != nil {
		_, _ = h.db.Exec(c.Request.Context(),
			`UPDATE users SET is_active=$1, updated_at=NOW() WHERE id=$2 AND role='client'`,
			*req.IsActive, id)
	}
	response.OK(c, map[string]string{"status": "updated"})
}

// ============================================================
// SYSTEM STATS
// ============================================================

// GetStats GET /admin/stats
func (h *AdminHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	var totalExperts, activeExperts int
	h.db.QueryRow(ctx, `SELECT COUNT(*), COUNT(*) FILTER (WHERE is_active=TRUE) FROM experts WHERE deleted_at IS NULL`).Scan(&totalExperts, &activeExperts)

	var totalClients int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role='client' AND deleted_at IS NULL`).Scan(&totalClients)

	var totalProjects int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM projects WHERE deleted_at IS NULL`).Scan(&totalProjects)

	var totalMessages int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM messages`).Scan(&totalMessages)

	var totalChunks int
	h.db.QueryRow(ctx, `SELECT COALESCE(SUM(total_chunks),0) FROM experts`).Scan(&totalChunks)

	var violationCount int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM master_event_log WHERE event_type='china_wall_violation'`).Scan(&violationCount)

	var avgRating float64
	h.db.QueryRow(ctx, `SELECT COALESCE(AVG(score),0) FROM ratings`).Scan(&avgRating)

	gwStats := h.gateway.GetStats()

	response.OK(c, map[string]interface{}{
		"experts": map[string]int{
			"total":  totalExperts,
			"active": activeExperts,
		},
		"clients":          totalClients,
		"projects":         totalProjects,
		"messages":         totalMessages,
		"total_chunks":     totalChunks,
		"violations":       violationCount,
		"avg_rating":       fmt.Sprintf("%.2f", avgRating),
		"llm_total_calls":  gwStats["total_calls"],
		"llm_total_cost":   gwStats["total_cost"],
	})
}

// GetViolations GET /admin/violations
func (h *AdminHandler) GetViolations(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT id, project_id, expert_id, client_id,
		       event_type, COALESCE(reasoning,''), created_at
		FROM master_event_log
		WHERE event_type='china_wall_violation'
		ORDER BY created_at DESC
		LIMIT 100`)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()

	type violationRow struct {
		ID        int64      `json:"id"`
		ProjectID uuid.UUID  `json:"project_id"`
		ExpertID  *uuid.UUID `json:"expert_id"`
		ClientID  uuid.UUID  `json:"client_id"`
		EventType string     `json:"event_type"`
		Reasoning string     `json:"reasoning"`
		CreatedAt time.Time  `json:"created_at"`
	}
	var violations []violationRow
	for rows.Next() {
		var v violationRow
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.ExpertID, &v.ClientID,
			&v.EventType, &v.Reasoning, &v.CreatedAt); err != nil {
			continue
		}
		violations = append(violations, v)
	}
	if violations == nil {
		violations = []violationRow{}
	}
	response.OK(c, violations)
}

// GetRatings GET /admin/ratings
func (h *AdminHandler) GetRatings(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT e.name, e.domain,
		       COUNT(r.id) as total_ratings,
		       ROUND(AVG(r.score)::numeric, 2) as avg_score,
		       COUNT(*) FILTER (WHERE r.score >= 4) as good_ratings,
		       COUNT(*) FILTER (WHERE r.score <= 2) as bad_ratings
		FROM experts e
		LEFT JOIN ratings r ON r.expert_id = e.id
		WHERE e.deleted_at IS NULL
		GROUP BY e.id, e.name, e.domain
		ORDER BY avg_score DESC NULLS LAST`)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()

	type ratingRow struct {
		ExpertName   string  `json:"expert_name"`
		Domain       string  `json:"domain"`
		TotalRatings int     `json:"total_ratings"`
		AvgScore     float64 `json:"avg_score"`
		GoodRatings  int     `json:"good_ratings"`
		BadRatings   int     `json:"bad_ratings"`
	}
	var ratings []ratingRow
	for rows.Next() {
		var r ratingRow
		if err := rows.Scan(&r.ExpertName, &r.Domain, &r.TotalRatings,
			&r.AvgScore, &r.GoodRatings, &r.BadRatings); err != nil {
			continue
		}
		ratings = append(ratings, r)
	}
	if ratings == nil {
		ratings = []ratingRow{}
	}
	response.OK(c, ratings)
}

// ============================================================
// SETTINGS
// ============================================================

// GetSettings GET /admin/settings
func (h *AdminHandler) GetSettings(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(),
		`SELECT key, value, COALESCE(description,''), updated_at FROM system_settings ORDER BY key`)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()

	type settingRow struct {
		Key         string      `json:"key"`
		Value       interface{} `json:"value"`
		Description string      `json:"description"`
		UpdatedAt   time.Time   `json:"updated_at"`
	}
	var settings []settingRow
	for rows.Next() {
		var s settingRow
		var valueJSON []byte
		if err := rows.Scan(&s.Key, &valueJSON, &s.Description, &s.UpdatedAt); err != nil {
			continue
		}
		var v interface{}
		if err := json.Unmarshal(valueJSON, &v); err == nil {
			s.Value = v
		}
		settings = append(settings, s)
	}
	if settings == nil {
		settings = []settingRow{}
	}
	response.OK(c, settings)
}

// UpdateSetting PATCH /admin/settings/:key
func (h *AdminHandler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")
	var req struct {
		Value interface{} `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	valueJSON, err := json.Marshal(req.Value)
	if err != nil {
		response.BadRequest(c, "INVALID_VALUE", "value must be valid JSON")
		return
	}
	adminID := c.MustGet("user_id").(uuid.UUID)
	_, err = h.db.Exec(c.Request.Context(),
		`INSERT INTO system_settings (key, value, updated_by)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (key) DO UPDATE SET
			value=EXCLUDED.value,
			updated_by=EXCLUDED.updated_by,
			updated_at=NOW()`,
		key, string(valueJSON), adminID,
	)
	if err != nil {
		h.logger.Error("update setting failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, map[string]string{"status": "updated", "key": key})
}

// ============================================================
// LLM SETTINGS (provider + API keys from admin panel)
// ============================================================

// GetLLMSettings GET /admin/llm-settings
// Returns current provider and masked API keys.
// Keys are masked (last 4 chars only) — never returned in full.
func (h *AdminHandler) GetLLMSettings(c *gin.Context) {
	ctx := c.Request.Context()

	// Read provider from system_settings
	provider := "openrouter" // default
	var providerJSON []byte
	if err := h.db.QueryRow(ctx,
		`SELECT value FROM system_settings WHERE key = 'llm_provider'`,
	).Scan(&providerJSON); err == nil {
		var p string
		if json.Unmarshal(providerJSON, &p) == nil && p != "" {
			provider = p
		}
	}

	// Read API keys from system_settings (masked)
	maskedKeys := map[string]string{}
	var keysJSON []byte
	if err := h.db.QueryRow(ctx,
		`SELECT value FROM system_settings WHERE key = 'llm_api_keys'`,
	).Scan(&keysJSON); err == nil {
		var keys map[string]string
		if json.Unmarshal(keysJSON, &keys) == nil {
			for k, v := range keys {
				if len(v) > 4 {
					maskedKeys[k] = "****" + v[len(v)-4:]
				} else if v != "" {
					maskedKeys[k] = "****"
				}
			}
		}
	}

	// Read fallback provider from system_settings
	fallbackProvider := "" // empty = not configured
	var fallbackJSON []byte
	if err := h.db.QueryRow(ctx,
		`SELECT value FROM system_settings WHERE key = 'llm_fallback_provider'`,
	).Scan(&fallbackJSON); err == nil {
		var fp string
		if json.Unmarshal(fallbackJSON, &fp) == nil {
			fallbackProvider = fp
		}
	}

	response.OK(c, map[string]interface{}{
		"active_provider":     provider,
		"fallback_provider":   fallbackProvider,
		"available_providers": []string{"openrouter", "deepseek", "anthropic", "gemini", "codecraftapi", "cavoti"},
		"api_keys_configured": maskedKeys,
		"note": "API keys are masked. To update, POST to this endpoint with new values.",
	})
}

// GetLLMHealth GET /admin/llm-health
//
// Reports the two questions a provider health surface has to answer: is anything
// being skipped right now (breaker), and is anything answering slowly (latency).
// Both come from the gateway's own choke point, so they describe real calls
// rather than a separate probe that could disagree with reality.
func (h *AdminHandler) GetLLMHealth(c *gin.Context) {
	response.OK(c, map[string]interface{}{
		"breakers": h.gateway.BreakerSnapshot(),
		"latency":  h.gateway.LatencySnapshot(),
		"note": "Latency is percentiles over the last 200 calls per provider. " +
			"A provider the breaker is skipping records no timing, because it was not called.",
	})
}

// GetLLMModelLimits GET /admin/llm-settings/model-limits
//
// Returns every configured per-model token limit. An empty list is a valid
// answer: with no rows the gateway uses each provider's own maximum, which is
// the behaviour that existed before this table.
func (h *AdminHandler) GetLLMModelLimits(c *gin.Context) {
	limits, err := h.gateway.ListModelLimits(c.Request.Context())
	if err != nil {
		h.logger.Error("list model limits failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	if limits == nil {
		limits = []gateway.ModelLimit{}
	}
	response.OK(c, map[string]interface{}{
		"limits": limits,
		"providers": []string{
			"openrouter", "deepseek", "anthropic", "gemini", "codecraftapi", "cavoti",
		},
		"tiers": []string{
			gateway.LimitTierStrong, gateway.LimitTierFast,
			gateway.LimitTierCheap, gateway.LimitTierDefault,
		},
		"note": "0 = not configured, so the provider's own maximum applies. An input limit is enforced only when set.",
	})
}

// UpdateLLMModelLimits PUT /admin/llm-settings/model-limits
//
// Body: {limits: [{provider, tier, max_input_tokens, max_output_tokens}, ...]}
// Upserts the given rows and leaves every other row untouched, so the screen can
// save one edited row without resending the whole table.
func (h *AdminHandler) UpdateLLMModelLimits(c *gin.Context) {
	var req struct {
		Limits []gateway.ModelLimit `json:"limits"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	validProviders := map[string]bool{
		"openrouter": true, "deepseek": true,
		"anthropic": true, "gemini": true,
		"codecraftapi": true, "cavoti": true,
	}
	validTiers := map[string]bool{
		gateway.LimitTierStrong:  true,
		gateway.LimitTierFast:    true,
		gateway.LimitTierCheap:   true,
		gateway.LimitTierDefault: true,
	}
	for _, l := range req.Limits {
		if !validProviders[l.Provider] {
			response.BadRequest(c, "INVALID_PROVIDER", "provider must be one of: openrouter, deepseek, anthropic, gemini, codecraftapi, cavoti")
			return
		}
		if !validTiers[l.Tier] {
			response.BadRequest(c, "INVALID_TIER", "tier must be: strong, fast, cheap, *")
			return
		}
		if l.MaxInputTokens < 0 || l.MaxOutputTokens < 0 {
			response.BadRequest(c, "INVALID_TOKEN_LIMIT", "token limits must be >= 0 (0 = not configured)")
			return
		}
	}

	adminID := c.MustGet("user_id").(uuid.UUID)
	if err := h.gateway.UpsertModelLimits(c.Request.Context(), &adminID, req.Limits); err != nil {
		h.logger.Error("save model limits failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, map[string]interface{}{"saved": len(req.Limits)})
}

// UpdateLLMSettings POST /admin/llm-settings
// Body: {provider: "deepseek", fallback_provider: "openrouter", api_keys: {"deepseek": "sk-xxx"},
//        codecraftapi_model_cheap: "...", codecraftapi_model_strong: "...", codecraftapi_model_fast: "..."}
// Saves provider + fallback + keys + CodeCraftAPI per-tier model names to system_settings.
// Takes effect immediately (no restart needed).
func (h *AdminHandler) UpdateLLMSettings(c *gin.Context) {
	var req struct {
		Provider         string            `json:"provider"`
		// FallbackProvider: optional. When set, Call() tries this provider
		// if the primary fails all 3 attempts. Empty string = no fallback.
		// Set to "" explicitly to clear an existing fallback.
		FallbackProvider string            `json:"fallback_provider"`
		APIKeys          map[string]string `json:"api_keys"`
		// CodeCraftAPI per-tier model names (optional — only used when provider=codecraftapi)
		CodeCraftAPIModelCheap  string `json:"codecraftapi_model_cheap"`
		CodeCraftAPIModelStrong string `json:"codecraftapi_model_strong"`
		CodeCraftAPIModelFast   string `json:"codecraftapi_model_fast"`
		// Cavoti per-tier model names (optional — only used when provider=cavoti)
		CavotiModelCheap  string `json:"cavoti_model_cheap"`
		CavotiModelStrong string `json:"cavoti_model_strong"`
		CavotiModelFast   string `json:"cavoti_model_fast"`
		// ClearFallback: set to true to explicitly remove the fallback provider.
		// WHY a separate flag: empty string in FallbackProvider is ambiguous
		// ("not provided" vs "clear it"). ClearFallback=true is unambiguous.
		ClearFallback bool `json:"clear_fallback"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	validProviders := map[string]bool{
		"openrouter": true, "deepseek": true,
		"anthropic": true, "gemini": true,
		"codecraftapi": true, "cavoti": true,
	}
	if req.Provider != "" && !validProviders[req.Provider] {
		response.BadRequest(c, "INVALID_PROVIDER",
			"provider must be: openrouter, deepseek, anthropic, gemini, codecraftapi, cavoti")
		return
	}
	if req.FallbackProvider != "" && !validProviders[req.FallbackProvider] {
		response.BadRequest(c, "INVALID_FALLBACK_PROVIDER",
			"fallback_provider must be: openrouter, deepseek, anthropic, gemini, codecraftapi, cavoti")
		return
	}

	ctx := c.Request.Context()
	adminID := c.MustGet("user_id").(uuid.UUID)

	// Save provider
	if req.Provider != "" {
		providerJSON, _ := json.Marshal(req.Provider)
		_, err := h.db.Exec(ctx,
			`INSERT INTO system_settings (key, value, updated_by)
			 VALUES ('llm_provider', $1, $2)
			 ON CONFLICT (key) DO UPDATE SET
				value = EXCLUDED.value,
				updated_by = EXCLUDED.updated_by,
				updated_at = NOW()`,
			string(providerJSON), adminID,
		)
		if err != nil {
			h.logger.Error("save llm provider failed", zap.Error(err))
			response.InternalError(c)
			return
		}
	}

	// Save fallback provider
	if req.ClearFallback {
		// Explicitly clear fallback — delete the key so getFallbackProvider() returns nil.
		_, err := h.db.Exec(ctx,
			`DELETE FROM system_settings WHERE key = 'llm_fallback_provider'`,
		)
		if err != nil {
			h.logger.Error("clear llm fallback provider failed", zap.Error(err))
			response.InternalError(c)
			return
		}
	} else if req.FallbackProvider != "" {
		fallbackJSON, _ := json.Marshal(req.FallbackProvider)
		_, err := h.db.Exec(ctx,
			`INSERT INTO system_settings (key, value, updated_by)
			 VALUES ('llm_fallback_provider', $1, $2)
			 ON CONFLICT (key) DO UPDATE SET
				value = EXCLUDED.value,
				updated_by = EXCLUDED.updated_by,
				updated_at = NOW()`,
			string(fallbackJSON), adminID,
		)
		if err != nil {
			h.logger.Error("save llm fallback provider failed", zap.Error(err))
			response.InternalError(c)
			return
		}
	}

	// Save API keys (merge with existing, don't overwrite unset keys)
	if len(req.APIKeys) > 0 {
		// Load existing keys first
		existing := map[string]string{}
		var existingJSON []byte
		if err := h.db.QueryRow(ctx,
			`SELECT value FROM system_settings WHERE key = 'llm_api_keys'`,
		).Scan(&existingJSON); err == nil {
			json.Unmarshal(existingJSON, &existing)
		}
		// Merge: new values override existing
		for k, v := range req.APIKeys {
			if v != "" {
				existing[k] = v
			}
		}
		mergedJSON, _ := json.Marshal(existing)
		_, err := h.db.Exec(ctx,
			`INSERT INTO system_settings (key, value, updated_by)
			 VALUES ('llm_api_keys', $1, $2)
			 ON CONFLICT (key) DO UPDATE SET
				value = EXCLUDED.value,
				updated_by = EXCLUDED.updated_by,
				updated_at = NOW()`,
			string(mergedJSON), adminID,
		)
		if err != nil {
			h.logger.Error("save llm api keys failed", zap.Error(err))
			response.InternalError(c)
			return
		}
	}

	// Save per-tier model names for CodeCraftAPI and Cavoti (only when provided)
	modelSettings := map[string]string{
		"codecraftapi_model_cheap":  req.CodeCraftAPIModelCheap,
		"codecraftapi_model_strong": req.CodeCraftAPIModelStrong,
		"codecraftapi_model_fast":   req.CodeCraftAPIModelFast,
		"cavoti_model_cheap":        req.CavotiModelCheap,
		"cavoti_model_strong":       req.CavotiModelStrong,
		"cavoti_model_fast":         req.CavotiModelFast,
	}
	for key, value := range modelSettings {
		if value == "" {
			continue // Don't overwrite existing value with empty
		}
		valueJSON, _ := json.Marshal(value)
		_, err := h.db.Exec(ctx,
			`INSERT INTO system_settings (key, value, updated_by)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (key) DO UPDATE SET
				value = EXCLUDED.value,
				updated_by = EXCLUDED.updated_by,
				updated_at = NOW()`,
			key, string(valueJSON), adminID,
		)
		if err != nil {
			h.logger.Error("save codecraftapi model name failed",
				zap.String("key", key), zap.Error(err))
			response.InternalError(c)
			return
		}
	}

	h.logger.Info("LLM settings updated",
		zap.String("provider", req.Provider),
		zap.String("fallback_provider", req.FallbackProvider),
		zap.Int("keys_updated", len(req.APIKeys)),
	)
	response.OK(c, map[string]string{
		"status":            "updated",
		"provider":          req.Provider,
		"fallback_provider": req.FallbackProvider,
		"note":              "Changes take effect on next LLM call. No restart needed.",
	})
}

// isUniqueViolation checks if error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unique")
}

// ============================================================
// CODECRAFTAPI MODELS (proxy to /v1/models)
// ============================================================

// GetCodeCraftModels GET /admin/codecraftapi/models
// Proxies to CodeCraftAPI's GET /v1/models endpoint.
// Admin UI calls this to populate model picker dropdowns.
//
// WHY proxy instead of direct frontend call:
//   API key must never leave the server. Frontend never calls CodeCraftAPI directly.
//
// Mental execution:
//   1. Read codecraftapi key from system_settings (llm_api_keys["codecraftapi"])
//   2. If empty → 400 KEY_NOT_CONFIGURED
//   3. GET https://codecraftapi.com/v1/models
//   4. Return model list to admin UI
func (h *AdminHandler) GetCodeCraftModels(c *gin.Context) {
	ctx := c.Request.Context()

	// Read CodeCraftAPI key from system_settings
	apiKey := ""
	var keysJSON []byte
	if err := h.db.QueryRow(ctx,
		`SELECT value FROM system_settings WHERE key = 'llm_api_keys'`,
	).Scan(&keysJSON); err == nil {
		var keys map[string]string
		if json.Unmarshal(keysJSON, &keys) == nil {
			apiKey = keys["codecraftapi"]
		}
	}

	if apiKey == "" {
		response.BadRequest(c, "KEY_NOT_CONFIGURED",
			"CodeCraftAPI key not configured. Add key in LLM Settings first.")
		return
	}

	// Read base URL from config (via gateway, which already has it)
	// We use the default if not overridden
	baseURL := "https://codecraftapi.com/v1"

	// Proxy GET /v1/models to CodeCraftAPI
	httpClient := &http.Client{Timeout: 15 * time.Second}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/models", nil)
	if err != nil {
		h.logger.Error("codecraftapi models: build request failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		h.logger.Error("codecraftapi models: http call failed", zap.Error(err))
		c.JSON(502, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "UPSTREAM_ERROR",
				"message": "Could not fetch models from CodeCraftAPI: " + err.Error(),
			},
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Warn("codecraftapi models: upstream returned non-200",
			zap.Int("status", resp.StatusCode),
		)
		c.JSON(502, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "UPSTREAM_ERROR",
				"message": fmt.Sprintf("CodeCraftAPI returned status %d", resp.StatusCode),
			},
		})
		return
	}

	// Parse CodeCraftAPI response.
	// CodeCraftAPI returns OpenAI-compatible format:
	//   { "object": "list", "data": [{"id": "model-1"}, ...] }
	// We extract the "data" array and return it directly so the frontend
	// receives a clean []CodeCraftModel, not a nested object.
	// WHY extract here (not in frontend):
	//   Backend is the single source of truth for the wire contract.
	//   Frontend should never need to know CodeCraftAPI's internal envelope.
	var rawResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		h.logger.Error("codecraftapi models: decode response failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	// Extract the models array from the response.
	// Handle two cases:
	//   Case 1: { "object": "list", "data": [...] }  ← standard OpenAI format
	//   Case 2: [...] (bare array, some providers do this)
	var models []interface{}
	if dataField, ok := rawResponse["data"]; ok {
		if arr, ok := dataField.([]interface{}); ok {
			models = arr
		}
	}
	if models == nil {
		// Fallback: response was not the expected shape — return empty list
		// rather than crashing. Admin will see "no models" and can check
		// their CodeCraftAPI key/account.
		h.logger.Warn("codecraftapi models: unexpected response shape, returning empty list")
		models = []interface{}{}
	}

	response.OK(c, models)
}

// ============================================================
// EMBEDDING SETTINGS
// ============================================================

// GetEmbeddingSettings GET /admin/embedding-settings
// Returns current embedding provider config.
func (h *AdminHandler) GetEmbeddingSettings(c *gin.Context) {
	ctx := c.Request.Context()

	embeddingProvider := "sidecar" // default
	var provJSON []byte
	if err := h.db.QueryRow(ctx,
		`SELECT value FROM system_settings WHERE key = 'embedding_provider'`,
	).Scan(&provJSON); err == nil {
		var p string
		if json.Unmarshal(provJSON, &p) == nil && p != "" {
			embeddingProvider = p
		}
	}

	embeddingModel := ""
	var modelJSON []byte
	if err := h.db.QueryRow(ctx,
		`SELECT value FROM system_settings WHERE key = 'embedding_model'`,
	).Scan(&modelJSON); err == nil {
		var m string
		if json.Unmarshal(modelJSON, &m) == nil {
			embeddingModel = m
		}
	}

	response.OK(c, map[string]interface{}{
		"embedding_provider":   embeddingProvider,
		"embedding_model":      embeddingModel,
		"available_providers": []string{"sidecar", "codecraftapi"},
		"note": "Changing embedding provider requires re-ingesting ALL transcripts. Existing vectors will be incompatible.",
	})
}

// UpdateEmbeddingSettings POST /admin/embedding-settings
// Body: {embedding_provider: "codecraftapi", embedding_model: "cc-embed-X"}
// Saves embedding config to system_settings. Takes effect on next Embed() call.
//
// Mental execution:
//   Switch to codecraftapi:
//     Input: {embedding_provider: "codecraftapi", embedding_model: "cc-embed-X"}
//     1. Validate provider is "sidecar" or "codecraftapi"
//     2. If codecraftapi and model empty → 400
//     3. UPSERT embedding_provider + embedding_model
//     4. Return {status: "updated"}
//
//   Switch back to sidecar:
//     Input: {embedding_provider: "sidecar"}
//     1. Validate
//     2. UPSERT embedding_provider = "sidecar"
//     3. Return {status: "updated"}
func (h *AdminHandler) UpdateEmbeddingSettings(c *gin.Context) {
	var req struct {
		EmbeddingProvider string `json:"embedding_provider"`
		EmbeddingModel    string `json:"embedding_model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	validEmbeddingProviders := map[string]bool{
		"sidecar": true, "codecraftapi": true,
	}
	if !validEmbeddingProviders[req.EmbeddingProvider] {
		response.BadRequest(c, "INVALID_PROVIDER",
			"embedding_provider must be: sidecar, codecraftapi")
		return
	}

	if req.EmbeddingProvider == "codecraftapi" && req.EmbeddingModel == "" {
		response.BadRequest(c, "MODEL_REQUIRED",
			"embedding_model is required when embedding_provider is codecraftapi")
		return
	}

	ctx := c.Request.Context()
	adminID := c.MustGet("user_id").(uuid.UUID)

	// Save embedding_provider
	providerJSON, _ := json.Marshal(req.EmbeddingProvider)
	_, err := h.db.Exec(ctx,
		`INSERT INTO system_settings (key, value, updated_by)
		 VALUES ('embedding_provider', $1, $2)
		 ON CONFLICT (key) DO UPDATE SET
			value = EXCLUDED.value,
			updated_by = EXCLUDED.updated_by,
			updated_at = NOW()`,
		string(providerJSON), adminID,
	)
	if err != nil {
		h.logger.Error("save embedding_provider failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	// Save embedding_model (only when codecraftapi; clear when switching back to sidecar)
	modelValue := req.EmbeddingModel
	modelJSON, _ := json.Marshal(modelValue)
	_, err = h.db.Exec(ctx,
		`INSERT INTO system_settings (key, value, updated_by)
		 VALUES ('embedding_model', $1, $2)
		 ON CONFLICT (key) DO UPDATE SET
			value = EXCLUDED.value,
			updated_by = EXCLUDED.updated_by,
			updated_at = NOW()`,
		string(modelJSON), adminID,
	)
	if err != nil {
		h.logger.Error("save embedding_model failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	h.logger.Info("embedding settings updated",
		zap.String("provider", req.EmbeddingProvider),
		zap.String("model", req.EmbeddingModel),
	)
	response.OK(c, map[string]string{
		"status": "updated",
		"note":   "Changes take effect on next Embed() call. No restart needed.",
	})
}

// ============================================================
// GATE THRESHOLDS (B4) — per-domain Gate 1 usable/strong config
// ============================================================
//
// Thresholds live in gate_thresholds (migration 025), not in code (P8).
// Calibration from ratings only WRITES proposals (source='calibrated');
// GateSystem only consumes source IN ('applied','manual'). Admin must
// explicitly Apply — no blind auto-tune (P7; eval harness not yet in repo).

// ListGateThresholds GET /admin/gate-thresholds
// Returns every known domain + its row (if any) + the effective pair
// GateSystem would use right now (applied/manual, else package default).
func (h *AdminHandler) ListGateThresholds(c *gin.Context) {
	rows, err := workflow.ListGateThresholds(c.Request.Context(), h.db)
	if err != nil {
		h.logger.Error("list gate thresholds failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	def := workflow.DefaultGateThresholds()
	response.OK(c, gin.H{
		"defaults": gin.H{
			"usable": def.Usable,
			"strong": def.Strong,
			"source": def.Source,
		},
		"thresholds": rows,
	})
}

// CalibrateGateThresholds POST /admin/gate-thresholds/calibrate
// Aggregates ratings per expert domain and writes calibrated proposals
// for domains with enough samples. Does NOT apply them.
func (h *AdminHandler) CalibrateGateThresholds(c *gin.Context) {
	proposed, err := workflow.ProposeGateThresholds(c.Request.Context(), h.db, h.logger)
	if err != nil {
		h.logger.Error("calibrate gate thresholds failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{
		"proposed_count": len(proposed),
		"proposed":       proposed,
		"note":           "Proposals are source='calibrated' only. Call POST /admin/gate-thresholds/:domain/apply to make one live.",
	})
}

// ApplyGateThreshold POST /admin/gate-thresholds/:domain/apply
// Promotes a calibrated (or existing) row to source='applied'. Live for
// GateSystem within thresholdCacheTTL (30s) or on next cold miss.
func (h *AdminHandler) ApplyGateThreshold(c *gin.Context) {
	domain := c.Param("domain")
	adminID := c.MustGet("user_id").(uuid.UUID)
	row, err := workflow.ApplyGateThreshold(c.Request.Context(), h.db, domain, adminID)
	if err != nil {
		h.logger.Warn("apply gate threshold failed",
			zap.String("domain", domain), zap.Error(err))
		response.BadRequest(c, "APPLY_FAILED", err.Error())
		return
	}
	response.OK(c, row)
}

// SetGateThreshold PATCH /admin/gate-thresholds/:domain
// Manual override — source='manual', immediately live.
func (h *AdminHandler) SetGateThreshold(c *gin.Context) {
	domain := c.Param("domain")
	var req struct {
		Usable float64 `json:"usable" binding:"required"`
		Strong float64 `json:"strong" binding:"required"`
		Notes  string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	adminID := c.MustGet("user_id").(uuid.UUID)
	row, err := workflow.SetGateThreshold(
		c.Request.Context(), h.db, domain, req.Usable, req.Strong, adminID, req.Notes,
	)
	if err != nil {
		h.logger.Warn("set gate threshold failed",
			zap.String("domain", domain), zap.Error(err))
		response.BadRequest(c, "SET_FAILED", err.Error())
		return
	}
	response.OK(c, row)
}
