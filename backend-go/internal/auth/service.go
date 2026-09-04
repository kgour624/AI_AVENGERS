package auth

import (
	"context"
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
