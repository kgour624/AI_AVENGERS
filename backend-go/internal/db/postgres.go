package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/config"
)

// Pool wraps pgxpool.Pool with health check capability.
// Single instance — shared across all services.
// Never create a new pool per request.
type Pool struct {
	*pgxpool.Pool
	logger *zap.Logger
}

// Connect creates a new PostgreSQL connection pool.
// Verifies connectivity before returning.
// Returns error if connection cannot be established within 10 seconds.
func Connect(ctx context.Context, cfg config.DatabaseConfig, logger *zap.Logger) (*Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Connection pool settings
	// WHY these values:
	// MaxConns=20: Handles concurrent requests without overwhelming DB
	// MinConns=5: Keep warm connections ready — avoids cold start latency
	// MaxConnLifetime=1h: Recycle connections to prevent stale state
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connectivity
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connected",
		zap.String("max_conns", fmt.Sprintf("%d", cfg.MaxConns)),
		zap.String("min_conns", fmt.Sprintf("%d", cfg.MinConns)),
	)

	return &Pool{Pool: pool, logger: logger}, nil
}

// HealthCheck verifies the database is reachable.
// Used by health check endpoint.
func (p *Pool) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// Close gracefully closes all connections in the pool.
// Called during server shutdown.
func (p *Pool) Close() {
	p.Pool.Close()
	p.logger.Info("database connection pool closed")
}
