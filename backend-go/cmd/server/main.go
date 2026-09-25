package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	adminpkg "ai_avengers/backend/internal/admin"
	"ai_avengers/backend/internal/auth"
	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/byoexpert"
	"ai_avengers/backend/internal/category"
	"ai_avengers/backend/internal/chat"
	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/config"
	appcontext "ai_avengers/backend/internal/context"
	"ai_avengers/backend/internal/db"
	"ai_avengers/backend/internal/decision"
	"ai_avengers/backend/internal/entitlement"
	"ai_avengers/backend/internal/eval"
	"ai_avengers/backend/internal/explain"
	"ai_avengers/backend/internal/expert"
	"ai_avengers/backend/internal/expertversion"
	"ai_avengers/backend/internal/docextract"
	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/jobevents"
	"ai_avengers/backend/internal/knowledge"
	"ai_avengers/backend/internal/memory"
	"ai_avengers/backend/internal/message"
	"ai_avengers/backend/internal/middleware"
	"ai_avengers/backend/internal/ml"
	"ai_avengers/backend/internal/monitoring"
	"ai_avengers/backend/internal/observability"
	"ai_avengers/backend/internal/orchestrator"
	"ai_avengers/backend/internal/outbox"
	"ai_avengers/backend/internal/project"
	"ai_avengers/backend/internal/provenance"
	"ai_avengers/backend/internal/rating"
	"ai_avengers/backend/internal/reliability"
	"ai_avengers/backend/internal/repo"
	"ai_avengers/backend/internal/response"
	"ai_avengers/backend/internal/selflearning"
	"ai_avengers/backend/internal/tenant"
	"ai_avengers/backend/internal/training"
	"ai_avengers/backend/internal/usage"
	"ai_avengers/backend/internal/validation"
	"ai_avengers/backend/internal/workflow"
)

func main() {
	// Build logger first — needed for all subsequent steps
	logger, err := buildLogger("info")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("AI Avengers backend starting")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}

	// Rebuild logger with configured level
	logger, err = buildLogger(cfg.Log.Level)
	if err != nil {
		logger.Fatal("failed to rebuild logger", zap.Error(err))
	}

	// Cancellable root context — background jobs (outbox dispatcher, cleaners) stop on shutdown.
	ctx, cancelRoot := context.WithCancel(context.Background())
	defer cancelRoot()

	// Connect to PostgreSQL
	postgres, err := db.Connect(ctx, cfg.Database, logger)
	if err != nil {
		logger.Fatal("failed to connect to PostgreSQL", zap.Error(err))
	}
	defer postgres.Close()

	// Connect to Redis
	redisClient, err := db.ConnectRedis(ctx, cfg.Redis, logger)
	if err != nil {
		logger.Fatal("failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize services
	jwtService := auth.NewJWTService(cfg.JWT, redisClient.Client, logger)
	authService := auth.NewAuthService(postgres.Pool, jwtService, logger)
	// Self-registration is off by default (accounts are admin-provisioned).
	// Set SELF_REGISTRATION_ENABLED=true to re-enable /auth/register.
	authService.SetSelfRegistrationEnabled(cfg.Auth.SelfRegistrationEnabled)

	// Stage 0 seams: ports + adapters (modular monolith, extract-ready).
	// Entitlement wraps user_expert_grants; knowledge is Expert Knowledge read port;
	// outbox is the RDBMS event broker (SKIP LOCKED dispatcher below).
	entitlementChecker := entitlement.NewChecker(postgres.Pool)
	_ = entitlementChecker // composed into auth.MustHaveExpertAccess; held for future DI
	knowledgeReader := knowledge.NewReader(postgres.Pool)
	_ = knowledgeReader // available for incremental call-site migration
	outboxStore := outbox.NewStore(postgres.Pool, logger)
	authService.SetEventPublisher(outboxStore)
	outboxDispatcher := outbox.NewDispatcher(outboxStore, func(ctx context.Context, aggregateType, eventType, aggregateID string, payload []byte) error {
		observability.Global.IncOutboxPublished()
		logger.Debug("outbox dispatched",
			zap.String("aggregate_type", aggregateType),
			zap.String("event_type", eventType),
			zap.String("aggregate_id", aggregateID),
			zap.Int("payload_bytes", len(payload)),
		)
		return nil
	}, logger)
	go outboxDispatcher.Run(ctx)

	modelGateway := gateway.NewModelGateway(cfg.LLM, logger)
	modelGateway.SetDB(postgres.Pool) // enables runtime provider override from admin panel
	mlClient := ml.NewSidecarClient(cfg.ML, logger)

	// Verify ML sidecar is running
	if err := mlClient.HealthCheck(ctx); err != nil {
		logger.Warn("ML sidecar not available — embedding features will be degraded",
			zap.Error(err),
		)
		// Don't fatal — server can start without ML sidecar
		// WHY: Allow server to start even if ML sidecar is still loading models
	}

	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// DomainRegistry: seed defaults, load all profiles into memory.
	// Must complete before buildRouter — Enforce() calls registry.Get() per question.
	domainRegistry := chinawall.NewDomainRegistry(postgres.Pool, logger)
	if err := domainRegistry.Init(ctx); err != nil {
		logger.Fatal("failed to initialize domain registry", zap.Error(err))
	}
	// SetGlobalRegistry enables IsProblemSolvingDomain() used by Gate 1.
	chinawall.SetGlobalRegistry(domainRegistry)

	// CategoryRegistry: loads expert_categories rows into memory.
	// Separate from domainRegistry (chinawall package) by design —
	// CATEGORY_TEMPLATE_HANDOFF.md CT-L1 requires the existing
	// DomainRegistry/DomainProfile to stay untouched. Must complete
	// before buildRouter — admin category_id validation (CreateExpert/
	// UpdateExpert) calls categoryRegistry.Get() per request.
	categoryRegistry := category.NewRegistry(postgres.Pool, logger)
	if err := categoryRegistry.Init(ctx); err != nil {
		logger.Fatal("failed to initialize category registry", zap.Error(err))
	}

	// Build CodeCraftAPI embedder.
	// Constructor takes db (reads embedding_model from system_settings at call time)
	// and a dedicated http.Client with 300s timeout (batch embedding can be slow).
	// Does NOT take model name — reads it from system_settings on every Embed() call.
	ccEmbedder := ml.NewCodeCraftAPIEmbedder(
		cfg.LLM.CodeCraftAPIKey,
		cfg.LLM.CodeCraftAPIBaseURL,
		postgres.Pool,
		&http.Client{Timeout: 300 * time.Second},
		logger,
	)

	// DynamicEmbedder: reads embedding_provider from system_settings at call time.
	// Falls back to mlClient (sidecar) when provider = "sidecar" or DB read fails.
	// WHY dynamic: admin switches embedding provider from UI — takes effect
	// on next Embed() call with no server restart needed.
	embedder := ml.NewDynamicEmbedder(postgres.Pool, mlClient, ccEmbedder, logger)

	// Build router — single call, single definition
	router := buildRouter(ctx, cfg, logger, postgres, redisClient, jwtService, authService, modelGateway, mlClient, embedder, domainRegistry, categoryRegistry)

	// 24-hour auto-fail checker for paused ingestion jobs.
	// WHY here not in admin_handler: server-lifecycle concern, not per-request.
	// Starts once at boot, runs every 5 minutes for the lifetime of the process.
	// WHY 5-minute interval: precise enough for a 24h window, single admin user.
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				res, err := postgres.Exec(context.Background(),
					`UPDATE ingestion_jobs
					 SET status        = 'failed',
					     error_message = 'Auto-failed: paused for >24h without admin action',
					     completed_at  = NOW()
					 WHERE status = 'paused'
					   AND paused_at < NOW() - INTERVAL '24 hours'`,
				)
				if err != nil {
					logger.Warn("paused job auto-fail checker: DB error", zap.Error(err))
					continue
				}
				if res.RowsAffected() > 0 {
					logger.Info("paused job auto-fail checker: expired jobs failed",
						zap.Int64("count", res.RowsAffected()),
					)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Workspace cleanup job: delete workspaces older than 7 days.
	// WHY here: server-lifecycle concern, runs for lifetime of process.
	// Runs every 24 hours, deletes workspaces from completed/failed/cancelled workflows.
	// WHY 7 days: balance between debugging needs and disk space.
	// WHY 24-hour interval: daily cleanup is sufficient, low overhead.
	//
	// WHY re-read AIDER_WORKSPACE_ROOT here instead of sharing buildRouter's
	// local: buildRouter's workspaceRoot is scoped to that function (used to
	// construct AiderRunner) and is not visible from main(). Same env var,
	// same default — kept in sync with buildRouter's copy below.
	workspaceRoot := os.Getenv("AIDER_WORKSPACE_ROOT")
	if workspaceRoot == "" {
		workspaceRoot = "/tmp/ai_avengers_workspaces"
	}
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		// Run immediately on startup (don't wait 24 hours)
		cleanupWorkspaces := func() {
			ctx := context.Background()
			// Find completed/failed/cancelled workflows
			rows, err := postgres.Query(ctx,
				`SELECT id FROM workflows
				 WHERE status IN ('completed', 'failed', 'cancelled')
				   AND updated_at < NOW() - INTERVAL '7 days'`,
			)
			if err != nil {
				logger.Warn("workspace cleanup: DB query error", zap.Error(err))
				return
			}
			defer rows.Close()

			var deletedCount int
			for rows.Next() {
				var workflowID uuid.UUID
				if err := rows.Scan(&workflowID); err != nil {
					logger.Warn("workspace cleanup: scan error", zap.Error(err))
					continue
				}

				workspaceDir := filepath.Join(workspaceRoot, workflowID.String())
				// Check if workspace exists
				if _, err := os.Stat(workspaceDir); os.IsNotExist(err) {
					continue
				}

				// Delete workspace
				if err := os.RemoveAll(workspaceDir); err != nil {
					logger.Warn("workspace cleanup: delete error",
						zap.String("workspace", workspaceDir),
						zap.Error(err),
					)
					continue
				}

				deletedCount++
				logger.Info("workspace cleanup: deleted workspace",
					zap.String("workflow_id", workflowID.String()),
				)
			}

			if deletedCount > 0 {
				logger.Info("workspace cleanup: completed",
					zap.Int("deleted", deletedCount),
				)
			}
		}
		// Run immediately on startup
		cleanupWorkspaces()
		// Then run every 24 hours
		for {
			select {
			case <-ticker.C:
				cleanupWorkspaces()
			case <-ctx.Done():
				return
			}
		}
	}()

	// Build HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second, // Long for SSE streaming
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server listening",
			zap.String("addr", server.Addr),
			zap.String("env", cfg.Server.Env),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Fatal("server error", zap.Error(err))
	case sig := <-quit:
		logger.Info("shutdown signal received", zap.String("signal", sig.String()))
	}

	// Stop background jobs (outbox dispatcher, cleaners) before draining HTTP.
	cancelRoot()

	// Graceful shutdown
	// WHY 30 seconds: Allow in-flight requests to complete
	// SSE connections may take longer — they'll be cut off after 30s
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", zap.Error(err))
	}

	logger.Info("server stopped gracefully")
}

