package category

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// SectionType is one of the allowed template_schema section "type" values.
type SectionType string

const (
	SectionTypeProse     SectionType = "prose"
	SectionTypeCode      SectionType = "code"
	SectionTypeTestCases SectionType = "test_cases"
)

// ValidSectionTypes is used by admin handlers to validate template_schema
// before it is ever written to the DB. Kept here (not in admin package)
// so the registry and the validator agree on the same source of truth.
var ValidSectionTypes = map[SectionType]bool{
	SectionTypeProse:     true,
	SectionTypeCode:      true,
	SectionTypeTestCases: true,
}

// TemplateSection is one entry in a category's template_schema.sections array.
type TemplateSection struct {
	Key      string      `json:"key"`
	Label    string      `json:"label"`
	Type     SectionType `json:"type"`
	Required bool        `json:"required"`
	// Description (2026-09-08 RCA fix): optional admin-authored guidance
	// on WHAT CONTENT belongs in this section - distinct from Label
	// (a display name, e.g. "Pattern") which tells the model nothing
	// about what to actually write. Root cause of a real production
	// issue: buildStructuredPrompt (chinawall/template.go) previously
	// gave the model only key/type/label for each section, with zero
	// semantic guidance - unlike the flat-text path, which explicitly
	// instructs "Structure your answer as: Approach -> Code -> Complexity".
	// Without this, the model produced thin, citation-heavy, low-content
	// prose sections (a DSA expert answering almost entirely via
	// citations, near-empty Pattern/Idea/Walkthrough text). Empty string
	// is valid - buildStructuredPrompt falls back to a built-in
	// guidance-by-key table for common labels (pattern/idea/walkthrough/
	// approach/complexity/etc.) when this is unset, so existing
	// categories created before this field existed are NOT silently
	// degraded - see buildStructuredPrompt's sectionGuidance().
	Description string `json:"description,omitempty"`
}

// Template is one named answer format inside a category (e.g. "Code",
// "Approach"). A category may hold several so the SAME expert can answer in a
// different shape per question — the admin's "sometimes code, sometimes
// approach" requirement.
type Template struct {
	Name     string            `json:"name"`
	Sections []TemplateSection `json:"sections"`
}

// TemplateSchema is the full JSONB shape stored on expert_categories.template_schema.
//
// Two shapes are accepted, and BOTH are valid:
//   - Sections: the original single template (migration 010). Kept so existing
//     categories and the admin UI keep working untouched.
//   - Templates + Default: named variants. When present, Templates wins and
//     Sections is treated as legacy/unused.
//
// Parsed from one JSONB column, so adding variants needs NO migration.
type TemplateSchema struct {
	Sections  []TemplateSection `json:"sections,omitempty"`
	Templates []Template        `json:"templates,omitempty"`
	// Default names the variant used when the caller does not pick one.
	Default string `json:"default,omitempty"`
}

// ResolveSections returns the sections to answer with, and the template name
// actually used ("" for the legacy single template).
//
// Resolution order — first match wins, and a miss is never an error (an empty
// result simply means flat text, the same safe default as "no category"):
//  1. the requested template name, case-insensitively
//  2. the schema's declared Default
//  3. the first variant that actually has sections
//  4. the legacy Sections array
func (c *Category) ResolveSections(requested string) ([]TemplateSection, string) {
	if c == nil {
		return nil, ""
	}
	if len(c.TemplateSchema.Templates) > 0 {
		want := strings.TrimSpace(requested)
		if want == "" {
			want = strings.TrimSpace(c.TemplateSchema.Default)
		}
		for _, t := range c.TemplateSchema.Templates {
			if strings.EqualFold(strings.TrimSpace(t.Name), want) && len(t.Sections) > 0 {
				return t.Sections, t.Name
			}
		}
		for _, t := range c.TemplateSchema.Templates {
			if len(t.Sections) > 0 {
				return t.Sections, t.Name
			}
		}
	}
	return c.TemplateSchema.Sections, ""
}

// TemplateNames lists the available variant names (for the admin/chat picker).
func (c *Category) TemplateNames() []string {
	if c == nil {
		return nil
	}
	out := make([]string, 0, len(c.TemplateSchema.Templates))
	for _, t := range c.TemplateSchema.Templates {
		if strings.TrimSpace(t.Name) != "" {
			out = append(out, t.Name)
		}
	}
	return out
}

