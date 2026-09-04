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

	// Build router
	router := buildRouter(cfg, logger, postgres, redisClient, jwtService, authService, modelGateway, mlClient)

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
func buildRouter(
	cfg *config.Config,
	logger *zap.Logger,
	postgres *db.Pool,
	redisClient *db.RedisClient,
	jwtService *auth.JWTService,
	authService *auth.AuthService,
	modelGateway *gateway.ModelGateway,
	mlClient *ml.SidecarClient,
) *gin.Engine {
	router := gin.New()

	// Middleware order: Recovery -> RequestID -> Logger -> RateLimit
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.RateLimitMiddleware(redisClient.Client, cfg.RateLimit.PerIP, logger))

	// Initialize services
	memManager := memory.NewManager(postgres.Pool, redisClient.Client, mlClient, logger)
	chinawallEnforcer := chinawall.NewEnforcer(cfg.ChinaWall, modelGateway, mlClient, logger)
	decisionEngine := decision.NewEngine(postgres.Pool, modelGateway, chinawallEnforcer, logger)
	contextAssembler := appcontext.NewAssembler(
		postgres.Pool, mlClient, memManager,
		cfg.Context.MaxTokens, cfg.Context.RecentMessages,
		cfg.Context.SemanticTopK, cfg.Context.CourseChunksTopK,
		logger,
	)
	orch := orchestrator.NewOrchestrator(postgres.Pool, contextAssembler, decisionEngine, memManager, logger)

	projectSvc := project.NewService(postgres.Pool, logger)
	chatSvc := chat.NewService(postgres.Pool, logger)
	ratingSvc := rating.NewService(postgres.Pool, memManager, logger)
	repoSvc := repo.NewService(
		postgres.Pool, mlClient,
		cfg.Security.EncryptionKey,
		"", "", // GitHub OAuth (set via env)
		"", "", // GitLab OAuth (set via env)
		"",     // Base URL
		logger,
	)

	projectHandler := project.NewHandler(projectSvc, logger)
	chatHandler := chat.NewHandler(chatSvc, logger)
	messageHandler := message.NewHandler(chatSvc, orch, modelGateway, mlClient, memManager, logger)
	ratingHandler := rating.NewHandler(ratingSvc, logger)
	expertHandler := expert.NewHandler(postgres.Pool, logger)
	repoHandler := repo.NewHandler(repoSvc, logger)
	adminHandler := adminpkg.NewAdminHandler(postgres.Pool, modelGateway, mlClient, logger)

	// Health check
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

	// Auth routes
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", handleRegister(authService))
		authGroup.POST("/login", handleLogin(authService))
		authGroup.POST("/admin/login", handleAdminLogin(authService))
		authGroup.POST("/refresh", handleRefresh(jwtService))
		authGroup.POST("/logout", middleware.AuthMiddleware(jwtService, logger), handleLogout(jwtService))
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService, logger))
	{
		// Expert routes
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
			// Repo integration routes
			projects.POST("/:id/repo", repoHandler.ConnectRepo)
			projects.POST("/:id/repo/sync", repoHandler.SyncRepo)
			projects.GET("/:id/repo/status", repoHandler.GetSyncStatus)
		}

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
	}

	// Admin routes
	adminGroup := v1.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(jwtService, logger))
	adminGroup.Use(middleware.AdminMiddleware())
	{
		adminGroup.GET("/experts", adminHandler.ListExperts)
		adminGroup.POST("/experts", adminHandler.CreateExpert)
		adminGroup.PATCH("/experts/:id", adminHandler.UpdateExpert)
		adminGroup.POST("/experts/:id/ingest", adminHandler.IngestTranscript)
		adminGroup.GET("/experts/:id/jobs", adminHandler.GetIngestionJobs)
		adminGroup.GET("/clients", adminHandler.ListClients)
		adminGroup.PATCH("/clients/:id", adminHandler.UpdateClient)
		adminGroup.GET("/stats", adminHandler.GetStats)
		adminGroup.GET("/violations", adminHandler.GetViolations)
		adminGroup.GET("/ratings", adminHandler.GetRatings)
		adminGroup.GET("/settings", adminHandler.GetSettings)
		adminGroup.PATCH("/settings/:key", adminHandler.UpdateSetting)
	}

	return router
}

// handleGetProjectMemory GET /projects/:id/memory
func handleGetProjectMemory(memManager *memory.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid project ID")
			return
		}
		events, err := memManager.GetTimeline(c.Request.Context(), projectID, 20, 0)
		if err != nil {
			response.InternalError(c)
			return
		}
		response.OK(c, events)
	}
}