// buildRouter creates the Gin router with all routes and middleware.
// ONE definition only — the stub draft has been deleted.
//
// WHY gin.New() not gin.Default():
// gin.Default() adds Logger and Recovery middleware automatically.
// We use our own structured zap logger and recovery middleware.
// gin.New() gives us full control over middleware order.
func buildRouter(
	ctx context.Context,
	cfg *config.Config,
	logger *zap.Logger,
	postgres *db.Pool,
	redisClient *db.RedisClient,
	jwtService *auth.JWTService,
	authService *auth.AuthService,
	modelGateway *gateway.ModelGateway,
	mlClient *ml.SidecarClient,
	embedder ml.Embedder,
	domainRegistry *chinawall.DomainRegistry,
	categoryRegistry *category.Registry,
) *gin.Engine {
	router := gin.New()

	// Middleware order: CORS -> Recovery -> RequestID -> Logger -> RateLimit
	// WHY CORS is first: OPTIONS preflight has no auth header.
	// If any other middleware runs before CORS, preflight gets rejected
	// and browser blocks all API calls.
	allowedOrigins := cfg.CORSAllowedOrigins
	if len(allowedOrigins) == 0 {
		// Default: allow dev frontend. Production sets CORS_ALLOWED_ORIGINS.
		allowedOrigins = []string{"http://localhost:3000"}
	}
	router.Use(middleware.CORSMiddleware(allowedOrigins))
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.RateLimitMiddleware(redisClient.Client, cfg.RateLimit.PerIP, logger))

	// Initialize core services
	memManager := memory.NewManager(postgres.Pool, redisClient.Client, embedder, logger)
	// C1: signed provenance chain service. Signing key falls back to the
	// AES key (config.applyDefaults), so it is always present in deploy.
	provSvc := provenance.NewService(
		postgres.Pool, []byte(cfg.Provenance.SigningKey), cfg.Provenance.Enabled, logger,
	)
	// C2: expert versioning + capability drift service.
	versionSvc := expertversion.NewService(
		postgres.Pool, cfg.Versioning.DriftThreshold, logger,
	)
	// C3: golden-set eval run store (viewed from admin; run by cmd/eval).
	evalStore := eval.NewStore(postgres.Pool)
	// C4: tenant isolation + enterprise controls. Kill switch TENANT_ISOLATION_ENABLED.
	tenantSvc := tenant.NewService(postgres.Pool, cfg.Tenant.IsolationEnabled, logger)
	// C5: cost/usage analytics product. The gateway records every real LLM
	// call here (single choke point, G5); admin reads it back grouped.
	usageSvc := usage.NewService(postgres.Pool, logger)
	modelGateway.SetUsageRecorder(usageSvc)
	// C6: knowledge freshness scanner (staleness + embedding mismatch + orphan
	// citations). nil when disabled (FRESHNESS_ENABLED=false).
	var freshnessSvc *knowledge.Freshness
	if cfg.Freshness.Enabled {
		freshnessSvc = knowledge.NewFreshness(postgres.Pool,
			knowledge.Policy{MaxCorpusAgeDays: cfg.Freshness.MaxCorpusAgeDays}, logger)
	}
	// T1: durable ingestion timeline (append-only event log + Redis notification).
	// Shared by both ingestion pipelines (admin + BYO) and by the SSE stream, so
	// the writer and the readers use exactly one implementation.
	eventsStore := jobevents.NewStore(postgres.Pool, redisClient.Client, logger)

	// D2: document extraction (stage 0). Uploads are no longer assumed to be
	// plain text — PDF/DOCX/XLSX/PPTX/... are converted to text before the
	// pipeline runs. Parsing happens in the ML sidecar (hostile-input isolation);
	// .txt/.md are decoded in-process so they keep working if the sidecar is down.
	docExtractor := docextract.NewExtractor(cfg.ML.SidecarURL, cfg.ML.TimeoutSeconds, logger)

	// C8: bring-your-own-expert. The training pipeline is stateless, so a
	// second instance is safe and keeps admin/BYO wiring independent.
	byoIngestor := training.NewIngestionPipeline(postgres.Pool, embedder, mlClient, modelGateway, eventsStore, docExtractor, logger)
	byoSvc := byoexpert.NewService(postgres.Pool, tenantSvc, byoIngestor,
		byoexpert.Policy{
			Enabled:           cfg.ByoExpert.Enabled,
			DefaultMaxExperts: cfg.ByoExpert.DefaultMaxExperts,
		}, logger)
	// B7: background L2 consolidation + preference decay. Stops on root ctx cancel.
	// Cheap model, 6h interval, phase-end (cooldown) not per-event.
	go memory.NewConsolidator(memManager, modelGateway, logger).Run(ctx)
	// domainRegistry is already initialized in main() and passed here.
	chinawallEnforcer := chinawall.NewEnforcer(cfg.ChinaWall, modelGateway, mlClient, logger, domainRegistry)
	decisionEngine := decision.NewEngine(postgres.Pool, modelGateway, chinawallEnforcer, logger)
	contextAssembler := appcontext.NewAssembler(
		postgres.Pool, embedder, mlClient, memManager,
		cfg.Context.MaxTokens, cfg.Context.RecentMessages,
		cfg.Context.SemanticTopK, cfg.Context.CourseChunksTopK,
		logger,
	)

	// Self-Learning Mode: Understand → Extract → Verify.
	// Converts raw questions into domain-specific signal before RAG.
	// WHY always enabled: accuracy improvement justifies 3×ModelCheap cost.
	// To disable: pass nil instead of questionProcessor to NewOrchestrator.
	questionProcessor := selflearning.NewQuestionProcessor(modelGateway, logger)

	orch := orchestrator.NewOrchestrator(postgres.Pool, contextAssembler, decisionEngine, memManager, categoryRegistry, questionProcessor, modelGateway, logger)

	// Initialize domain services
	projectSvc := project.NewService(postgres.Pool, logger)
	chatSvc := chat.NewService(postgres.Pool, logger)
	ratingSvc := rating.NewService(postgres.Pool, memManager, logger)
	repoSvc := repo.NewService(
		postgres.Pool, embedder, redisClient.Client,
		cfg.Security.EncryptionKey,
		cfg.OAuth.GitHubClientID, cfg.OAuth.GitHubClientSecret,
		cfg.OAuth.GitLabClientID, cfg.OAuth.GitLabClientSecret,
		cfg.OAuth.BaseURL, cfg.OAuth.FrontendURL,
		logger,
	)

	// Initialize HTTP handlers
	projectHandler := project.NewHandler(projectSvc, logger)
	chatHandler := chat.NewHandler(chatSvc, logger)
	messageHandler := message.NewHandler(postgres.Pool, chatSvc, orch, modelGateway, embedder, memManager, provSvc, tenantSvc, logger)
	ratingHandler := rating.NewHandler(ratingSvc, logger)
	expertHandler := expert.NewHandler(postgres.Pool, tenantSvc, logger)
	byoHandler := byoexpert.NewHandler(byoSvc, logger)
	explainSvc := explain.NewService(postgres.Pool, provSvc, tenantSvc, logger)
	explainHandler := explain.NewHandler(explainSvc, logger)
	// C10: reliability-as-product — published SLO surface + audit-grade event
	// log. Probes reuse the existing dependency HealthCheck methods.
	relSvc := reliability.NewService(postgres.Pool, reliability.Policy{
		Enabled:               cfg.Reliability.Enabled,
		AvailabilityTarget:    cfg.Reliability.AvailabilityTarget,
		ErrorBudgetWindowDays: cfg.Reliability.ErrorBudgetWindowDays,
		AtRiskThreshold:       cfg.Reliability.AtRiskThreshold,
	}, logger)
	relSvc.RegisterProbe("postgres", postgres.HealthCheck)
	relSvc.RegisterProbe("redis", redisClient.HealthCheck)
	relSvc.RegisterProbe("ml_sidecar", mlClient.HealthCheck)
	relHandler := reliability.NewHandler(relSvc, "1.0.0", logger)
	repoHandler := repo.NewHandler(repoSvc, tenantSvc, logger)
	adminHandler := adminpkg.NewAdminHandler(postgres.Pool, modelGateway, mlClient, embedder, categoryRegistry, domainRegistry, versionSvc, evalStore, tenantSvc, usageSvc, freshnessSvc, byoSvc, eventsStore, docExtractor, logger)

	// Collaboration layer (Phase C + D)
	bbStore := blackboard.NewStore(postgres.Pool, redisClient.Client, logger)
	bbStore.SetProvenanceRecorder(provSvc) // C1: signed chain per artifact
	bbSubscriber := blackboard.NewSubscriber(bbStore, redisClient.Client, logger)
	wfEngine := workflow.NewEngine(postgres.Pool, logger)
	// A8: let the engine announce terminal states on the blackboard, which is
	// what unlocks the kanban/files SSE "done" events (kanban_sse.go:90).
	wfEngine.SetEventPoster(bbStore)
	validationPipeline := validation.NewPipeline(modelGateway, logger)
	wfTools := workflow.NewTools(bbStore, wfEngine, bbSubscriber, validationPipeline, logger)
	// WorkflowRunner: drives workflows from start to completion.
	wfPlanner := workflow.NewPlanner(modelGateway, logger)
	wfAgentLoop := workflow.NewAgentLoop(postgres.Pool, wfTools, bbStore, modelGateway, contextAssembler, memManager, logger)
	// AiderRunner: executes implementation/qa phases using Aider.
	// WHY separate from AgentLoop: File system as context, git history as memory.
	// Workspace root: configurable via AIDER_WORKSPACE_ROOT env var.
	// Default: /tmp/ai_avengers_workspaces (dev convenience, auto-cleanup on reboot)
	// Production: set AIDER_WORKSPACE_ROOT=/data/workspaces
	// Docker: set AIDER_WORKSPACE_ROOT=/workspaces (volume mount)
	workspaceRoot := os.Getenv("AIDER_WORKSPACE_ROOT")
	if workspaceRoot == "" {
		workspaceRoot = "/tmp/ai_avengers_workspaces"
	}
	aiderServiceURL := os.Getenv("AIDER_SERVICE_URL")
	if aiderServiceURL == "" {
		aiderServiceURL = "http://localhost:8082"
	}
	wfAiderRunner := workflow.NewAiderRunner(postgres.Pool, bbStore, modelGateway, mlClient, validationPipeline, workspaceRoot, aiderServiceURL, logger)
	wfWorkspaceMerger := workflow.NewWorkspaceMerger(logger)
	logger.Info("aider workspace configured", zap.String("root", workspaceRoot))
	wfCrossVerifier := workflow.NewCrossVerifier(bbStore, modelGateway, wfAgentLoop, wfAiderRunner, wfTools, logger)
	// C7: adversarial debate overlay on high-stakes artifacts (attack→defend
	// →verdict, hop-bound). Kill switch = DEBATE_ENABLED=false.
	wfCrossVerifier.SetDebatePolicy(workflow.DebatePolicy{
		Enabled: cfg.Debate.Enabled,
		MaxHops: cfg.Debate.MaxHops,
	})

	// wfSections moved up from where it used to be constructed (previously only
	// needed by the chat wiring below) because AuthoringRunner needs it too, and
	// AuthoringRunner must exist before wfRunner does.
	wfSections := workflow.NewDesignSectionStore(postgres.Pool, logger)
	// AuthoringRunner (§9): the implementation phase now authors design
	// sections through this, not application code through wfAiderRunner
	// directly. wfAiderRunner is passed in and reused for its HTTP-call
	// plumbing, workspace conventions and git helpers — see authoring.go's
	// file comment for exactly which methods are shared and why duplicating
	// them would risk drift the next time one side gets a fix.
	wfAuthoringRunner := workflow.NewAuthoringRunner(wfAiderRunner, wfSections, logger)

	// CostMonitor: enforces per-workflow budget caps (soft + hard limits).
	// monthly_limit_usd and alert_threshold read from system_settings at startup.
	// Defaults: $1000/month, 80% alert threshold (architecture doc §17).
	wfCostMonitor := monitoring.NewCostMonitor(
		postgres.Pool,
		modelGateway,
		1000.0, // monthly_limit_usd
		0.8,    // alert_threshold (80%)
		logger,
	)
	// GateSystem holds no mutable state (assembler + logger only)
	wfGateSystem := workflow.NewGateSystem(contextAssembler, postgres.Pool, logger)

	wfQARunner := workflow.NewQARunner(
		postgres.Pool, bbStore, wfGateSystem, modelGateway, wfSections, workspaceRoot, logger,
	)
	wfRunner := workflow.NewWorkflowRunner(
		postgres.Pool, wfEngine, wfPlanner, wfAgentLoop, wfAiderRunner, wfAuthoringRunner, wfQARunner, wfWorkspaceMerger, wfCrossVerifier,
		wfCostMonitor, wfTools, bbStore, modelGateway, logger,
	)
	// Projector: blackboard events -> workflow_tasks projection (single write path).
	wfProjector := workflow.NewProjector(postgres.Pool, bbStore, bbSubscriber, logger)
	wfHandler := workflow.NewHandler(wfEngine, bbStore, redisClient.Client, wfProjector, logger)

	// Phase 3D: the human-approved working set for existing-codebase workflows.
	// repoSvc is the repository index; codebaseSvc owns the approval rules.
	wfCodebaseSvc := workflow.NewCodebaseService(postgres.Pool, bbStore, repoSvc, logger)
	// 3F needs the workspace root to diff the workflow's work against its
	// baseline; without it, patch generation reports itself unavailable.
	wfCodebaseSvc.SetWorkspaceRoot(workspaceRoot)
	wfCodebaseHandler := workflow.NewCodebaseHandler(wfCodebaseSvc, postgres.Pool, logger)
	// 3E: the runner seeds only the approved working set into the workspace and
	// keeps unapproved repository paths out of the merge.
	wfRunner.SetCodebaseWorkspace(wfCodebaseSvc)

	// Workflow chat (docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md §6).
	//
	// Entirely separate from the product chat (chatHandler / messageHandler
	// above): its own tables, its own handler, no shared code path. It reuses the
	// workflow engine's own pieces instead — GateSystem so an answer about a
	// deliverable obeys the same knowledge rules that produced it, and
	// DesignSectionStore so the design file list it shows is always the real one.
	//
	// A second GateSystem instance is constructed here rather than reaching into
	// AgentLoop's: GateSystem holds no mutable state (assembler + logger only),
	// and AgentLoop keeps its own private one (agent_loop.go:63).
	// wfSections constructed above (with wfAuthoringRunner) — reused here, not
	// rebuilt: DesignSectionStore has no state that would make a second
	// instance wrong, but ListSections/AssignSection must agree on the same
	// rows, so one instance shared everywhere is the simpler invariant to keep.
	// wfToolRegistry: the §7 tool loop's catalogue. workspaceRoot matches
	// AiderRunner's below exactly — read_design/search_design read the same
	// {workspaceRoot}/{workflowID}/main/ tree AiderRunner's own workspace
	// convention already defines.
	wfToolRegistry := workflow.NewToolRegistry(workspaceRoot, logger)
	wfChatSvc := workflow.NewWorkflowChatService(
		postgres.Pool, bbStore, wfGateSystem, modelGateway, wfSections,
		wfToolRegistry, workspaceRoot, logger,
	)
	wfChatHandler := workflow.NewChatHandler(wfChatSvc, logger)

	// Change requests (§6/§7 redesign): client free-text goals from workflow chat
	// become change_requests rows + blackboard events; the runner watches and
	// re-runs the relevant design waves. Constructed once and injected into both
	// the chat route mount and the runner — either side alone leaves the feature
	// half-dead (route without watcher, or watcher without a way to create).
	// MUST be attached before ResumeOrphanWorkflows below so resumed runs also
	// start the watcher (WithChangeRequestService sets r.crSvc).
	wfChangeReqSvc := workflow.NewChangeRequestService(postgres.Pool, bbStore, logger)
	wfRunner.WithChangeRequestService(wfChangeReqSvc)

	// Delivery: harness git-push export (§18) and the code-feedback loop (§17).
	//
	// Both take repoSvc as their credential source, through the one-method
	// RepoTokenSource interface rather than as a *repo.Service — so the
	// encryption key and the decrypt routine stay private to internal/repo,
	// which is the property migration 001 stored the token encrypted to
	// protect. See internal/workflow/client_repo.go for the credential rules
	// these two features are bound by.
	//
	// workspaceRoot is the same value AiderRunner and the tool registry get:
	// all three read {workspaceRoot}/{workflowID}/main/, and the export pushes
	// exactly the tree the chat's read_design shows.
	wfGitExporter := workflow.NewGitExporter(postgres.Pool, bbStore, repoSvc, workspaceRoot, logger)
	wfCodeFeedback := workflow.NewCodeFeedbackService(
		postgres.Pool, bbStore, wfSections, repoSvc, workspaceRoot, logger,
	)
	wfDeliveryHandler := workflow.NewDeliveryHandler(wfGitExporter, wfCodeFeedback, logger)

	// Amendments (§7.5): the write-back path for everything the chat's mutating
	// tools propose. Shares wfSections (the amendable-file allowlist is derived
	// from the assigned sections) and the same workspaceRoot, because an approved
	// amendment is committed into the very tree the export pushes.
	wfAmendments := workflow.NewAmendmentService(postgres.Pool, bbStore, wfSections, workspaceRoot, logger)
	wfAmendmentHandler := workflow.NewAmendmentHandler(wfAmendments, logger)

	// Resume any workflows that were running before pod restart.
	// WHY background context: must outlive the HTTP server startup.
	go wfRunner.ResumeOrphanWorkflows(context.Background())

	// ============================================================
	// Health check — no auth required
	// ============================================================
	router.GET("/health", func(c *gin.Context) {
		health := map[string]string{"status": "ok", "version": "1.0.0"}
		if err := postgres.HealthCheck(c.Request.Context()); err != nil {
			health["postgres"] = "unhealthy"
			health["status"] = "degraded"
		} else {
			health["postgres"] = "healthy"
		}
		if err := redisClient.HealthCheck(c.Request.Context()); err != nil {
			health["redis"] = "unhealthy"
			health["status"] = "degraded"
		} else {
			health["redis"] = "healthy"
		}
		if err := mlClient.HealthCheck(c.Request.Context()); err != nil {
			health["ml_sidecar"] = "unhealthy"
		} else {
			health["ml_sidecar"] = "healthy"
		}
		response.OK(c, health)
	})

	// GET /metrics — in-process counters (JSON). No Prometheus client dep.
	// See FUTURE_UPDATES.md Item 1 + internal/observability/metrics.go.
	router.GET("/metrics", func(c *gin.Context) {
		response.OK(c, observability.Global.Snapshot())
	})

	// GET /status — public reliability surface (C10): dependency components +
	// the published SLO snapshot. No auth; probe error strings are withheld
	// publicly (the admin view carries them).
	router.GET("/status", relHandler.PublicStatus)

	v1 := router.Group("/api/v1")

	// ============================================================
	// Auth routes — no JWT required
	// ============================================================
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", handleRegister(authService))
		authGroup.POST("/login", handleLogin(authService))
		authGroup.POST("/admin/login", handleAdminLogin(authService))
		// Hidden admin bootstrap — requires unguessable one-time token.
		// Not linked from public login/register pages.
		authGroup.POST("/admin/bootstrap/start", handleBootstrapStart(authService))
		authGroup.POST("/admin/bootstrap/complete", handleBootstrapComplete(authService))
		authGroup.POST("/refresh", handleRefresh(jwtService))
		authGroup.POST("/logout", middleware.AuthMiddleware(jwtService, logger), handleLogout(jwtService))
		authGroup.POST("/forgot-password", handleForgotPassword(authService, logger))
		authGroup.POST("/reset-password", handleResetPassword(authService))
	}

	// OAuth callback — no JWT (browser redirect from GitHub/GitLab)
	// WHY public: OAuth provider redirects browser here with ?code=...
	// The browser has no JWT at this point — it's a fresh redirect.
	v1.GET("/repo/callback/:provider", repoHandler.OAuthCallback)

	// ============================================================
	// Service-to-service routes — shared secret, NOT a user JWT
	// ============================================================
	// LLM proxy for AiderService: routes Aider's LLM calls through
	// ModelGateway so cost, retry and provider selection stay centralized.
	//
	// WHY not on the JWT-protected group (where it used to live):
	// AiderService is a process, not a logged-in user. It cannot obtain an
	// access token, so every Aider LLM call was rejected with 401 before it
	// reached the gateway. A shared secret is the credential a service can
	// actually hold.
	//
	// Two paths, one handler: Aider's OPENAI_API_BASE points at /llm/proxy,
	// and the OpenAI client library inside litellm appends
	// "/chat/completions" to its base URL. The plain path stays for direct
	// callers and for probing the endpoint by hand.
	aiderProxyToken := os.Getenv("AIDER_PROXY_TOKEN")
	if aiderProxyToken == "" {
		logger.Warn("AIDER_PROXY_TOKEN is not set — /llm/proxy will reject all calls with 503, Aider phases cannot run")
	}
	llmProxy := v1.Group("/llm")
	llmProxy.Use(middleware.ServiceTokenMiddleware(aiderProxyToken, logger))
	{
		llmProxy.POST("/proxy", modelGateway.ProxyHandler)
		llmProxy.POST("/proxy/chat/completions", modelGateway.ProxyHandler)
	}

	// ============================================================
	// Protected routes — JWT required
	// ============================================================
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService, logger))
	{
		// GET /auth/me — session restore after page reload
		// WHY needed: access token is in-memory only (XSS mitigation).
		// After hard reload, frontend has no user data. This endpoint
		// lets it restore fullName + email + role from the server
		// using the refresh token cookie.
		protected.GET("/auth/me", handleGetMe(authService))

		// Authenticated admin self-service TOTP management
		protected.GET("/auth/totp", handleGetTOTPStatus(authService))
		protected.POST("/auth/totp/setup", handleSetupTOTP(authService))
		protected.POST("/auth/totp/enable", handleEnableTOTP(authService))
		protected.POST("/auth/totp/disable", handleDisableTOTP(authService))

		// Expert routes (read-only for clients / domain experts)
		experts := protected.Group("/experts")
		{
			experts.GET("", expertHandler.ListActive)
			experts.GET("/:id", expertHandler.GetByID)
			experts.GET("/:id/topics", expertHandler.GetTopics)
		}
		// C8: bring-your-own-expert (tenant self-service). Static segments
		// only (no wildcard) so it cannot collide with /experts/:id.
		byoHandler.RegisterRoutes(protected)
		// Project routes
		projects := protected.Group("/projects")
		{
			projects.POST("", projectHandler.Create)
			projects.GET("", projectHandler.List)
			projects.GET("/:id", projectHandler.GetByID)
			projects.PATCH("/:id", projectHandler.Update)
			projects.DELETE("/:id", projectHandler.Delete)

			projects.POST("/:id/experts", projectHandler.AddExpert)
			projects.DELETE("/:id/experts/:expertId", projectHandler.RemoveExpert)
			projects.POST("/:id/chats", projectHandler.CreateChat)
			projects.GET("/:id/chats", projectHandler.ListChats)
			projects.GET("/:id/memory", handleGetProjectMemory(memManager))
			projects.GET("/:id/timeline", handleGetProjectTimeline(memManager))

			// Repo integration
			projects.POST("/:id/repo", repoHandler.ConnectRepo)
			projects.POST("/:id/repo/sync", repoHandler.SyncRepo)
			projects.GET("/:id/repo/status", repoHandler.GetSyncStatus)
			// Phase 3A: read-only access to the stored file tree + file bodies.
			projects.GET("/:id/repo/tree", repoHandler.ListRepoTree)
			projects.GET("/:id/repo/file", repoHandler.GetRepoFile)
			// Phase 3B: dependency neighbourhood + requirement-driven file ranking.
			projects.GET("/:id/repo/graph", repoHandler.GetRepoGraph)
			projects.GET("/:id/repo/suggest", repoHandler.SuggestRepoFiles)
		}

		// OAuth initiation — protected so we know which user is connecting
		// WHY protected: we need clientID to associate the connection
		protected.GET("/repo/oauth/:provider", repoHandler.GetOAuthURL)

		// Chat routes
		chats := protected.Group("/chats")
		{
			chats.GET("/:id", chatHandler.GetByID)
			chats.PATCH("/:id", chatHandler.Update)
			chats.DELETE("/:id", chatHandler.Archive)
			chats.POST("/:id/unarchive", chatHandler.Unarchive)
			chats.DELETE("/:id/permanent", chatHandler.PermanentDelete)
			chats.POST("/:id/messages", messageHandler.Send)
			chats.GET("/:id/messages", chatHandler.ListMessages)
		}

		// Message routes
		messages := protected.Group("/messages")
		{
			messages.POST("/:id/rate", ratingHandler.Rate)
			messages.DELETE("/:id", messageHandler.DeleteMessage)
			messages.PATCH("/:id", messageHandler.UpdateMessage)
			// C1: signed provenance chain for an answer.
			messages.GET("/:id/provenance", handleGetMessageProvenance(provSvc))
			// C9: unified "why this answer" view (view over stored facts).
			messages.GET("/:id/explanation", explainHandler.Get)
		}

		// Global search (#27) — cross-project. Mounted on its own prefix so it
		// can never collide with the /chats/:id wildcard.
		search := protected.Group("/search")
		{
			search.GET("/chats", handleSearchChats(chatSvc, tenantSvc))
		}

		// Workflow routes (Phase C — collaboration layer)
		workflows := protected.Group("/workflows")
		{
			workflows.GET("", wfHandler.ListWorkflows)
			workflows.POST("", wfHandler.CreateWorkflow)
			workflows.GET("/:id", wfHandler.GetWorkflow)
			workflows.POST("/:id/start", wfHandler.StartWorkflow)
			workflows.GET("/:id/blackboard", wfHandler.GetBlackboard)
			// C1: signed provenance chain for one artifact event.
			workflows.GET("/:id/artifacts/:eventId/provenance", handleGetArtifactProvenance(provSvc))
			workflows.GET("/:id/kanban", wfHandler.GetKanban)
			workflows.POST("/:id/run", wfHandler.RunWorkflow(wfRunner))
			workflows.GET("/:id/kanban/stream", wfHandler.StreamKanban)
			workflows.GET("/:id/files/stream", wfHandler.StreamFiles)
			workflows.POST("/:id/approvals/:aid/respond", wfHandler.RespondToApproval)
			workflows.POST("/:id/cancel", wfHandler.CancelWorkflow)
			workflows.POST("/:id/tasks/:taskId/retry", wfHandler.RetryTask)

			// Phase 3D: the human-approved working set. Every route is behind
			// the client-ownership check in CodebaseHandler, not the tenant
			// guard: an approval is this client's private decision.
			workflows.GET("/:id/codebase/files", wfCodebaseHandler.ListFiles)
			workflows.POST("/:id/codebase/suggest", wfCodebaseHandler.Suggest)
			workflows.POST("/:id/codebase/files", wfCodebaseHandler.AddFile)
			workflows.POST("/:id/codebase/files/decide-bulk", wfCodebaseHandler.DecideBulk)
			workflows.POST("/:id/codebase/files/:fileId/decide", wfCodebaseHandler.Decide)
			workflows.GET("/:id/codebase/manifest", wfCodebaseHandler.Manifest)
			// Phase 3F: deliver the workflow's changes as a patch against the
			// pinned base. Nothing here pushes to the client's repository.
			workflows.POST("/:id/codebase/patch", wfCodebaseHandler.GeneratePatch)
			workflows.GET("/:id/codebase/patch", wfCodebaseHandler.GetPatch)
			workflows.GET("/:id/codebase/patch/download", wfCodebaseHandler.DownloadPatch)
		}

		// Workflow chat (§6). Registered on `protected`, so these inherit
		// AuthMiddleware like every route above — the handlers read
		// c.MustGet("user_id") and every one of them checks chat ownership
		// against it.
		//
		// Mounts /workflows/:id/chats and /workflow-chats/:cid/* plus
		// POST /workflow-chats/:cid/propose-change (RegisterRoutesWithCR).
		// The workflow paths reuse the `:id` parameter name the group above
		// already uses, so gin's tree merges them instead of reporting a
		// wildcard conflict. wfChangeReqSvc was constructed next to the chat
		// handler above and is also attached to wfRunner.
		wfChatHandler.RegisterRoutesWithCR(protected, wfChangeReqSvc)

		// Delivery (§17, §18). Same `protected` group and the same `:id`
		// parameter name for the same two reasons: both handlers check workflow
		// ownership against c.MustGet("user_id"), and a different wildcard name
		// at this path position would conflict with the workflow routes above.
		wfDeliveryHandler.RegisterRoutes(protected)

		// Amendments (§7.5). Separate from the approvals route above on purpose
		// — see amendment_handler.go's file comment for the three reasons that
		// endpoint cannot be reused (no edited-content field, it resumes the
		// workflow, and there is no list endpoint).
		wfAmendmentHandler.RegisterRoutes(protected)
	}

	// ============================================================
	// Admin routes — JWT + admin role required
	// ============================================================
	adminGroup := v1.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(jwtService, logger))
	adminGroup.Use(middleware.AdminMiddleware())
	{
		adminGroup.GET("/experts", adminHandler.ListExperts)
		adminGroup.POST("/experts", adminHandler.CreateExpert)
		adminGroup.PATCH("/experts/:id", adminHandler.UpdateExpert)
		adminGroup.POST("/experts/:id/ingest", adminHandler.IngestTranscript)
		adminGroup.POST("/experts/:id/regenerate-charter", adminHandler.RegenerateCharter)
		// C2: expert versioning + capability drift.
		adminGroup.GET("/experts/:id/versions", adminHandler.ListExpertVersions)
		adminGroup.POST("/experts/:id/versions/snapshot", adminHandler.SnapshotExpertVersion)
		adminGroup.POST("/experts/:id/versions/:versionId/pin", adminHandler.PinExpertVersion)
		adminGroup.GET("/experts/:id/drift", adminHandler.ListExpertDrift)
		adminGroup.POST("/experts/:id/drift/:driftId/ack", adminHandler.AcknowledgeExpertDrift)
		// C3: evaluation harness run view + baseline promote.
		adminGroup.GET("/evals/runs", adminHandler.GetEvalRuns)
		adminGroup.POST("/evals/runs/:id/baseline", adminHandler.PromoteEvalBaseline)
		// C4: tenant isolation & enterprise controls.
		adminGroup.GET("/tenants", adminHandler.ListTenants)
		adminGroup.POST("/tenants", adminHandler.CreateTenant)
		adminGroup.POST("/tenants/:id/users", adminHandler.AssignTenantUser)
		adminGroup.POST("/tenants/:id/experts", adminHandler.AssignTenantExpert)
		// C8: tenant self-service ("bring your own") experts.
		adminGroup.POST("/tenants/:id/entitlement", adminHandler.SetTenantEntitlement)
		adminGroup.GET("/byo/events", adminHandler.ListByoEvents)
		// C10: reliability-as-product (full detail + audit trail).
		adminGroup.GET("/reliability/status", relHandler.AdminStatus)
		adminGroup.GET("/reliability/events", relHandler.ListEvents)
		// C5: cost & usage analytics product.
		adminGroup.GET("/usage", adminHandler.GetUsage)
		adminGroup.GET("/usage/budgets", adminHandler.GetUsageBudgets)
		adminGroup.PUT("/usage/budgets", adminHandler.SetUsageBudget)
		adminGroup.GET("/usage/alerts", adminHandler.GetUsageAlerts)
		// C6: knowledge freshness / staleness + refresh tasks.
		adminGroup.GET("/freshness/tasks", adminHandler.ListFreshnessTasks)
		adminGroup.POST("/freshness/scan", adminHandler.ScanAllFreshness)
		adminGroup.POST("/freshness/tasks/:taskId/ack", adminHandler.AcknowledgeFreshnessTask)
		adminGroup.POST("/freshness/tasks/:taskId/resolve", adminHandler.ResolveFreshnessTask)
		adminGroup.GET("/experts/:id/freshness", adminHandler.GetExpertFreshness)
		adminGroup.POST("/experts/:id/freshness/scan", adminHandler.ScanExpertFreshness)
		adminGroup.GET("/experts/:id/jobs", adminHandler.GetIngestionJobs)
		adminGroup.GET("/experts/:id/jobs/stream", adminHandler.StreamIngestionJob)
		adminGroup.GET("/experts/:id/jobs/events", adminHandler.GetIngestionJobEvents)
		adminGroup.POST("/experts/:id/jobs/:jobID/resume", adminHandler.ResumeIngestionJob)
		adminGroup.POST("/experts/:id/jobs/:jobID/retry", adminHandler.RetryIngestionJob)
		adminGroup.GET("/clients", adminHandler.ListClients)
		adminGroup.PATCH("/clients/:id", adminHandler.UpdateClient)
		// Managed accounts: admin + domain_expert CRUD + expert grants
		adminGroup.GET("/accounts", handleListManagedAccounts(authService))
		adminGroup.POST("/accounts", handleCreateManagedAccount(authService))
		adminGroup.PATCH("/accounts/:id", handleUpdateManagedAccount(authService))
		adminGroup.DELETE("/accounts/:id", handleDeleteManagedAccount(authService))
		adminGroup.PUT("/accounts/:id/experts", handleSetAccountExperts(authService))
		adminGroup.POST("/bootstrap-tokens", handleIssueBootstrapToken(authService, logger))
		adminGroup.GET("/stats", adminHandler.GetStats)
		adminGroup.GET("/violations", adminHandler.GetViolations)
		adminGroup.GET("/ratings", adminHandler.GetRatings)
		adminGroup.GET("/settings", adminHandler.GetSettings)
		adminGroup.PATCH("/settings/:key", adminHandler.UpdateSetting)
		// LLM provider + API key management (no env file needed)
		adminGroup.GET("/llm-settings", adminHandler.GetLLMSettings)
		adminGroup.POST("/llm-settings", adminHandler.UpdateLLMSettings)
		// CodeCraftAPI model catalog proxy + embedding settings
		// WHY proxy: API key must never leave the server.
		adminGroup.GET("/codecraftapi/models", adminHandler.GetCodeCraftModels)
		adminGroup.GET("/embedding-settings", adminHandler.GetEmbeddingSettings)
		adminGroup.POST("/embedding-settings", adminHandler.UpdateEmbeddingSettings)
		// Expert categories (migration 010, CT-A3) — admin-owned template layer.
		adminGroup.GET("/expert-categories", adminHandler.ListExpertCategories)
		adminGroup.POST("/expert-categories", adminHandler.CreateExpertCategory)
		adminGroup.GET("/expert-categories/:id", adminHandler.GetExpertCategory)
		adminGroup.PATCH("/expert-categories/:id", adminHandler.UpdateExpertCategory)
		// Domain profiles (China Wall per-domain config) — admin-configurable,
		// no redeploy needed. Includes MaxTokensFlat/MaxTokensStructured
		// (2026-09-08 addition, see chinawall/domain_profile.go's field docs).
		adminGroup.GET("/domain-profiles", adminHandler.ListDomainProfiles)
		adminGroup.GET("/domain-profiles/:domain", adminHandler.GetDomainProfile)
		adminGroup.PATCH("/domain-profiles/:domain", adminHandler.UpdateDomainProfile)
		// Gate 1 thresholds (B4) — per-domain usable/strong config.
		// Calibrate writes proposals only; Apply / Set make them live.
		adminGroup.GET("/gate-thresholds", adminHandler.ListGateThresholds)
		adminGroup.POST("/gate-thresholds/calibrate", adminHandler.CalibrateGateThresholds)
		adminGroup.POST("/gate-thresholds/:domain/apply", adminHandler.ApplyGateThreshold)
		adminGroup.PATCH("/gate-thresholds/:domain", adminHandler.SetGateThreshold)
	}

	return router
}