// Category is the in-memory representation of an expert_categories row.
type Category struct {
	ID                     uuid.UUID      `json:"id"`
	Name                   string         `json:"name"`
	Slug                   string         `json:"slug"`
	Description            string         `json:"description"`
	TemplateSchema         TemplateSchema `json:"template_schema"`
	DefaultLanguage        string         `json:"default_language"`
	AskStructurePermission bool           `json:"ask_structure_permission"`
	CreatedBy              *uuid.UUID     `json:"created_by,omitempty"`
}

// Registry is the in-memory cache of expert categories.
//
// LIFECYCLE (mirrors chinawall.DomainRegistry exactly — see
// backend-go/internal/chinawall/domain_registry.go, which this file is
// intentionally analogous to, per CATEGORY_TEMPLATE_HANDOFF.md CT-L1
// which requires domain_registry.go itself to stay untouched):
//  1. NewRegistry() — creates registry with DB connection
//  2. Init(ctx) — loads all categories into memory (no seeding here;
//     the one seed category is inserted by migration 010 itself, not
//     by application code — unlike chinawall's DefaultProfiles which
//     are seeded from Go code. Categories are fully admin-owned data.)
//  3. Get(id) / GetBySlug(slug) — O(1) lookup
//  4. Upsert(ctx, category) — DB write + cache update
//  5. Reload(ctx) — force full reload from DB
//
// THREAD SAFETY: RWMutex — multiple concurrent reads, exclusive writes.
// Same rationale as DomainRegistry: multiple experts may be answered
// concurrently, and category lookup must not block the read path.
type Registry struct {
	db     *pgxpool.Pool
	logger *zap.Logger

	mu     sync.RWMutex
	byID   map[uuid.UUID]*Category
	bySlug map[string]*Category // key: lowercase slug
}

// NewRegistry creates a category registry. Call Init() before use.
func NewRegistry(db *pgxpool.Pool, logger *zap.Logger) *Registry {
	return &Registry{
		db:     db,
		logger: logger,
		byID:   make(map[uuid.UUID]*Category),
		bySlug: make(map[string]*Category),
	}
}

// Init loads all categories from DB into memory.
// Must be called once at application startup before serving requests.
//
// WHY no seeding here (unlike chinawall.DomainRegistry.Init, which
// seeds DefaultProfiles): the one default category ("coding") is
// seeded directly by migration 010 in SQL, because it is real product
// data (a template an admin can see and edit from day one), not a
// code-level safety default like chinawall.BaseProfile. Seeding it
// from Go code here would create two sources of truth for the same
// row and risk the exact "admin override silently reverted" failure
// mode domain_registry.go's seedDefault() ON CONFLICT DO NOTHING logic
// exists to prevent — simpler to only ever write it once, in SQL.
func (r *Registry) Init(ctx context.Context) error {
	if err := r.loadAll(ctx); err != nil {
		return fmt.Errorf("category registry: init: %w", err)
	}
	r.logger.Info("category registry: initialized",
		zap.Int("categories_loaded", len(r.byID)),
	)
	return nil
}

// Get returns the Category for the given id, or nil if not found.
// O(1) — hot path, called whenever an expert with a category_id answers.
//
// WHY return nil (not BaseProfile-style fallback like chinawall.Get):
// unlike chinawall.DomainRegistry.Get (which must always return a
// usable *DomainProfile because every question needs a China Wall
// policy), a missing category here is a legitimate, common state
// (CT-L2: category_id is nullable and most experts may have none).
// Callers must explicitly branch on nil — "no category" and "unknown
// category id" both correctly fall back to flat-text behavior with no
// special-casing needed at the call site.
func (r *Registry) Get(id uuid.UUID) *Category {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byID[id]
}

// GetBySlug returns the Category for the given slug (case-insensitive),
// or nil if not found. Used by admin UI lookups and by any future
// slug-based reference (e.g. retrofit scripts, tests).
func (r *Registry) GetBySlug(slug string) *Category {
	key := strings.ToLower(strings.TrimSpace(slug))
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.bySlug[key]
}

// List returns all cached categories. Used by the admin "list categories"
// endpoint so it can serve from cache instead of hitting the DB on every
// request (categories change rarely — admin-managed config, not hot data).
func (r *Registry) List() []*Category {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*Category, 0, len(r.byID))
	for _, c := range r.byID {
		result = append(result, c)
	}
	return result
}

