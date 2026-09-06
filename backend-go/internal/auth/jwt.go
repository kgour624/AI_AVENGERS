package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/config"
)

// TokenType identifies the type of JWT token.
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Claims is the JWT payload.
// UserID and Role are the critical fields — used for authorization.
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`   // admin | client
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenPair holds both access and refresh tokens.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	// UserID is set internally so handlers can fetch the User record.
	// json:"-" means it is never serialized into the API response.
	UserID       uuid.UUID `json:"-"`
}

// JWTService handles token generation and validation.
// Refresh tokens are stored in Redis for revocation support.
// WHY Redis for refresh tokens:
// - Access tokens are stateless (short-lived, 15 min)
// - Refresh tokens need revocation (logout, security breach)
// - Redis TTL matches token expiry automatically
type JWTService struct {
	cfg    config.JWTConfig
	redis  *redis.Client
	logger *zap.Logger
}

// NewJWTService creates a new JWT service.
func NewJWTService(cfg config.JWTConfig, redisClient *redis.Client, logger *zap.Logger) *JWTService {
	return &JWTService{
		cfg:    cfg,
		redis:  redisClient,
		logger: logger,
	}
}

// IssueTokenPair generates both access and refresh tokens for a user.
// Stores refresh token in Redis with TTL.
func (s *JWTService) IssueTokenPair(ctx context.Context, userID uuid.UUID, email, role string) (*TokenPair, error) {
	accessExpiry := time.Now().Add(time.Duration(s.cfg.AccessExpiryMinutes) * time.Minute)
	refreshExpiry := time.Now().Add(time.Duration(s.cfg.RefreshExpiryDays) * 24 * time.Hour)

	// Generate access token
	accessToken, err := s.generateToken(userID, email, role, TokenTypeAccess, accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.generateToken(userID, email, role, TokenTypeRefresh, refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in Redis
	// Key: refresh:{userID}:{tokenID}
	// WHY include tokenID: allows revoking specific sessions, not all sessions
	//
	// WHY this is now a hard error (was previously logged-and-ignored):
	// ValidateRefreshToken requires this exact Redis key to exist later -
	// a token pair returned here as "successful" despite a failed Set()
	// would validate fine as a JWT (signature/expiry) but ALWAYS fail
	// ValidateRefreshToken's Redis-existence check on first use, since
	// it was never actually stored. That silently strands the session
	// (see main.go's handleRefresh comment for the full failure chain) -
	// far worse than failing loudly here, where the caller can retry.
	refreshClaims, _ := s.ParseToken(refreshToken)
	redisKey := fmt.Sprintf("refresh:%s:%s", userID, refreshClaims.ID)
	ttl := time.Until(refreshExpiry)

	if err := s.redis.Set(ctx, redisKey, refreshToken, ttl).Err(); err != nil {
		s.logger.Error("failed to store refresh token in Redis - failing token issuance",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to persist refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
		UserID:       userID,
	}, nil
}

// RefreshExpiryDays returns the configured refresh token expiry in days.
func (s *JWTService) RefreshExpiryDays() int {
	return s.cfg.RefreshExpiryDays
}

// ParseToken validates and parses a JWT token.
// Returns Claims if valid, error if invalid or expired.
func (s *JWTService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method is HMAC
			// WHY: Prevent algorithm confusion attacks
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(s.cfg.Secret), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ValidateAccessToken validates an access token.
// Checks: signature, expiry, token type.
// Does NOT check Redis — access tokens are stateless.
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := s.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeAccess {
		return nil, errors.New("not an access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token.
// Checks: signature, expiry, token type, AND Redis presence.
// WHY check Redis: refresh tokens can be revoked (logout).
func (s *JWTService) ValidateRefreshToken(ctx context.Context, tokenString string) (*Claims, error) {
	claims, err := s.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeRefresh {
		return nil, errors.New("not a refresh token")
	}

	// Check Redis — token must exist (not revoked)
	redisKey := fmt.Sprintf("refresh:%s:%s", claims.UserID, claims.ID)
	exists, err := s.redis.Exists(ctx, redisKey).Result()
	if err != nil {
		// Redis error — log and allow (fail open for availability)
		s.logger.Warn("Redis check failed for refresh token",
			zap.String("user_id", claims.UserID.String()),
			zap.Error(err),
		)
		return claims, nil
	}

	if exists == 0 {
		return nil, errors.New("refresh token has been revoked")
	}

	return claims, nil
}

// RevokeRefreshToken removes a refresh token from Redis.
// Called on logout.
func (s *JWTService) RevokeRefreshToken(ctx context.Context, tokenString string) error {
	claims, err := s.ParseToken(tokenString)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	redisKey := fmt.Sprintf("refresh:%s:%s", claims.UserID, claims.ID)
	if err := s.redis.Del(ctx, redisKey).Err(); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}

// RevokeAllUserTokens removes all refresh tokens for a user.
// Called on password change or security breach.
func (s *JWTService) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	pattern := fmt.Sprintf("refresh:%s:*", userID)
	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find user tokens: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	if err := s.redis.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	s.logger.Info("revoked all tokens for user",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(keys)),
	)

	return nil
}

// generateToken creates a signed JWT with the given claims.
func (s *JWTService) generateToken(
	userID uuid.UUID,
	email, role string,
	tokenType TokenType,
	expiry time.Time,
) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // Unique token ID for revocation
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiry),
			Issuer:    "ai-avengers",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, nil
}