// ============================================================
// Inline route handlers
// ============================================================

// handleSearchChats GET /search/chats?q=&limit= (#27)
// Global, cross-project search over the caller's own chats (title + message
// content). Tenant-aware via the C4 scope — a tenant caller only ever sees
// chats whose project belongs to their tenant.
//
// WHY a closure (not a chat.Handler method): the query needs both the chat
// service and the tenant service, and main.go already builds every
// cross-cutting handler this way (provenance, explanation, project memory).
func handleSearchChats(chatSvc *chat.Service, tenantSvc *tenant.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID, ok := c.MustGet("user_id").(uuid.UUID)
		if !ok {
			response.Unauthorized(c, "invalid session")
			return
		}
		q := strings.TrimSpace(c.Query("q"))
		if len([]rune(q)) < 2 {
			response.BadRequest(c, "QUERY_TOO_SHORT", "q must be at least 2 characters")
			return
		}
		limit := 20
		if l := strings.TrimSpace(c.Query("limit")); l != "" {
			if n, err := strconv.Atoi(l); err == nil {
				limit = n
			}
		}

		// C4: resolve the caller's tenant scope. nil tenant → global (admin
		// or isolation disabled), in which case the SQL predicate is skipped.
		role, _ := c.Get("role")
		roleStr, _ := role.(string)
		scope, err := tenantSvc.Resolve(c.Request.Context(), clientID, roleStr)
		if err != nil {
			response.Forbidden(c, "Tenant scope could not be resolved")
			return
		}

		results, err := chatSvc.SearchChats(c.Request.Context(), clientID, scope.TenantID, q, limit)
		if err != nil {
			response.InternalError(c)
			return
		}
		response.OK(c, gin.H{"query": q, "results": results})
	}
}

