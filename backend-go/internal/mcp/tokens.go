package mcp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TokenPrefix identifies a key as an MCP token, so it is recognizable in a
// config file or a log without being usable on its own.
const TokenPrefix = "mcp_"

// TokenInfo is what the server knows about the caller once a token resolved.
type TokenInfo struct {
	ID    string
	Label string
}

// TokenStore resolves a raw bearer token to the scope it grants.
//
// WHY an interface: the HTTP transport must be testable without a database, and
// the stdio path does not need it at all.
type TokenStore interface {
	Resolve(ctx context.Context, rawToken string) (Scope, TokenInfo, error)
	Touch(ctx context.Context, tokenID string) error
}

// PGTokenStore reads tokens from Postgres.
//
// It never stores or logs the plaintext: the caller presents a token, this
// hashes it and looks the hash up, so a database leak yields no usable
// credential.
type PGTokenStore struct {
	db *pgxpool.Pool
}

// NewPGTokenStore wraps a pool.
func NewPGTokenStore(db *pgxpool.Pool) *PGTokenStore { return &PGTokenStore{db: db} }

// HashToken returns the stored representation of a raw token.
func HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(rawToken)))
	return hex.EncodeToString(sum[:])
}

// NewToken mints a token. Returned once, never recoverable afterwards.
func NewToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("mcp: mint token: %w", err)
	}
	return TokenPrefix + hex.EncodeToString(buf), nil
}

// Resolve maps a raw token to its scope. An unknown or revoked token is a
// caller-facing FORBIDDEN, not an internal error — the client can fix it by
// using the right token.
func (s *PGTokenStore) Resolve(ctx context.Context, rawToken string) (Scope, TokenInfo, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return Scope{}, TokenInfo{}, NewToolError("FORBIDDEN", "missing bearer token")
	}

	var id, label string
	var domains, tools []string
	err := s.db.QueryRow(ctx,
		`SELECT id, label, domains, tools FROM mcp_tokens
		  WHERE token_hash = $1 AND revoked_at IS NULL`, HashToken(rawToken),
	).Scan(&id, &label, &domains, &tools)
	if err != nil {
		// Deliberately one message for "unknown" and "revoked": telling a caller
		// which of the two it hit would confirm that a token once existed.
		return Scope{}, TokenInfo{}, NewToolError("FORBIDDEN", "invalid or revoked token")
	}

	return Scope{Label: label, Domains: normaliseList(domains), Tools: normaliseList(tools)}, TokenInfo{ID: id, Label: label}, nil
}

// Touch records use. Best effort: a failed bookkeeping update must never fail
// the call the user actually made.
func (s *PGTokenStore) Touch(ctx context.Context, tokenID string) error {
	if tokenID == "" {
		return nil
	}
	_, err := s.db.Exec(ctx,
		`UPDATE mcp_tokens SET last_used_at = NOW(), request_count = request_count + 1 WHERE id = $1`,
		tokenID)
	return err
}

// normaliseList turns an empty pg array into a nil allow-list, which Scope reads
// as "everything".
func normaliseList(values []string) []string {
	out := []string{}
	for _, v := range values {
		if t := strings.TrimSpace(v); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
