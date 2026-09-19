package middleware

import (
	"crypto/subtle"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/auth"
	"ai_avengers/backend/internal/response"
)

// AuthMiddleware validates JWT access tokens.
// Sets user_id, email, role in gin context for downstream handlers.
// WHY extract to middleware: Every protected route needs auth.
// Centralizing prevents forgetting auth on a route.
func AuthMiddleware(jwtService *auth.JWTService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token — two sources, in priority order:
		// 1. Authorization: Bearer <token>  (standard for all requests)
		// 2. ?token=<token> query param      (SSE only — EventSource cannot set headers)
		var tokenString string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				response.Unauthorized(c, "Invalid authorization format. Use: Bearer <token>")
				c.Abort()
				return
			}
			tokenString = parts[1]
		} else if qToken := c.Query("token"); qToken != "" {
			// SSE fallback: EventSource cannot send Authorization header.
			// Frontend passes access token as ?token= for SSE endpoints only.
			tokenString = qToken
		} else {
			response.Unauthorized(c, "Authorization required")
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			logger.Debug("invalid access token", zap.Error(err))
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Set claims in context for downstream handlers
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// ServiceTokenMiddleware authenticates a backend service by shared secret.
//
// WHY not AuthMiddleware: AuthMiddleware validates a user's JWT access token.
// AiderService is a process, not a user — it has no login, no refresh cookie
// and no way to mint an access token. Putting /llm/proxy behind
// AuthMiddleware meant every Aider LLM call was answered with 401.
//
// The token travels as "Authorization: Bearer <token>" because that is what
// the OpenAI client library inside Aider sends for its api_key. No custom
// header is possible on that path.
//
// An empty configured token rejects every request with 503 rather than
// letting the route through. A blank shared secret must never mean
// "authentication disabled" on an endpoint that spends money.
func ServiceTokenMiddleware(expectedToken string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expectedToken == "" {
			logger.Error("service token route called but AIDER_PROXY_TOKEN is not set")
			response.ServiceUnavailable(c, "Service token not configured")
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "Invalid authorization format. Use: Bearer <token>")
			c.Abort()
			return
		}

		// Constant-time compare: a shared secret checked with == leaks its
		// prefix length through response timing.
		if subtle.ConstantTimeCompare([]byte(parts[1]), []byte(expectedToken)) != 1 {
			logger.Warn("service token mismatch", zap.String("path", c.Request.URL.Path))
			response.Unauthorized(c, "Invalid service token")
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminMiddleware ensures the authenticated user is an admin.
// Must be used AFTER AuthMiddleware.
// WHY separate from AuthMiddleware: Some routes need auth but not admin.
// Composing middleware is cleaner than one mega-middleware.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		if role.(string) != "admin" {
			response.Forbidden(c, "Admin access required")
			c.Abort()
			return
		}

		c.Next()
	}
}