// Upsert writes a category to DB and updates the in-memory cache.
// Called by admin panel create/update handlers.
//
// WHY the caller passes an *existing* ID for updates and a nil/zero ID
// is treated as "insert new": mirrors the existing admin_handler.go
// pattern (CreateExpert does its own INSERT ... RETURNING id; UpdateExpert
// does field-by-field UPDATEs) rather than a single ORM-style save —
// kept deliberately simple here since admin_handler.go's category CRUD
// handlers (CT-A3) call the DB directly for create/update and use this
// registry only for the read path (Get/GetBySlug/List) plus explicit
// cache invalidation after a write. Upsert exists for completeness and
// for any future caller that wants a single write+cache-update call.
func (r *Registry) Upsert(ctx context.Context, cat *Category) error {
	schemaJSON, err := json.Marshal(cat.TemplateSchema)
	if err != nil {
		return fmt.Errorf("category registry: marshal template_schema: %w", err)
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO expert_categories
			(id, name, slug, description, template_schema, default_language, ask_structure_permission, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			slug = EXCLUDED.slug,
			description = EXCLUDED.description,
			template_schema = EXCLUDED.template_schema,
			default_language = EXCLUDED.default_language,
			ask_structure_permission = EXCLUDED.ask_structure_permission,
			updated_at = NOW()
	`, cat.ID, cat.Name, cat.Slug, cat.Description, string(schemaJSON),
		cat.DefaultLanguage, cat.AskStructurePermission, cat.CreatedBy)
	if err != nil {
		return fmt.Errorf("category registry: upsert slug=%s: %w", cat.Slug, err)
	}

	// Update cache immediately — no TTL, always consistent with DB.
	r.mu.Lock()
	r.byID[cat.ID] = cat
	r.bySlug[strings.ToLower(cat.Slug)] = cat
	r.mu.Unlock()

	r.logger.Info("category registry: upserted", zap.String("slug", cat.Slug))
	return nil
}

// Reload reloads all categories from DB into memory.
// Call this after any direct-SQL write to expert_categories (e.g. from
// admin_handler.go's CRUD handlers) so the cache does not go stale.
func (r *Registry) Reload(ctx context.Context) error {
	if err := r.loadAll(ctx); err != nil {
		return fmt.Errorf("category registry: reload: %w", err)
	}
	r.logger.Info("category registry: reloaded", zap.Int("categories", len(r.byID)))
	return nil
}

// loadAll reads all rows from expert_categories and populates the
// in-memory maps. Atomic swap — readers see either the old maps or the
// new maps, never a partial rebuild (same pattern as
// chinawall.DomainRegistry.loadAll).
func (r *Registry) loadAll(ctx context.Context) error {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, slug, COALESCE(description,''), template_schema,
		       default_language, ask_structure_permission, created_by
		FROM expert_categories
	`)
	if err != nil {
		return fmt.Errorf("query expert_categories: %w", err)
	}
	defer rows.Close()

	newByID := make(map[uuid.UUID]*Category)
	newBySlug := make(map[string]*Category)

	for rows.Next() {
		var c Category
		var schemaJSON []byte
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Slug, &c.Description, &schemaJSON,
			&c.DefaultLanguage, &c.AskStructurePermission, &c.CreatedBy,
		); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}

		// A corrupt/empty template_schema must not take down the whole
		// registry load — fall back to an empty schema (== flat-text
		// behavior for any expert in this category, same safe default
		// as "no category at all" per CT-L2), log, and continue with
		// the rest of the rows. Mirrors chinawall.DomainRegistry.loadAll's
		// "skip corrupt row, use BaseProfile" non-fatal handling.
		if len(schemaJSON) > 0 {
			if err := json.Unmarshal(schemaJSON, &c.TemplateSchema); err != nil {
				r.logger.Warn("category registry: corrupt template_schema, using empty schema",
					zap.String("slug", c.Slug),
					zap.Error(err),
				)
				c.TemplateSchema = TemplateSchema{Sections: nil}
			}
		}

		newByID[c.ID] = &c
		newBySlug[strings.ToLower(c.Slug)] = &c
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}

	r.mu.Lock()
	r.byID = newByID
	r.bySlug = newBySlug
	r.mu.Unlock()

	return nil
}