// handleGetProjectTimeline GET /projects/:id/timeline
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

// handleAdminViolations GET /admin/violations
func handleAdminViolations(memManager *memory.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		violations, err := memManager.GetViolations(c.Request.Context(), 100)
		if err != nil {
			response.InternalError(c)
			return
		}
		response.OK(c, violations)
	}
}

	// Health check — no auth required
	router.GET("/health", func(c *gin.Context) {
		health := map[string]string{
			"status":  "ok",
			"version": "1.0.0",
		}

		// Check dependencies
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

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Auth routes — no JWT required
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", handleRegister(authService))
		authGroup.POST("/login", handleLogin(authService))
		authGroup.POST("/admin/login", handleAdminLogin(authService))
		authGroup.POST("/refresh", handleRefresh(jwtService))
		authGroup.POST("/logout", middleware.AuthMiddleware(jwtService, logger), handleLogout(jwtService))
	}

	// Protected routes — JWT required
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService, logger))
	{
		// Expert routes (read-only for clients)
		experts := protected.Group("/experts")
		{
			experts.GET("", handleListExperts(postgres))
			experts.GET("/:id", handleGetExpert(postgres))
			experts.GET("/:id/topics", handleGetExpertTopics(postgres))
		}

		// Project routes
		projects := protected.Group("/projects")
		{
			projects.POST("", handleCreateProject(postgres))
			projects.GET("", handleListProjects(postgres))
			projects.GET("/:id", handleGetProject(postgres))
			projects.PATCH("/:id", handleUpdateProject(postgres))
			projects.DELETE("/:id", handleDeleteProject(postgres))
			projects.POST("/:id/experts", handleAddExpertToProject(postgres))
			projects.DELETE("/:id/experts/:expertId", handleRemoveExpertFromProject(postgres))
			projects.POST("/:id/chats", handleCreateChat(postgres))
			projects.GET("/:id/chats", handleListChats(postgres))
			projects.GET("/:id/memory", handleGetProjectMemory(postgres))
			projects.GET("/:id/timeline", handleGetProjectTimeline(postgres))
		}

		// Chat routes
		chats := protected.Group("/chats")
		{
			chats.GET("/:id", handleGetChat(postgres))
			chats.PATCH("/:id", handleUpdateChat(postgres))
			chats.DELETE("/:id", handleArchiveChat(postgres))
			chats.POST("/:id/messages", handleSendMessage(postgres, modelGateway, mlClient, logger))
			chats.GET("/:id/messages", handleListMessages(postgres))
		}

		// Message routes
		messages := protected.Group("/messages")
		{
			messages.POST("/:id/rate", handleRateMessage(postgres))
		}
	}

	// Admin routes — JWT + admin role required
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthMiddleware(jwtService, logger))
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/experts", handleAdminListExperts(postgres))
		admin.POST("/experts", handleAdminCreateExpert(postgres))
		admin.PATCH("/experts/:id", handleAdminUpdateExpert(postgres))
		admin.POST("/experts/:id/ingest", handleAdminIngestTranscript(postgres, mlClient, modelGateway))
		admin.GET("/experts/:id/jobs", handleAdminGetIngestionJobs(postgres))
		admin.GET("/clients", handleAdminListClients(postgres))
		admin.PATCH("/clients/:id", handleAdminUpdateClient(postgres))
		admin.GET("/stats", handleAdminGetStats(postgres, modelGateway))
		admin.GET("/violations", handleAdminGetViolations(postgres))
		admin.GET("/ratings", handleAdminGetRatings(postgres))
		admin.GET("/settings", handleAdminGetSettings(postgres))
		admin.PATCH("/settings/:key", handleAdminUpdateSetting(postgres))
	}

	return router
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

// ============================================================
// Route handlers — stubs for Phase 1
// Full implementation in Phase 3-4
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

		response.Created(c, tokens)
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

		response.OK(c, tokens)
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

		response.OK(c, tokens)
	}
}

func handleRefresh(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "INVALID_INPUT", err.Error())
			return
		}

		claims, err := jwtService.ValidateRefreshToken(c.Request.Context(), req.RefreshToken)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired refresh token")
			return
		}

		// Revoke old refresh token
		_ = jwtService.RevokeRefreshToken(c.Request.Context(), req.RefreshToken)

		// Issue new token pair
		tokens, err := jwtService.IssueTokenPair(c.Request.Context(), claims.UserID, claims.Email, claims.Role)
		if err != nil {
			response.InternalError(c)
			return
		}

		response.OK(c, tokens)
	}
}