// handleGetProjectMemory GET /projects/:id/memory
// Returns L2 group-memory entries (cross-expert decisions).
//
// Feature #5 fix (docs bug list): this handler previously called
// memManager.GetTimeline (L3 event log) - the SAME method
// handleGetProjectTimeline below calls, just with a different limit.
// It never queried L2 (project_memory_l2) at all. Repointed to the
// real L2 fetch, memManager.GetProjectMemory (memory/manager.go).
func handleGetProjectMemory(memManager *memory.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid project ID")
			return
		}
		entries, err := memManager.GetProjectMemory(c.Request.Context(), projectID, 20)
		if err != nil {
			response.InternalError(c)
			return
		}
		response.OK(c, entries)
	}
}

// handleGetProjectTimeline GET /projects/:id/timeline
// Returns L3 master event log for a project (paginated, first 50).
func handleGetProjectTimeline(memManager *memory.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid project ID")
			return
		}
		events, err := memManager.GetTimeline(c.Request.Context(), projectID, 50, 0)
		if err != nil {
			response.InternalError(c)
			return
		}
		response.OK(c, events)
	}
}

// handleGetMessageProvenance GET /messages/:id/provenance (C1)
// Returns the stored signed provenance chain for an answer, plus a
// server-side signature verification result so the client can trust it.
func handleGetMessageProvenance(provSvc *provenance.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		messageID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid message ID")
			return
		}
		rec, err := provSvc.Get(c.Request.Context(), provenance.OutputChatMessage, messageID)
		if err != nil {
			response.InternalError(c)
			return
		}
		if rec == nil {
			response.NotFound(c, "provenance")
			return
		}
		response.OK(c, gin.H{"record": rec, "verified": provSvc.Verify(rec)})
	}
}

