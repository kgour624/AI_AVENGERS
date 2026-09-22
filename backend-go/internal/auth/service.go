package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user record from the database.
// Only fields needed for auth — not the full user model.
type User struct {
	ID             uuid.UUID
	Email          string
	HashedPassword string
	FullName       string
	Role           string
	IsActive       bool
	TOTPSecret     string
	TOTPEnabled    bool
}

// AuthService handles user authentication.
// Separate from JWTService — single responsibility.
type AuthService struct {
	db     *pgxpool.Pool
	jwt    *JWTService
	logger *zap.Logger
}

// NewAuthService creates a new auth service.
func NewAuthService(db *pgxpool.Pool, jwt *JWTService, logger *zap.Logger) *AuthService {
	return &AuthService{
		db:     db,
		jwt:    jwt,
		logger: logger,
	}
}

// RegisterRequest holds registration input.
type RegisterRequest struct {
	Email    string
	Password string
	FullName string
}

// LoginRequest holds login input.
type LoginRequest struct {
	Email    string
	Password string
}

// AdminLoginRequest holds admin login input (includes TOTP).
type AdminLoginRequest struct {
	Email    string
	Password string
	TOTPCode string // 6-digit code from Google Authenticator
}

// Sentinel errors — use errors.Is() to check.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserInactive       = errors.New("account is disabled")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidTOTP        = errors.New("invalid TOTP code")
	ErrTOTPRequired       = errors.New("TOTP code required for admin login")
	ErrNotAdmin           = errors.New("admin access required")
	ErrInvalidResetToken  = errors.New("invalid or expired reset token")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
)

// Register creates a new client account.
// Hashes password with bcrypt cost 12.
// WHY cost 12: Balance between security and performance.
// Cost 10 = ~100ms, Cost 12 = ~400ms, Cost 14 = ~1.5s
// 400ms is acceptable for registration, not for every request.
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*TokenPair, error) {
	// Check if email already exists
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`,
		req.Email,
	).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, ErrEmailTaken
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert user
	var userID uuid.UUID
	err = s.db.QueryRow(ctx,
		`INSERT INTO users (email, hashed_password, full_name, role)
		 VALUES ($1, $2, $3, 'client')
		 RETURNING id`,
		req.Email, string(hashed), req.FullName,
	).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("user registered",
		zap.String("user_id", userID.String()),
		zap.String("email", req.Email),
	)

	// Issue tokens
	return s.jwt.IssueTokenPair(ctx, userID, req.Email, "client")
}

// Login authenticates a client user.
// Returns token pair on success.
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*TokenPair, error) {
	user, err := s.getUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Return same error as wrong password — prevent email enumeration
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update last login
	go s.updateLastLogin(context.Background(), user.ID)

	return s.jwt.IssueTokenPair(ctx, user.ID, user.Email, user.Role)
}

// AdminLogin authenticates an admin user with TOTP.
// Requires: email + password + TOTP code.
// WHY TOTP mandatory for admin: Admin has access to all data, expert management.
// A compromised admin account is catastrophic.
func (s *AuthService) AdminLogin(ctx context.Context, req AdminLoginRequest) (*TokenPair, error) {
	user, err := s.getUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Must be admin
	if user.Role != "admin" {
		return nil, ErrNotAdmin
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verify TOTP if enabled
	if user.TOTPEnabled {
		if req.TOTPCode == "" {
			return nil, ErrTOTPRequired
		}
		if !totp.Validate(req.TOTPCode, user.TOTPSecret) {
			s.logger.Warn("invalid TOTP attempt",
				zap.String("user_id", user.ID.String()),
				zap.String("email", user.Email),
			)
			return nil, ErrInvalidTOTP
		}
	}

	go s.updateLastLogin(context.Background(), user.ID)

	return s.jwt.IssueTokenPair(ctx, user.ID, user.Email, user.Role)
}

// SetupTOTP generates a new TOTP secret for an admin user.
// Returns the secret and QR code URL for Google Authenticator.
// Admin must verify with VerifyAndEnableTOTP before it's active.
func (s *AuthService) SetupTOTP(ctx context.Context, userID uuid.UUID) (secret, qrURL string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "AI Avengers",
		AccountName: userID.String(),
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Store secret (not yet enabled)
	_, err = s.db.Exec(ctx,
		`UPDATE users SET totp_secret = $1, updated_at = NOW() WHERE id = $2`,
		key.Secret(), userID,
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to store TOTP secret: %w", err)
	}

	return key.Secret(), key.URL(), nil
}

// VerifyAndEnableTOTP verifies a TOTP code and enables TOTP for the user.
// Must be called after SetupTOTP to confirm the user has scanned the QR code.
func (s *AuthService) VerifyAndEnableTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	var secret string
	err := s.db.QueryRow(ctx,
		`SELECT totp_secret FROM users WHERE id = $1`,
		userID,
	).Scan(&secret)
	if err != nil {
		return fmt.Errorf("failed to get TOTP secret: %w", err)
	}

	if secret == "" {
		return errors.New("TOTP not set up — call SetupTOTP first")
	}

	if !totp.Validate(code, secret) {
		return ErrInvalidTOTP
	}

	// Enable TOTP
	_, err = s.db.Exec(ctx,
		`UPDATE users SET totp_enabled = TRUE, updated_at = NOW() WHERE id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to enable TOTP: %w", err)
	}

	s.logger.Info("TOTP enabled for user", zap.String("user_id", userID.String()))
	return nil
}

