package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
)

// CORSMiddleware handles Cross-Origin Resource Sharing.
// MUST be registered BEFORE auth middleware.
//
// WHY before auth:
// Browser sends OPTIONS preflight with NO Authorization header.
// If auth middleware runs first -> 401 on OPTIONS -> preflight fails
// -> browser blocks the actual POST/GET request entirely.
// CORS middleware must short-circuit OPTIONS before any auth check.
//
// WHY CORS is needed:
// Frontend runs on localhost:3000, backend on localhost:8080.
// Different ports = different origin = browser enforces CORS.
// Without this, every API call fails with "Network Error" in browser.
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowedSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowedSet[strings.TrimRight(o, "/")] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Determine if origin is allowed
		allowOrigin := ""
		if allowedSet[origin] {
			allowOrigin = origin
		} else if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
			allowOrigin = "*"
		}

		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers",
				"Origin,Content-Type,Authorization,X-Request-ID,Cache-Control")
			c.Header("Access-Control-Expose-Headers", "X-Request-ID")
			c.Header("Access-Control-Max-Age", "86400") // 24h preflight cache
		}

		// Handle preflight — respond immediately, don't pass to router
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent) // 204
			return
		}

		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to every request.
// Used for tracing requests across logs.
// WHY: When debugging production issues, request ID links all log lines.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if client sent a request ID (for distributed tracing)
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// LoggerMiddleware logs every request with timing and status.
// Uses structured logging (zap) for machine-parseable logs.
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		// Choose log level based on status code
		// WHY: 5xx errors are more urgent than 4xx
		logFields := []zap.Field{
			zap.String("request_id", c.GetString("request_id")),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("ip", c.ClientIP()),
		}

		switch {
		case status >= 500:
			logger.Error("request failed", logFields...)
		case status >= 400:
			logger.Warn("request error", logFields...)
		default:
			logger.Info("request completed", logFields...)
		}
	}
}

// RecoveryMiddleware catches panics and returns 500.
// Must be the FIRST middleware — catches panics from all other middleware.
// WHY: A panic in any handler would crash the server without this.
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered",
					zap.String("request_id", c.GetString("request_id")),
					zap.String("path", c.Request.URL.Path),
					zap.Any("error", err),
				)
				response.InternalError(c)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// RateLimitMiddleware limits requests per IP using Redis sliding window.
// WHY sliding window over fixed window:
// Fixed window allows burst at window boundary (e.g., 100 req at 0:59 + 100 req at 1:00)
// Sliding window prevents this — always looks at last N seconds.
func RateLimitMiddleware(redisClient *redis.Client, perIPLimit int, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("ratelimit:ip:%s", ip)
		now := time.Now().UnixMilli()
		window := int64(60 * 1000) // 60 seconds in milliseconds

		// Sliding window using Redis sorted set
		// Score = timestamp, Member = unique request ID
		// Remove old entries outside window
		pipe := redisClient.Pipeline()
		pipe.ZRemRangeByScore(c.Request.Context(), key, "0", fmt.Sprintf("%d", now-window))
		pipe.ZAdd(c.Request.Context(), key, redis.Z{
			Score:  float64(now),
			Member: fmt.Sprintf("%d", now),
		})
		pipe.ZCard(c.Request.Context(), key)
		pipe.Expire(c.Request.Context(), key, 2*time.Minute)

		cmds, err := pipe.Exec(c.Request.Context())
		if err != nil {
			// Redis error — allow request (fail open)
			logger.Warn("rate limit Redis error", zap.Error(err))
			c.Next()
			return
		}

		// Get count from ZCard result (index 2)
		count := cmds[2].(*redis.IntCmd).Val()

		if int(count) > perIPLimit {
			logger.Warn("rate limit exceeded",
				zap.String("ip", ip),
				zap.Int64("count", count),
			)
			response.TooManyRequests(c)
			c.Abort()
			return
		}

		c.Next()
	}
}
