package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	adminpkg "ai_avengers/backend/internal/admin"
	"ai_avengers/backend/internal/auth"
	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/category"
	"ai_avengers/backend/internal/chat"
	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/config"
	appcontext "ai_avengers/backend/internal/context"
	"ai_avengers/backend/internal/db"
	"ai_avengers/backend/internal/decision"
	"ai_avengers/backend/internal/expert"
	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/memory"
	"ai_avengers/backend/internal/message"
	"ai_avengers/backend/internal/middleware"
	"ai_avengers/backend/internal/ml"
	"ai_avengers/backend/internal/orchestrator"
	"ai_avengers/backend/internal/project"
	"ai_avengers/backend/internal/rating"
	"ai_avengers/backend/internal/repo"
	"ai_avengers/backend/internal/response"
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

	// Background context for startup
	ctx := context.Background()

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

	// Build router — single call, single definition
	router := buildRouter(cfg, logger, postgres, redisClient, jwtService, authService, modelGateway, mlClient, domainRegistry, categoryRegistry)

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
	cfg *config.Config,
	logger *zap.Logger,
	postgres *db.Pool,
	redisClient *db.RedisClient,
	jwtService *auth.JWTService,
	authService *auth.AuthService,
	modelGateway *gateway.ModelGateway,
	mlClient *ml.SidecarClient,
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
	memManager := memory.NewManager(postgres.Pool, redisClient.Client, mlClient, logger)
	// domainRegistry is already initialized in main() and passed here.
	chinawallEnforcer := chinawall.NewEnforcer(cfg.ChinaWall, modelGateway, mlClient, logger, domainRegistry)
	decisionEngine := decision.NewEngine(postgres.Pool, modelGateway, chinawallEnforcer, logger)
	contextAssembler := appcontext.NewAssembler(
		postgres.Pool, mlClient, memManager,
		cfg.Context.MaxTokens, cfg.Context.RecentMessages,
		cfg.Context.SemanticTopK, cfg.Context.CourseChunksTopK,
		logger,
	)
	orch := orchestrator.NewOrchestrator(postgres.Pool, contextAssembler, decisionEngine, memManager, categoryRegistry, logger)

	// Initialize domain services
	projectSvc := project.NewService(postgres.Pool, logger)
	chatSvc := chat.NewService(postgres.Pool, logger)
	ratingSvc := rating.NewService(postgres.Pool, memManager, logger)
	repoSvc := repo.NewService(
		postgres.Pool, mlClient, redisClient.Client,
		cfg.Security.EncryptionKey,
		cfg.OAuth.GitHubClientID, cfg.OAuth.GitHubClientSecret,
		cfg.OAuth.GitLabClientID, cfg.OAuth.GitLabClientSecret,
		cfg.OAuth.BaseURL, cfg.OAuth.FrontendURL,
		logger,
	)

	// Initialize HTTP handlers
	projectHandler := project.NewHandler(projectSvc, logger)
	chatHandler := chat.NewHandler(chatSvc, logger)
	messageHandler := message.NewHandler(chatSvc, orch, modelGateway, mlClient, memManager, logger)
	ratingHandler := rating.NewHandler(ratingSvc, logger)
	expertHandler := expert.NewHandler(postgres.Pool, logger)
	repoHandler := repo.NewHandler(repoSvc, logger)
	adminHandler := adminpkg.NewAdminHandler(postgres.Pool, modelGateway, mlClient, categoryRegistry, logger)

	// Collaboration layer (Phase C + D)
	bbStore := blackboard.NewStore(postgres.Pool, redisClient.Client, logger)
	bbSubscriber := blackboard.NewSubscriber(bbStore, redisClient.Client, logger)
	wfEngine := workflow.NewEngine(postgres.Pool, logger)
	validationPipeline := validation.NewPipeline(modelGateway, logger)
	wfTools := workflow.NewTools(bbStore, wfEngine, bbSubscriber, validationPipeline, logger)
	_ = wfTools // used by expert execution loop (Phase E)
	wfHandler := workflow.NewHandler(wfEngine, bbStore, logger)

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

	v1 := router.Group("/api/v1")

	// ============================================================
	// Auth routes — no JWT required
	// ============================================================
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", handleRegister(authService))
		authGroup.POST("/login", handleLogin(authService))
		authGroup.POST("/admin/login", handleAdminLogin(authService))
		authGroup.POST("/refresh", handleRefresh(jwtService))
		authGroup.POST("/logout", middleware.AuthMiddleware(jwtService, logger), handleLogout(jwtService))
	}

	// OAuth callback — no JWT (browser redirect from GitHub/GitLab)
	// WHY public: OAuth provider redirects browser here with ?code=...
	// The browser has no JWT at this point — it's a fresh redirect.
	v1.GET("/repo/callback/:provider", repoHandler.OAuthCallback)

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

		// Expert routes (read-only for clients)
		experts := protected.Group("/experts")
		{
			experts.GET("", expertHandler.ListActive)
			experts.GET("/:id", expertHandler.GetByID)
			experts.GET("/:id/topics", expertHandler.GetTopics)
		}

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
			chats.POST("/:id/messages", messageHandler.Send)
			chats.GET("/:id/messages", chatHandler.ListMessages)
		}

		// Message routes
		messages := protected.Group("/messages")
		{
			messages.POST("/:id/rate", ratingHandler.Rate)
		}

		// Workflow routes (Phase C — collaboration layer)
		workflows := protected.Group("/workflows")
		{
			workflows.POST("", wfHandler.CreateWorkflow)
			workflows.GET("/:id", wfHandler.GetWorkflow)
			workflows.POST("/:id/start", wfHandler.StartWorkflow)
			workflows.GET("/:id/blackboard", wfHandler.GetBlackboard)
			workflows.GET("/:id/kanban", wfHandler.GetKanban)
			workflows.POST("/:id/approvals/:aid/respond", wfHandler.RespondToApproval)
		}
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
		adminGroup.GET("/experts/:id/jobs", adminHandler.GetIngestionJobs)
		adminGroup.GET("/experts/:id/jobs/stream", adminHandler.StreamIngestionJob)
		adminGroup.POST("/experts/:id/jobs/:jobID/resume", adminHandler.ResumeIngestionJob)
		adminGroup.GET("/clients", adminHandler.ListClients)
		adminGroup.PATCH("/clients/:id", adminHandler.UpdateClient)
		adminGroup.GET("/stats", adminHandler.GetStats)
		adminGroup.GET("/violations", adminHandler.GetViolations)
		adminGroup.GET("/ratings", adminHandler.GetRatings)
		adminGroup.GET("/settings", adminHandler.GetSettings)
		adminGroup.PATCH("/settings/:key", adminHandler.UpdateSetting)
		// LLM provider + API key management (no env file needed)
		adminGroup.GET("/llm-settings", adminHandler.GetLLMSettings)
		adminGroup.POST("/llm-settings", adminHandler.UpdateLLMSettings)
		// Expert categories (migration 010, CT-A3) — admin-owned template layer.
		adminGroup.GET("/expert-categories", adminHandler.ListExpertCategories)
		adminGroup.POST("/expert-categories", adminHandler.CreateExpertCategory)
		adminGroup.GET("/expert-categories/:id", adminHandler.GetExpertCategory)
		adminGroup.PATCH("/expert-categories/:id", adminHandler.UpdateExpertCategory)
	}

	return router
}

// ============================================================
// Inline route handlers
// ============================================================

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
			switch err {
			case auth.ErrEmailTaken:
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
			switch err {
			case auth.ErrInvalidCredentials, auth.ErrUserNotFound:
				response.Unauthorized(c, "Invalid email or password")
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