// handleGetArtifactProvenance GET /workflows/:id/artifacts/:eventId/provenance (C1)
// Returns the stored signed provenance chain for one artifact event
// (the :id workflow segment is validated against the record for safety).
func handleGetArtifactProvenance(provSvc *provenance.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		workflowID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
			return
		}
		eventID, err := uuid.Parse(c.Param("eventId"))
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid event ID")
			return
		}
		rec, err := provSvc.Get(c.Request.Context(), provenance.OutputWorkflowArtifact, eventID)
		if err != nil {
			response.InternalError(c)
			return
		}
		if rec == nil || rec.WorkflowID == nil || *rec.WorkflowID != workflowID {
			response.NotFound(c, "provenance")
			return
		}
		response.OK(c, gin.H{"record": rec, "verified": provSvc.Verify(rec)})
	}
}

// handleGetMe GET /auth/me
// Returns the authenticated user's profile.
// WHY needed: access token is in-memory only (XSS mitigation).
// After hard page reload, frontend has no user data.
// This endpoint restores fullName + email + role from the server.
func handleGetMe(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uuid.UUID)
		user, err := svc.GetMe(c.Request.Context(), userID)
		if err != nil {
			response.Unauthorized(c, "User not found")
			return
		}
		// camelCase to match openapi.yaml User schema - see buildAuthResponse's
		// comment above for why this is safe to change on the response side alone.
		response.OK(c, map[string]interface{}{
			"id":          user.ID,
			"email":       user.Email,
			"fullName":    user.FullName,
			"role":        user.Role,
			"totpEnabled": user.TOTPEnabled,
		})
	}
}