func handleLogout(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = c.ShouldBindJSON(&req)

		if req.RefreshToken != "" {
			_ = jwtService.RevokeRefreshToken(c.Request.Context(), req.RefreshToken)
		}

		response.OK(c, map[string]string{"message": "logged out"})
	}
}

// Stub handlers — return 501 until implemented in later phases
func stubHandler(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		response.OK(c, map[string]string{"status": "not_implemented", "handler": name})
	}
}

func handleListExperts(db *db.Pool) gin.HandlerFunc          { return stubHandler("list_experts") }
func handleGetExpert(db *db.Pool) gin.HandlerFunc            { return stubHandler("get_expert") }
func handleGetExpertTopics(db *db.Pool) gin.HandlerFunc      { return stubHandler("get_expert_topics") }
func handleCreateProject(db *db.Pool) gin.HandlerFunc        { return stubHandler("create_project") }
func handleListProjects(db *db.Pool) gin.HandlerFunc         { return stubHandler("list_projects") }
func handleGetProject(db *db.Pool) gin.HandlerFunc           { return stubHandler("get_project") }
func handleUpdateProject(db *db.Pool) gin.HandlerFunc        { return stubHandler("update_project") }
func handleDeleteProject(db *db.Pool) gin.HandlerFunc        { return stubHandler("delete_project") }
func handleAddExpertToProject(db *db.Pool) gin.HandlerFunc   { return stubHandler("add_expert") }
func handleRemoveExpertFromProject(db *db.Pool) gin.HandlerFunc { return stubHandler("remove_expert") }
func handleCreateChat(db *db.Pool) gin.HandlerFunc           { return stubHandler("create_chat") }
func handleListChats(db *db.Pool) gin.HandlerFunc            { return stubHandler("list_chats") }
func handleGetProjectMemory(db *db.Pool) gin.HandlerFunc     { return stubHandler("project_memory") }
func handleGetProjectTimeline(db *db.Pool) gin.HandlerFunc   { return stubHandler("project_timeline") }
func handleGetChat(db *db.Pool) gin.HandlerFunc              { return stubHandler("get_chat") }
func handleUpdateChat(db *db.Pool) gin.HandlerFunc           { return stubHandler("update_chat") }
func handleArchiveChat(db *db.Pool) gin.HandlerFunc          { return stubHandler("archive_chat") }
func handleListMessages(db *db.Pool) gin.HandlerFunc         { return stubHandler("list_messages") }
func handleRateMessage(db *db.Pool) gin.HandlerFunc          { return stubHandler("rate_message") }
func handleAdminListExperts(db *db.Pool) gin.HandlerFunc     { return stubHandler("admin_list_experts") }
func handleAdminCreateExpert(db *db.Pool) gin.HandlerFunc    { return stubHandler("admin_create_expert") }
func handleAdminUpdateExpert(db *db.Pool) gin.HandlerFunc    { return stubHandler("admin_update_expert") }
func handleAdminListClients(db *db.Pool) gin.HandlerFunc     { return stubHandler("admin_list_clients") }
func handleAdminUpdateClient(db *db.Pool) gin.HandlerFunc    { return stubHandler("admin_update_client") }
func handleAdminGetViolations(db *db.Pool) gin.HandlerFunc   { return stubHandler("admin_violations") }
func handleAdminGetRatings(db *db.Pool) gin.HandlerFunc      { return stubHandler("admin_ratings") }
func handleAdminGetSettings(db *db.Pool) gin.HandlerFunc     { return stubHandler("admin_settings") }
func handleAdminUpdateSetting(db *db.Pool) gin.HandlerFunc   { return stubHandler("admin_update_setting") }

func handleSendMessage(db *db.Pool, gw *gateway.ModelGateway, ml *ml.SidecarClient, logger *zap.Logger) gin.HandlerFunc {
	return stubHandler("send_message")
}
func handleAdminIngestTranscript(db *db.Pool, ml *ml.SidecarClient, gw *gateway.ModelGateway) gin.HandlerFunc {
	return stubHandler("admin_ingest_transcript")
}
func handleAdminGetIngestionJobs(db *db.Pool) gin.HandlerFunc { return stubHandler("admin_ingestion_jobs") }
func handleAdminGetStats(db *db.Pool, gw *gateway.ModelGateway) gin.HandlerFunc {
	return stubHandler("admin_stats")
}
