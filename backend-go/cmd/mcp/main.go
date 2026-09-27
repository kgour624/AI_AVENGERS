// Command mcp runs the Model Context Protocol server that lets an external
// coding agent (Claude Code and friends) work with this system's domain experts.
//
// WHY its own binary: it must be deployable, restartable and killable on its
// own. It shares no request path with the API server and only READS expert
// knowledge, so nothing here can slow down or break chat and workflows.
//
// Transport today is stdio: the client launches this binary and talks JSON-RPC
// over stdin/stdout. That needs no port, no TLS and no new exposed surface —
// the safest first step. Scope defaults to the whole roster, because a local
// stdio session belongs to whoever launched the binary.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"ai_avengers/backend/internal/mcp"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "mcp:", err)
		os.Exit(1)
	}
}

func run() error {
	// Logs go to stderr: stdout is the protocol channel and must carry nothing
	// but JSON-RPC, or the client cannot parse the session.
	logger := newLogger()

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is required")
	}

	// The session ends when the client exits; a signal must also stop it.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	catalog := mcp.NewPGCatalog(pool)

	tools := []mcp.Tool{
		mcp.NewListExpertsTool(catalog),
		mcp.NewGetStandardsTool(catalog),
	}

	// ask_expert and review_change need the running API server, because the
	// answer pipeline lives there and this binary must not grow a second copy of
	// it. Unset config disables exactly those two tools and says so — a server
	// that silently pretends to have them would just fail on first use.
	apiBase := strings.TrimSpace(os.Getenv("MCP_API_BASE_URL"))
	apiToken := strings.TrimSpace(os.Getenv("MCP_API_TOKEN"))
	apiChat := strings.TrimSpace(os.Getenv("MCP_CHAT_ID"))
	if apiBase != "" && apiToken != "" && apiChat != "" {
		answerer := mcp.NewHTTPAnswerer(apiBase, apiToken, apiChat)
		tools = append(tools,
			mcp.NewAskExpertTool(catalog, answerer),
			mcp.NewReviewChangeTool(catalog, answerer),
		)
	} else {
		logger.Warn("ask_expert and review_change are disabled; set MCP_API_BASE_URL, MCP_API_TOKEN and MCP_CHAT_ID to enable them")
	}

	registry := mcp.NewRegistry(tools...)

	version := strings.TrimSpace(os.Getenv("MCP_SERVER_VERSION"))
	if version == "" {
		version = "0.1.0"
	}

	// Comma-separated allow-lists. Empty means every domain and every tool,
	// which is right for a session on the owner's own machine and is the
	// boundary the remote/HTTP transport will tighten with real tokens.
	scope := mcp.Scope{
		Label:   "stdio",
		Domains: splitList(os.Getenv("MCP_ALLOWED_DOMAINS")),
		Tools:   splitList(os.Getenv("MCP_ALLOWED_TOOLS")),
	}

	logger.Info("mcp server starting",
		zap.String("version", version),
		zap.Int("domains", len(scope.Domains)),
		zap.Int("tools", len(scope.Tools)),
	)

	server := mcp.NewServer(registry, logger, version, scope)

	// HTTP mode: one URL per team, a token per developer. Started only when an
	// address is configured, so the default deployment stays stdio-only and
	// opens no port at all.
	if addr := strings.TrimSpace(os.Getenv("MCP_HTTP_ADDR")); addr != "" {
		tokenStore := mcp.NewPGTokenStore(pool)
		auditor := mcp.NewAuditor(ctx, pool, logger)
		server.SetTokenStore(tokenStore)
		server.SetAuditor(auditor)
		defer auditor.Close()

		httpServer := &http.Server{
			Addr:              addr,
			Handler:           server,
			ReadHeaderTimeout: 15 * time.Second,
			// No write timeout: an expert answer streams for minutes, and a
			// write deadline would cut a legitimate reply in half.
		}

		// Shutdown is bounded and owned by this function: cancel (a signal or
		// the client going away) stops accepting, then the in-flight handler
		// gets its grace period — no goroutine outlives run().
		shutdownDone := make(chan struct{})
		go func() {
			defer close(shutdownDone)
			<-ctx.Done()
			stopCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			if err := httpServer.Shutdown(stopCtx); err != nil {
				logger.Warn("mcp http shutdown", zap.Error(err))
			}
		}()

		logger.Info("mcp http listening", zap.String("addr", addr))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
		<-shutdownDone
		return nil
	}

	return server.ServeStdio(ctx, os.Stdin, os.Stdout)
}

// newLogger returns an stderr logger. Levels stay on in production (the same
// signal in dev and prod); the level is for alerting, not for silencing.
func newLogger() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stderr"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	l, err := cfg.Build()
	if err != nil {
		return zap.NewNop()
	}
	return l
}

// splitList turns "a, b ,c" into ["a","b","c"], dropping empties.
func splitList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	out := []string{}
	for _, part := range strings.Split(raw, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