// ============================================================
// Auth helpers
// ============================================================

// buildAuthResponse builds the response structure the frontend expects.
// Frontend api/auth.ts destructures: { user, tokenPair }
// Frontend TokenPair type: { accessToken: string, expiresInSeconds: number }
// refreshTokenCookieName is the name of the httpOnly refresh token cookie.
const refreshTokenCookieName = "refresh_token"

// setRefreshCookie sets the refresh token as an httpOnly cookie.
//
// WHY httpOnly:
// JavaScript cannot read httpOnly cookies.
// XSS attack cannot steal the refresh token.
// This is the standard secure pattern for refresh tokens.
//
// WHY Path=/api/v1/auth/refresh:
// Cookie is only sent to the refresh endpoint.
// Not sent on every API call — reduces attack surface.
//
// WHY SameSite=Lax not Strict:
// Strict blocks cookie on top-level navigation from external links.
// Lax allows it. Correct for a web app.
func setRefreshCookie(c *gin.Context, refreshToken string, expiryDays int) {
	maxAge := expiryDays * 24 * 60 * 60 // seconds

	// WHY Path=/ not /api/v1/auth/refresh:
	//   Restricted path caused cookie to not be sent on cross-origin
	//   requests from 192.168.1.7:3000 -> 192.168.1.7:8080.
	//   Path=/ ensures browser sends cookie on all backend requests.
	//   Security is maintained by httpOnly (JS can't read) + token
	//   validation on the refresh endpoint.
	c.SetCookie(
		refreshTokenCookieName,
		refreshToken,
		maxAge,
		"/",   // Path — send on all requests to this domain
		"",    // Domain — empty = current domain only
		false, // Secure — false for HTTP local dev; set true in production
		true,  // HttpOnly — JS cannot read this cookie
	)

	// WHY manual SameSite header:
	//   Gin's c.SetCookie() does not expose SameSite parameter.
	//   We append it to the Set-Cookie header directly.
	//   SameSite=Lax: cookie sent on same-host requests (same IP,
	//   different port counts as same-site in most browsers).
	//   This fixes the 192.168.1.7:3000 -> :8080 cross-port scenario.
	existing := c.Writer.Header().Get("Set-Cookie")
	if existing != "" {
		c.Writer.Header().Set("Set-Cookie", existing+"; SameSite=Lax")
	}
}

// clearRefreshCookie removes the refresh token cookie on logout.
func clearRefreshCookie(c *gin.Context) {
	// Path must match setRefreshCookie — both use /
	c.SetCookie(refreshTokenCookieName, "", -1, "/", "", false, true)
}

