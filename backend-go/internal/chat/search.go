package chat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ChatSearchResult is one hit of the global (cross-project) chat search (#27).
//
// WHY it carries project name: the whole point of a global search is that the
// user may not know which project a chat lives in — a title alone is not
// enough to recognize the hit, so each row is labelled with its project.
type ChatSearchResult struct {
	ChatID         uuid.UUID `json:"chat_id"`
	ChatTitle      string    `json:"chat_title"`
	ProjectID      uuid.UUID `json:"project_id"`
	ProjectName    string    `json:"project_name"`
	MessageCount   int       `json:"message_count"`
	IsArchived     bool      `json:"is_archived"`
	UpdatedAt      time.Time `json:"updated_at"`
	TitleMatch     bool      `json:"title_match"`
	ContentMatches int       `json:"content_matches"`
}

// SearchChats finds the caller's chats across EVERY project whose title or any
// message matches the query, newest activity first.
//
// SCOPE / SECURITY:
//   - Always filtered by `chats.client_id = $1` — a user only ever sees their
//     own chats, never another client's.
//   - tenantID is the caller's C4 tenant (nil = global/admin). When set, the
//     chat's project must belong to that tenant (projects.tenant_id), the same
//     deterministic pre-filter C4 uses everywhere else — no post-hoc filtering.
//
// MATCH SEMANTICS: case-insensitive substring (ILIKE) over the chat title and
// message content. `%`/`_` in the user's query are escaped so they match
// literally. This is a substring scan, not the semantic chat_index — a global
// search must also hit chats that were never indexed. An index-backed variant
// (pg_trgm) is a follow-up; correctness first.
func (s *Service) SearchChats(
	ctx context.Context,
	clientID uuid.UUID,
	tenantID *uuid.UUID,
	query string,
	limit int,
) ([]ChatSearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []ChatSearchResult{}, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	pattern := "%" + escapeLike(query) + "%"

	rows, err := s.db.Query(ctx, `
		SELECT c.id, c.title, c.project_id, p.name, c.message_count, c.is_archived, c.updated_at,
		       (c.title ILIKE $2 ESCAPE '\') AS title_match,
		       (SELECT COUNT(*)::int FROM messages m
		         WHERE m.chat_id = c.id AND m.content ILIKE $2 ESCAPE '\') AS content_matches
		FROM chats c
		JOIN projects p ON p.id = c.project_id
		WHERE c.client_id = $1
		  AND ($3::uuid IS NULL OR p.tenant_id = $3)
		  AND (
		        c.title ILIKE $2 ESCAPE '\'
		     OR EXISTS (
		          SELECT 1 FROM messages m2
		           WHERE m2.chat_id = c.id AND m2.content ILIKE $2 ESCAPE '\'
		        )
		      )
		ORDER BY c.updated_at DESC
		LIMIT $4`,
		clientID, pattern, tenantID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search chats: %w", err)
	}
	defer rows.Close()

	out := []ChatSearchResult{}
	for rows.Next() {
		var r ChatSearchResult
		if err := rows.Scan(
			&r.ChatID, &r.ChatTitle, &r.ProjectID, &r.ProjectName,
			&r.MessageCount, &r.IsArchived, &r.UpdatedAt,
			&r.TitleMatch, &r.ContentMatches,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// escapeLike escapes LIKE/ILIKE wildcards so a user query containing %, _ or \
// is matched literally. Pure — used with `ESCAPE '\'` in the SQL above.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