// GetMe returns the authenticated user's public profile.
// Called by GET /auth/me after JWT is validated by AuthMiddleware.
// Returns only safe fields — never hashed_password or totp_secret.
//
// WHY by userID not email:
// AuthMiddleware already validated the JWT and extracted userID.
// Querying by userID is a direct primary key lookup — O(1).
// Querying by email would require an index scan — unnecessary.
func (s *AuthService) GetMe(ctx context.Context, userID uuid.UUID) (*User, error) {
	var user User
	err := s.db.QueryRow(ctx,
		`SELECT id, email, full_name, role, is_active, totp_enabled
		 FROM users
		 WHERE id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&user.ID, &user.Email, &user.FullName, &user.Role, &user.IsActive, &user.TOTPEnabled)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if !user.IsActive {
		return nil, ErrUserInactive
	}
	return &user, nil
}

// RefreshExpiryDays returns the configured refresh token expiry in days.
// Used by handlers to set cookie MaxAge.
func (s *AuthService) RefreshExpiryDays() int {
	return s.jwt.RefreshExpiryDays()
}

// ForgotPassword generates a password-reset token for the given email.
//
// WHY silent no-op on unknown email:
//   Returning an error when the email is not found leaks whether an
//   account exists (email enumeration). We always return success to
//   the caller; the handler logs the raw token so an admin can relay
//   it manually (or a future email integration can send it).
//
// Returns (rawToken, nil) on success, ("", nil) when email not found.
// The caller MUST NOT expose the distinction to the end user.
func (s *AuthService) ForgotPassword(ctx context.Context, email string) (rawToken string, err error) {
	user, err := s.getUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Silent — do not reveal whether the email exists.
			return "", nil
		}
		return "", err
	}

	// Invalidate any previous unused tokens for this user.
	// WHY: only one active reset link at a time; prevents confusion if
	// the user clicks "Forgot Password" twice.
	_, err = s.db.Exec(ctx,
		`UPDATE password_reset_tokens
		    SET used_at = NOW()
		  WHERE user_id = $1 AND used_at IS NULL`,
		user.ID,
	)
	if err != nil {
		return "", fmt.Errorf("failed to invalidate old reset tokens: %w", err)
	}

	// Generate 32 cryptographically-random bytes → hex string (64 chars).
	// WHY 32 bytes: 256 bits of entropy — brute-force is infeasible.
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate reset token: %w", err)
	}
	rawToken = hex.EncodeToString(buf)

	// Store SHA-256(rawToken) — never the raw token itself.
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	_, err = s.db.Exec(ctx,
		`INSERT INTO password_reset_tokens (user_id, token_hash)
		 VALUES ($1, $2)`,
		user.ID, tokenHash,
	)
	if err != nil {
		return "", fmt.Errorf("failed to store reset token: %w", err)
	}

	s.logger.Info("password reset token generated",
		zap.String("user_id", user.ID.String()),
		zap.String("email", email),
	)

	return rawToken, nil
}

// ResetPassword validates a reset token and sets a new password.
//
// Steps:
//  1. Hash the incoming token and look it up.
//  2. Validate: exists, not expired, not already used.
//  3. bcrypt the new password and update users.
//  4. Mark the token used (single-use).
//  5. Revoke all active JWT sessions (force re-login).
func (s *AuthService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrPasswordTooShort
	}

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	var (
		tokenID   uuid.UUID
		userID    uuid.UUID
		expiresAt time.Time
		usedAt    *time.Time
	)
	err := s.db.QueryRow(ctx,
		`SELECT id, user_id, expires_at, used_at
		   FROM password_reset_tokens
		  WHERE token_hash = $1`,
		tokenHash,
	).Scan(&tokenID, &userID, &expiresAt, &usedAt)
	if err != nil {
		// No row or scan error — treat both as invalid token.
		return ErrInvalidResetToken
	}

	if usedAt != nil {
		return ErrInvalidResetToken // already used
	}
	if time.Now().After(expiresAt) {
		return ErrInvalidResetToken // expired
	}

	// Hash new password.
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password.
	_, err = s.db.Exec(ctx,
		`UPDATE users
		    SET hashed_password = $1, updated_at = NOW()
		  WHERE id = $2`,
		string(hashed), userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used — single-use guarantee.
	_, err = s.db.Exec(ctx,
		`UPDATE password_reset_tokens
		    SET used_at = NOW()
		  WHERE id = $1`,
		tokenID,
	)
	if err != nil {
		// Non-fatal: password is already changed. Log and continue.
		s.logger.Warn("failed to mark reset token used",
			zap.String("token_id", tokenID.String()),
			zap.Error(err),
		)
	}

	// Revoke all active JWT sessions — user must log in with new password.
	if err := s.jwt.RevokeAllUserTokens(ctx, userID); err != nil {
		s.logger.Warn("failed to revoke user tokens after password reset",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		// Non-fatal: password is changed; old tokens will expire naturally.
	}

	s.logger.Info("password reset successful",
		zap.String("user_id", userID.String()),
	)

	return nil
}

// getUserByEmail fetches a user by email.
// Returns ErrUserNotFound if not found.
func (s *AuthService) getUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := s.db.QueryRow(ctx,
		`SELECT id, email, hashed_password, full_name, role, is_active,
		        COALESCE(totp_secret, ''), totp_enabled
		 FROM users
		 WHERE email = $1 AND deleted_at IS NULL`,
		email,
	).Scan(
		&user.ID, &user.Email, &user.HashedPassword, &user.FullName,
		&user.Role, &user.IsActive, &user.TOTPSecret, &user.TOTPEnabled,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// updateLastLogin updates the last_login timestamp.
// Called async — non-critical, don't block login response.
func (s *AuthService) updateLastLogin(ctx context.Context, userID uuid.UUID) {
	_, err := s.db.Exec(ctx,
		`UPDATE users SET last_login = $1 WHERE id = $2`,
		time.Now(), userID,
	)
	if err != nil {
		s.logger.Warn("failed to update last login",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
	}
}