// buildAuthResponse builds the response the frontend expects.
// Verified against:
//   - frontend/src/types/auth.ts User interface
//   - frontend/src/types/auth.ts TokenPair interface
//   - backend-go/api/openapi.yaml AuthResponse schema
//
// NOTE: refresh token is NOT in this response body.
// It is set as an httpOnly cookie by the caller before calling this.
// frontend/src/types/auth.ts explicitly documents this design.
func buildAuthResponse(user *auth.User, tokens *auth.TokenPair) map[string]interface{} {
	// WHY camelCase keys directly (not snake_case + frontend bridge):
	// Interface-First audit (docs/INTERFACE_FIRST_CONTRACT.md §4) found
	// this handler emitted snake_case while openapi.yaml's User/TokenPair
	// schemas are camelCase. Go now owns the wire casing directly - the
	// frontend's camelizeKeys() response interceptor is a safe no-op on
	// already-camelCase keys (utils/casing.ts fast-paths keys with no
	// underscore), so this is safe to change unilaterally on this side.
	return map[string]interface{}{
		"user": map[string]interface{}{
			"id":          user.ID,
			"email":       user.Email,
			"fullName":    user.FullName,
			"role":        user.Role,
			"totpEnabled": user.TOTPEnabled,
		},
		"tokenPair": map[string]interface{}{
			"accessToken":      tokens.AccessToken,
			"expiresInSeconds": int(time.Until(tokens.ExpiresAt).Seconds()),
		},
	}
}

// ============================================================
// Auth handlers
// ============================================================

func handleRegister(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Self-registration disabled: respond gracefully (endpoint stays alive).
		if !svc.SelfRegistrationEnabled() {
			response.Forbidden(c, "Self-registration is disabled. Please contact your administrator to get an account.")
			return
		}

		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=8"`
			FullName string `json:"full_name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "INVALID_INPUT", err.Error())
			return
		}

		tokens, err := svc.Register(c.Request.Context(), auth.RegisterRequest{
			Email:    req.Email,
			Password: req.Password,
			FullName: req.FullName,
		})
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrRegistrationDisabled):
				response.Forbidden(c, "Self-registration is disabled. Please contact your administrator to get an account.")
			case errors.Is(err, auth.ErrEmailTaken):
				response.Conflict(c, "Email already registered")
			default:
				response.InternalError(c)
			}
			return
		}

		// Fetch user record to return alongside tokens
		// Frontend expects: { user: {...}, tokenPair: { accessToken, expiresInSeconds } }
		user, err := svc.GetMe(c.Request.Context(), tokens.UserID)
		if err != nil {
			response.InternalError(c)
			return
		}

		// Set refresh token as httpOnly cookie
		// WHY before response: cookie header must be set before body is written
		setRefreshCookie(c, tokens.RefreshToken, svc.RefreshExpiryDays())
		response.Created(c, buildAuthResponse(user, tokens))
	}
}

func handleLogin(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "INVALID_INPUT", err.Error())
			return
		}

		tokens, err := svc.Login(c.Request.Context(), auth.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrInvalidCredentials), errors.Is(err, auth.ErrUserNotFound):
				response.Unauthorized(c, "Invalid email or password")
			case errors.Is(err, auth.ErrUseAdminLogin):
				response.BadRequest(c, "USE_ADMIN_LOGIN", "Admin accounts must sign in via admin login with authenticator code")
			case errors.Is(err, auth.ErrUserInactive):
				response.Forbidden(c, "Account is disabled")
			default:
				response.InternalError(c)
			}
			return
		}

		user, err := svc.GetMe(c.Request.Context(), tokens.UserID)
		if err != nil {
			response.InternalError(c)
			return
		}

		setRefreshCookie(c, tokens.RefreshToken, svc.RefreshExpiryDays())
		response.OK(c, buildAuthResponse(user, tokens))
	}
}

func handleAdminLogin(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
			TOTPCode string `json:"totp_code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "INVALID_INPUT", err.Error())
			return
		}

		tokens, err := svc.AdminLogin(c.Request.Context(), auth.AdminLoginRequest{
			Email:    req.Email,
			Password: req.Password,
			TOTPCode: req.TOTPCode,
		})
		if err != nil {
			switch err {
			case auth.ErrInvalidCredentials, auth.ErrNotAdmin:
				response.Unauthorized(c, "Invalid credentials")
			case auth.ErrInvalidTOTP:
				response.Unauthorized(c, "Invalid TOTP code")
			case auth.ErrTOTPRequired:
				response.BadRequest(c, "TOTP_REQUIRED", "TOTP code required")
			case auth.ErrUserInactive:
				response.Forbidden(c, "Account is disabled")
			default:
				response.InternalError(c)
			}
			return
		}

		user, err := svc.GetMe(c.Request.Context(), tokens.UserID)
		if err != nil {
			response.InternalError(c)
			return
		}

		setRefreshCookie(c, tokens.RefreshToken, svc.RefreshExpiryDays())
		response.OK(c, buildAuthResponse(user, tokens))
	}
}

func handleRefresh(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read refresh token from httpOnly cookie (primary)
		// Fall back to request body (for non-browser clients / testing)
		//
		// WHY cookie first:
		// Browser sends cookie automatically via withCredentials:true.
		// refreshSession() in base.ts sends empty body {} — cookie carries the token.
		refreshToken, err := c.Cookie(refreshTokenCookieName)
		if err != nil || refreshToken == "" {
			// Cookie missing — try body (non-browser clients)
			var req struct {
				RefreshToken string `json:"refresh_token"`
			}
			_ = c.ShouldBindJSON(&req)
			refreshToken = req.RefreshToken
		}

		if refreshToken == "" {
			response.BadRequest(c, "MISSING_TOKEN", "refresh token required (cookie or body)")
			return
		}

		claims, err := jwtService.ValidateRefreshToken(c.Request.Context(), refreshToken)
		if err != nil {
			clearRefreshCookie(c)
			response.Unauthorized(c, "Invalid or expired refresh token")
			return
		}

		// WHY issue the NEW token pair BEFORE revoking the OLD one (this
		// order was previously reversed): if IssueTokenPair fails after
		// the old token was already revoked, the session is permanently
		// stranded - no valid refresh token exists anywhere (old
		// revoked, new never issued), and the browser's cookie still
		// holds the now-revoked old value, so every future refresh
		// attempt fails too. Issuing first means a failure here leaves
		// the OLD (still valid) token alone - the client can just retry.
		tokens, err := jwtService.IssueTokenPair(c.Request.Context(), claims.UserID, claims.Email, claims.Role)
		if err != nil {
			response.InternalError(c)
			return
		}

		// Revoke old refresh token with a 5-second grace period.
		// WHY grace period (not immediate delete):
		//   Race condition: if 3 requests 401 simultaneously, single-flight
		//   in base.ts ensures only 1 refresh call. But any request that
		//   was already in-flight with the OLD access token may arrive at
		//   the backend AFTER rotation. If the old refresh token is already
		//   gone from Redis, that in-flight request's retry (with new token)
		//   works fine — but if it somehow needs to re-validate the old
		//   refresh token, it would fail. 5s grace covers all in-flight
		//   requests without meaningful security impact.
		go func() {
			time.Sleep(5 * time.Second)
			_ = jwtService.RevokeRefreshToken(context.Background(), refreshToken)
		}()

		// Rotate cookie — new refresh token replaces old
		setRefreshCookie(c, tokens.RefreshToken, jwtService.RefreshExpiryDays())

		// Return only new access token — refresh token is in cookie.
		// camelCase to match openapi.yaml TokenPair schema. NOTE: this
		// endpoint's response is read by frontend/src/api/base.ts's
		// refreshSession() via a RAW axios call that bypasses the
		// camelizeKeys() interceptor entirely (by design, to avoid
		// interceptor re-entrancy on a 401 from refresh itself) - that
		// file was updated in the same commit as this change so the two
		// sides do not drift even for one commit.
		response.OK(c, map[string]interface{}{
			"accessToken":      tokens.AccessToken,
			"expiresInSeconds": int(time.Until(tokens.ExpiresAt).Seconds()),
		})
	}
}

func handleLogout(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Revoke refresh token from cookie
		if refreshToken, err := c.Cookie(refreshTokenCookieName); err == nil && refreshToken != "" {
			_ = jwtService.RevokeRefreshToken(c.Request.Context(), refreshToken)
		}
		// Also try body (non-browser clients)
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if _ = c.ShouldBindJSON(&req); req.RefreshToken != "" {
			_ = jwtService.RevokeRefreshToken(c.Request.Context(), req.RefreshToken)
		}

		clearRefreshCookie(c)
		response.OK(c, map[string]string{"message": "logged out"})
	}
}

// handleForgotPassword POST /auth/forgot-password
// Generates a password-reset token for the given email.
//
// WHY always 200: returning an error when the email is not found leaks
// whether an account exists (email enumeration attack). The raw token
// is logged server-side so an admin can relay it manually until an
// email integration is wired.
func handleForgotPassword(svc *auth.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "INVALID_INPUT", err.Error())
			return
		}

		rawToken, err := svc.ForgotPassword(c.Request.Context(), req.Email)
		if err != nil {
			logger.Error("forgot-password: service error",
				zap.String("email", req.Email),
				zap.Error(err),
			)
			response.InternalError(c)
			return
		}

		// Log the raw token so an admin can relay it manually.
		// TODO: replace with email delivery once SMTP is configured.
		if rawToken != "" {
			logger.Info("password reset token (relay to user manually until email is wired)",
				zap.String("email", req.Email),
				zap.String("reset_token", rawToken),
			)
		}

		// Always return the same response — do not reveal whether the email exists.
		response.OK(c, map[string]string{
			"message": "If that email is registered, a reset link has been sent.",
		})
	}
}

