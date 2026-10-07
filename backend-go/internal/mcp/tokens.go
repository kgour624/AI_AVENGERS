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
	var domains, expertIDs, tools []string
	err := s.db.QueryRow(ctx,
		`SELECT id, label, domains, expert_ids, tools FROM mcp_tokens
		  WHERE token_hash = $1 AND revoked_at IS NULL`, HashToken(rawToken),
	).Scan(&id, &label, &domains, &expertIDs, &tools)
	if err != nil {
		// Deliberately one message for "unknown" and "revoked": telling a caller
		// which of the two it hit would confirm that a token once existed.
		return Scope{}, TokenInfo{}, NewToolError("FORBIDDEN", "invalid or revoked token")
	}

	return Scope{Label: label, Domains: normaliseList(domains), ExpertIDs: normaliseList(expertIDs), Tools: normaliseList(tools), TokenID: id}, TokenInfo{ID: id, Label: label}, nil
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

// TokenRecord is one token as an admin sees it. The hash is never included: an
// admin needs to recognise and revoke a token, not to read the credential.
type TokenRecord struct {
	ID           string   `json:"id"`
	Label        string   `json:"label"`
	Domains      []string `json:"domains"`
	ExpertIDs    []string `json:"expert_ids"`
	Tools        []string `json:"tools"`
	Revoked      bool     `json:"revoked"`
	LastUsedAt   *string  `json:"last_used_at,omitempty"`
	RequestCount int64    `json:"request_count"`
	CreatedAt    string   `json:"created_at"`
}

// Create mints a token for a user and returns the PLAINTEXT once.
//
// WHY the plaintext is returned here and stored nowhere: the caller shows it to
// the person who asked for it and then forgets it. Only the hash is persisted,
// so a lost token is replaced, never recovered — and a database leak yields no
// working credential.
func (s *PGTokenStore) Create(ctx context.Context, userID, label string, domains, expertIDs, tools []string) (string, TokenRecord, error) {
	raw, err := NewToken()
	if err != nil {
		return "", TokenRecord{}, err
	}
	var rec TokenRecord
	var lastUsed *string
	err = s.db.QueryRow(ctx,
		`INSERT INTO mcp_tokens (user_id, token_hash, label, domains, expert_ids, tools)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 RETURNING id, label, domains, expert_ids, tools, revoked_at IS NOT NULL, last_used_at::text, request_count, created_at::text`,
		userID, HashToken(raw), label, normaliseOrEmpty(domains), normaliseOrEmpty(expertIDs), normaliseOrEmpty(tools),
	).Scan(&rec.ID, &rec.Label, &rec.Domains, &rec.ExpertIDs, &rec.Tools, &rec.Revoked, &lastUsed, &rec.RequestCount, &rec.CreatedAt)
	if err != nil {
		return "", TokenRecord{}, fmt.Errorf("mcp: create token: %w", err)
	}
	rec.LastUsedAt = lastUsed
	return raw, rec, nil
}

// List returns every token, newest first, for the admin screen.
func (s *PGTokenStore) List(ctx context.Context) ([]TokenRecord, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, label, domains, expert_ids, tools, revoked_at IS NOT NULL, last_used_at::text, request_count, created_at::text
		   FROM mcp_tokens
		  ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("mcp: list tokens: %w", err)
	}
	defer rows.Close()

	out := []TokenRecord{}
	for rows.Next() {
		var rec TokenRecord
			var lastUsed *string
			if err := rows.Scan(&rec.ID, &rec.Label, &rec.Domains, &rec.ExpertIDs, &rec.Tools, &rec.Revoked, &lastUsed, &rec.RequestCount, &rec.CreatedAt); err != nil {
			return nil, fmt.Errorf("mcp: scan token: %w", err)
		}
		rec.LastUsedAt = lastUsed
		out = append(out, rec)
	}
	return out, rows.Err()
}

// Revoke marks a token unusable. The row is kept so the audit trail keeps
// pointing at a readable record instead of a dangling id.
func (s *PGTokenStore) Revoke(ctx context.Context, id string) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE mcp_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("mcp: revoke token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return NewToolError("NOT_FOUND", "no active token with that id")
	}
	return nil
}

// Delete permanently removes a token from the database (hard delete).
// Works for both active and already-revoked tokens. After this the row is
// gone completely — use when you want clean DB, not just soft-revoke.
func (s *PGTokenStore) Delete(ctx context.Context, id string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM mcp_tokens WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("mcp: delete token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return NewToolError("NOT_FOUND", "no token with that id")
	}
	return nil
}

// normaliseOrEmpty keeps a pg array column happy: Postgres wants '{}' rather
// than NULL for an empty list.
func normaliseOrEmpty(values []string) []string {
	out := []string{}
	for _, v := range values {
		if t := strings.TrimSpace(v); t != "" {
			out = append(out, t)
		}
	}
	return out
}
