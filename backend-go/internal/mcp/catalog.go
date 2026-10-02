package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Expert is the read-only view of an expert that MCP needs. Deliberately NOT
// the admin API model: MCP must never expose admin-only fields (training state,
// job history, costs), so it reads exactly these columns.
type Expert struct {
	ID               string
	Name             string
	Slug             string
	Domain           string
	Description      string
	ReasoningCharter string
}

// Catalog is how the tools reach expert knowledge.
//
// WHY an input-only interface: the tools depend on this, not on a database
// handle, so a test can hand them a fixed list of experts and no Postgres is
// needed to prove a tool works.
type Catalog interface {
	ListExperts(ctx context.Context) ([]Expert, error)
	ExpertFor(ctx context.Context, domain, slugOrName string) (Expert, bool, error)
	ExpertByID(ctx context.Context, id string) (Expert, bool, error)
}

// PGCatalog reads experts straight from Postgres. Read-only by construction:
// there is no INSERT, UPDATE or DELETE in this file, so an MCP bug can never
// write to the product's data.
type PGCatalog struct {
	db *pgxpool.Pool
}

// NewPGCatalog wraps a pool.
func NewPGCatalog(db *pgxpool.Pool) *PGCatalog { return &PGCatalog{db: db} }

const expertColumns = `id, name, slug, domain, COALESCE(description, ''), COALESCE(reasoning_charter, '')`

const expertWhere = ` WHERE deleted_at IS NULL AND is_active = TRUE`

// ListExperts returns every active expert ordered by domain then name, so a
// caller always sees the same roster in the same order.
func (c *PGCatalog) ListExperts(ctx context.Context) ([]Expert, error) {
	rows, err := c.db.Query(ctx,
		`SELECT `+expertColumns+` FROM experts`+expertWhere+` ORDER BY domain ASC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("mcp catalog: list experts: %w", err)
	}
	defer rows.Close()

	out := []Expert{}
	for rows.Next() {
		var e Expert
		if err := rows.Scan(&e.ID, &e.Name, &e.Slug, &e.Domain, &e.Description, &e.ReasoningCharter); err != nil {
			return nil, fmt.Errorf("mcp catalog: scan expert: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ExpertFor finds one expert by slug OR name inside a domain.
//
// WHY both spellings: a client writing "master system design" or
// "MASTER-SYSTEM-DESIGN" is naming the same expert, and forcing the caller to
// know our slug format would just produce failed calls. Domain matching uses the
// same normalisation as the scope check, so an allowed domain is always
// reachable — access rules and lookups cannot disagree.
//
// The roster is read and matched in Go rather than filtered in SQL on purpose:
// the expert count per deployment is small, and this keeps "is this the same
// domain?" identical everywhere instead of re-implementing the rule in SQL.
func (c *PGCatalog) ExpertFor(ctx context.Context, domain, slugOrName string) (Expert, bool, error) {
	all, err := c.ListExperts(ctx)
	if err != nil {
		return Expert{}, false, err
	}

	wantDomain := NormDomainKey(domain)
	wantExpert := strings.ToLower(strings.TrimSpace(slugOrName))

	if wantExpert != "" {
		for _, e := range all {
			if wantDomain != "" && NormDomainKey(e.Domain) != wantDomain {
				continue
			}
			if strings.ToLower(e.Slug) == wantExpert ||
				strings.ToLower(e.Name) == wantExpert ||
				NormDomainKey(e.Name) == NormDomainKey(wantExpert) {
				return e, true, nil
			}
		}
	}
	// Domain only (no expert named): the domain's first expert answers.
	if wantExpert == "" && wantDomain != "" {
		for _, e := range all {
			if NormDomainKey(e.Domain) == wantDomain {
				return e, true, nil
			}
		}
	}
	return Expert{}, false, nil
}

// ExpertByID finds one expert by its stable UUID.
//
// WHY this exists: list_experts now prints id:`<uuid>` so a Claude roundtrip
// can copy-paste the ID without hallucinating it. Domain+slug is kept as an
// alias, but ID is the single source of truth.
func (c *PGCatalog) ExpertByID(ctx context.Context, id string) (Expert, bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Expert{}, false, nil
	}
	all, err := c.ListExperts(ctx)
	if err != nil {
		return Expert{}, false, err
	}
	for _, e := range all {
		if e.ID == id {
			return e, true, nil
		}
	}
	return Expert{}, false, nil
}