// handleResetPassword POST /auth/reset-password
// Validates a reset token and sets a new password.
func handleResetPassword(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Token       string `json:"token" binding:"required"`
			NewPassword string `json:"new_password" binding:"required,min=8"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "INVALID_INPUT", err.Error())
			return
		}

		if err := svc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
			switch err {
			case auth.ErrInvalidResetToken:
				response.BadRequest(c, "INVALID_TOKEN", "Reset token is invalid or has expired.")
			case auth.ErrPasswordTooShort:
				response.BadRequest(c, "PASSWORD_TOO_SHORT", err.Error())
			default:
				response.InternalError(c)
			}
			return
		}

		response.OK(c, map[string]string{
			"message": "Password reset successful. Please log in with your new password.",
		})
	}
}

// ============================================================
// Admin bootstrap + managed accounts + TOTP self-service
// ============================================================

func handleBootstrapStart(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Token    string `json:"token" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=8"`
			FullName string `json:"full_name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "INVALID_INPUT", err.Error())
			return
		}
		result, err := svc.BootstrapStart(c.Request.Context(), auth.BootstrapStartRequest{
			Token: req.Token, Email: req.Email, Password: req.Password, FullName: req.FullName,
		})
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrInvalidBootstrapToken):
				response.Unauthorized(c, "Invalid or expired bootstrap token")
			case errors.Is(err, auth.ErrEmailTaken):
				response.Conflict(c, "Email already registered")
			case errors.Is(err, auth.ErrPasswordTooShort):
				response.BadRequest(c, "PASSWORD_TOO_SHORT", err.Error())
			default:
				response.InternalError(c)
			}
			return
		}
		response.OK(c, result)
	}
}

func handleBootstrapComplete(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Token    string `json:"token" binding:"required"`
			TOTPCode string `json:"totp_code" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "INVALID_INPUT", err.Error())
			return
		}
		tokens, err := svc.BootstrapComplete(c.Request.Context(), auth.BootstrapCompleteRequest{
			Token: req.Token, TOTPCode: req.TOTPCode,
		})
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrInvalidBootstrapToken), errors.Is(err, auth.ErrBootstrapUnavailable):
				response.Unauthorized(c, "Invalid or expired bootstrap token")
			case errors.Is(err, auth.ErrInvalidTOTP):
				response.Unauthorized(c, "Invalid TOTP code")
			default:
				response.InternalError(c)
			}
			return
		}
		user, err := svc.GetMe(c.Request.Context(), tokens.UserID)
		if err != nil {
			response.InternalError(c)
			return
		}
		setRefreshCookie(c, tokens.RefreshToken, svc.RefreshExpiryDays())
		response.Created(c, buildAuthResponse(user, tokens))
	}
}

func handleIssueBootstrapToken(svc *auth.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := svc.IssueBootstrapToken(c.Request.Context(), 7*24*time.Hour)
		if err != nil {
			response.InternalError(c)
			return
		}
		logger.Info("admin bootstrap token issued via admin API")
		// Raw token returned once — caller must store it securely.
		response.Created(c, map[string]string{
			"token":   raw,
			"message": "One-time bootstrap token. Share out-of-band; not stored in plaintext.",
		})
	}
}

func handleListManagedAccounts(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		accounts, err := svc.ListManagedAccounts(c.Request.Context())
		if err != nil {
			response.InternalError(c)
			return
		}
		response.OK(c, accounts)
	}
}

func handleCreateManagedAccount(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		actorID := c.MustGet("user_id").(uuid.UUID)
		var req struct {
			Email     string   `json:"email" binding:"required,email"`
			Password  string   `json:"password" binding:"required,min=8"`
			FullName  string   `json:"full_name" binding:"required"`
			Role      string   `json:"role" binding:"required"`
			ExpertIDs []string `json:"expert_ids"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "INVALID_INPUT", err.Error())
			return
		}
		var expertIDs []uuid.UUID
		for _, s := range req.ExpertIDs {
			id, err := uuid.Parse(s)
			if err != nil {
				response.BadRequest(c, "INVALID_ID", "invalid expert_id")
				return
			}
			expertIDs = append(expertIDs, id)
		}
		acct, err := svc.CreateManagedAccount(c.Request.Context(), actorID, auth.CreateManagedAccountRequest{
			Email: req.Email, Password: req.Password, FullName: req.FullName,
			Role: req.Role, ExpertIDs: expertIDs,
		})
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrEmailTaken):
				response.Conflict(c, "Email already registered")
			case errors.Is(err, auth.ErrInvalidRole):
				response.BadRequest(c, "INVALID_ROLE", "role must be admin or domain_expert")
			case errors.Is(err, auth.ErrPasswordTooShort):
				response.BadRequest(c, "PASSWORD_TOO_SHORT", err.Error())
			default:
				response.InternalError(c)
			}
			return
		}
		response.Created(c, acct)
	}
}

func handleUpdateManagedAccount(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid account ID")
			return
		}
		// Partial update: any subset of fields. nil = unchanged.
		var req struct {
			FullName *string `json:"full_name"`
			Email    *string `json:"email"`
			Role     *string `json:"role"`
			IsActive *bool   `json:"is_active"`
			Password *string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "INVALID_INPUT", err.Error())
			return
		}
		err = svc.UpdateManagedAccount(c.Request.Context(), id, auth.UpdateManagedAccountRequest{
			FullName: req.FullName,
			Email:    req.Email,
			Role:     req.Role,
			IsActive: req.IsActive,
			Password: req.Password,
		})
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrUserNotFound):
				response.NotFound(c, "account")
			case errors.Is(err, auth.ErrEmailTaken):
				response.Conflict(c, "Email already registered")
			case errors.Is(err, auth.ErrInvalidRole):
				response.BadRequest(c, "INVALID_ROLE", "role must be admin, domain_expert or client")
			case errors.Is(err, auth.ErrPasswordTooShort):
				response.BadRequest(c, "PASSWORD_TOO_SHORT", err.Error())
			default:
				response.InternalError(c)
			}
			return
		}
		response.OK(c, map[string]string{"status": "updated"})
	}
}

func handleDeleteManagedAccount(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		actorID := c.MustGet("user_id").(uuid.UUID)
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid account ID")
			return
		}
		if err := svc.DeleteManagedAccount(c.Request.Context(), actorID, id); err != nil {
			switch {
			case errors.Is(err, auth.ErrUserNotFound):
				response.NotFound(c, "account")
			case errors.Is(err, auth.ErrCannotDeleteSelf):
				response.BadRequest(c, "CANNOT_DELETE_SELF", err.Error())
			default:
				response.InternalError(c)
			}
			return
		}
		response.OK(c, map[string]string{"status": "deleted"})
	}
}

func handleSetAccountExperts(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		actorID := c.MustGet("user_id").(uuid.UUID)
		userID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid account ID")
			return
		}
		var req struct {
			ExpertIDs []string `json:"expert_ids" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "INVALID_INPUT", err.Error())
			return
		}
		var expertIDs []uuid.UUID
		for _, s := range req.ExpertIDs {
			id, err := uuid.Parse(s)
			if err != nil {
				response.BadRequest(c, "INVALID_ID", "invalid expert_id")
				return
			}
			expertIDs = append(expertIDs, id)
		}
		if err := svc.SetAccountExpertGrants(c.Request.Context(), actorID, userID, expertIDs); err != nil {
			switch {
			case errors.Is(err, auth.ErrUserNotFound):
				response.NotFound(c, "account")
			case errors.Is(err, auth.ErrInvalidRole):
				response.BadRequest(c, "INVALID_ROLE", "expert grants only apply to domain_expert accounts")
			default:
				response.InternalError(c)
			}
			return
		}
		response.OK(c, map[string]string{"status": "updated"})
	}
}

func handleGetTOTPStatus(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(uuid.UUID)
		enabled, err := svc.GetTOTPStatus(c.Request.Context(), userID)
		if err != nil {
			response.NotFound(c, "user")
			return
		}
		response.OK(c, map[string]bool{"totpEnabled": enabled})
	}
}

func handleSetupTOTP(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			response.Forbidden(c, "Admin access required")
			return
		}
		userID := c.MustGet("user_id").(uuid.UUID)
		secret, qrURL, err := svc.SetupTOTP(c.Request.Context(), userID)
		if err != nil {
			response.InternalError(c)
			return
		}
		response.OK(c, map[string]string{"secret": secret, "qrUrl": qrURL})
	}
}

func handleEnableTOTP(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			response.Forbidden(c, "Admin access required")
			return
		}
		userID := c.MustGet("user_id").(uuid.UUID)
		var req struct {
			Code string `json:"code" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "INVALID_INPUT", err.Error())
			return
		}
		if err := svc.VerifyAndEnableTOTP(c.Request.Context(), userID, req.Code); err != nil {
			if errors.Is(err, auth.ErrInvalidTOTP) {
				response.Unauthorized(c, "Invalid TOTP code")
				return
			}
			response.BadRequest(c, "TOTP_SETUP", err.Error())
			return
		}
		response.OK(c, map[string]string{"status": "enabled"})
	}
}

func handleDisableTOTP(svc *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			response.Forbidden(c, "Admin access required")
			return
		}
		userID := c.MustGet("user_id").(uuid.UUID)
		if err := svc.DisableTOTP(c.Request.Context(), userID); err != nil {
			response.InternalError(c)
			return
		}
		response.OK(c, map[string]string{"status": "disabled"})
	}
}

// buildLogger creates a zap logger with the given level.
func buildLogger(level string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapLevel)
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return cfg.Build()
}
