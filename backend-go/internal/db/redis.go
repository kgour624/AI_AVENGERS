package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/config"
)

// RedisClient wraps redis.Client with health check.
// Used for:
// - L1 expert memory (hot path, O(1) access)
// - JWT refresh token storage (for revocation)
// - Rate limit counters
// - Session data
type RedisClient struct {
	*redis.Client
	logger *zap.Logger
}

// ConnectRedis creates a new Redis client and verifies connectivity.
func ConnectRedis(ctx context.Context, cfg config.RedisConfig, logger *zap.Logger) (*RedisClient, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	if cfg.Password != "" {
		opts.Password = cfg.Password
	}
	opts.DB = cfg.DB

	// Connection pool settings
	// WHY PoolSize=10: Redis is single-threaded, more connections don't help
	// MinIdleConns=3: Keep warm connections for L1 memory hot path
	opts.PoolSize = 10
	opts.MinIdleConns = 3
	opts.ConnMaxLifetime = 30 * time.Minute

	client := redis.NewClient(opts)

	// Verify connectivity
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("redis connected", zap.String("url", cfg.URL))

	return &RedisClient{Client: client, logger: logger}, nil
}

// HealthCheck verifies Redis is reachable.
func (r *RedisClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := r.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}
	return nil
}

// Close gracefully closes the Redis connection.
func (r *RedisClient) Close() error {
	err := r.Client.Close()
	r.logger.Info("redis connection closed")
	return err
}
