package middleware

import (
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
		// Extract token from Authorization header
		// Format: "Bearer <token>"
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "Invalid authorization format. Use: Bearer <token>")
			c.Abort()
			return
		}

		tokenString := parts[1]

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
