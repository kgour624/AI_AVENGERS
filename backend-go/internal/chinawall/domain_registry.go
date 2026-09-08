package chinawall

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// DomainRegistry is the in-memory cache of domain profiles.
//
// LIFECYCLE:
//   1. NewDomainRegistry() — creates registry with DB connection
//   2. Init(ctx) — seeds defaults, loads all profiles into memory
//   3. Get(domain) — O(1) lookup per question (hot path)
//   4. Upsert(ctx, profile) — DB write + cache update
//   5. Invalidate(domain) — force reload from DB
//
// THREAD SAFETY:
//   RWMutex — multiple concurrent reads, exclusive writes.
//   WHY: Multiple experts answer simultaneously. Read path must not block.
type DomainRegistry struct {
	db     *pgxpool.Pool
	logger *zap.Logger

	mu       sync.RWMutex
	profiles map[string]*DomainProfile // key: lowercase domain name
}

// NewDomainRegistry creates a registry. Call Init() before use.
func NewDomainRegistry(db *pgxpool.Pool, logger *zap.Logger) *DomainRegistry {
	return &DomainRegistry{
		db:       db,
		logger:   logger,
		profiles: make(map[string]*DomainProfile),
	}
}

// Init seeds default profiles and loads all profiles from DB into memory.
// Must be called once at application startup before serving requests.
func (r *DomainRegistry) Init(ctx context.Context) error {
	// Ensure table exists
	if err := r.ensureTable(ctx); err != nil {
		return fmt.Errorf("domain_registry: ensure table: %w", err)
	}

	// Seed defaults (INSERT ... ON CONFLICT DO NOTHING)
	// WHY ON CONFLICT DO NOTHING: admin overrides must not be overwritten on restart.
	for _, profile := range DefaultProfiles {
		if err := r.seedDefault(ctx, profile); err != nil {
			r.logger.Warn("domain_registry: seed default failed",
				zap.String("domain", profile.Domain),
				zap.Error(err),
			)
			// Non-fatal: continue with other defaults
		}
	}

	// Load all profiles from DB into memory
	if err := r.loadAll(ctx); err != nil {
		return fmt.Errorf("domain_registry: load all: %w", err)
	}

	r.logger.Info("domain_registry: initialized",
		zap.Int("profiles_loaded", len(r.profiles)),
	)
	return nil
}

// Get returns the DomainProfile for the given domain.
// Falls back to BaseProfile if domain is not found.
// O(1) — hot path, called for every question.
func (r *DomainRegistry) Get(domain string) *DomainProfile {
	key := strings.ToLower(strings.TrimSpace(domain))

	r.mu.RLock()
	profile, ok := r.profiles[key]
	r.mu.RUnlock()

	if !ok {
		// Unknown domain — use strict base profile as safety net.
		// WHY: Better to refuse than to hallucinate for unknown domains.
		r.logger.Debug("domain_registry: unknown domain, using base profile",
			zap.String("domain", domain),
		)
		return BaseProfile
	}
	return profile
}

// Upsert writes a profile to DB and updates the in-memory cache.
// Called by:
//   - Admin panel when updating domain config
//   - AI update path when conversation patterns suggest rule changes
func (r *DomainRegistry) Upsert(ctx context.Context, profile *DomainProfile) error {
	key := strings.ToLower(strings.TrimSpace(profile.Domain))

	configJSON, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("domain_registry: marshal profile: %w", err)
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO domain_profiles (domain, config, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (domain) DO UPDATE
		  SET config = EXCLUDED.config,
		      updated_at = NOW()
	`, key, configJSON)
	if err != nil {
		return fmt.Errorf("domain_registry: upsert domain=%s: %w", key, err)
	}

	// Update cache immediately — no TTL, always consistent with DB.
	r.mu.Lock()
	r.profiles[key] = profile
	r.mu.Unlock()

	r.logger.Info("domain_registry: upserted", zap.String("domain", key))
	return nil
}

// Invalidate removes a domain from cache, forcing next Get() to fall back
// to BaseProfile until loadAll() is called again.
// Use when you want to force a reload without restarting.
func (r *DomainRegistry) Invalidate(domain string) {
	key := strings.ToLower(strings.TrimSpace(domain))
	r.mu.Lock()
	delete(r.profiles, key)
	r.mu.Unlock()
	r.logger.Info("domain_registry: invalidated", zap.String("domain", key))
}

// Reload reloads all profiles from DB into memory.
// Can be called from admin panel to pick up external DB changes.
func (r *DomainRegistry) Reload(ctx context.Context) error {
	if err := r.loadAll(ctx); err != nil {
		return fmt.Errorf("domain_registry: reload: %w", err)
	}
	r.logger.Info("domain_registry: reloaded", zap.Int("profiles", len(r.profiles)))
	return nil
}

// ensureTable creates the domain_profiles table if it does not exist.
func (r *DomainRegistry) ensureTable(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS domain_profiles (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			domain     VARCHAR(100) NOT NULL UNIQUE,
			config     JSONB        NOT NULL,
			created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create table domain_profiles: %w", err)
	}
	return nil
}

// seedDefault inserts a default profile if it does not already exist.
// ON CONFLICT DO NOTHING — admin overrides are preserved.
func (r *DomainRegistry) seedDefault(ctx context.Context, profile *DomainProfile) error {
	configJSON, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO domain_profiles (domain, config)
		VALUES ($1, $2)
		ON CONFLICT (domain) DO NOTHING
	`, strings.ToLower(profile.Domain), configJSON)
	return err
}

// loadAll reads all rows from domain_profiles and populates the in-memory map.
func (r *DomainRegistry) loadAll(ctx context.Context) error {
	rows, err := r.db.Query(ctx, `SELECT domain, config FROM domain_profiles`)
	if err != nil {
		return fmt.Errorf("query domain_profiles: %w", err)
	}
	defer rows.Close()

	newMap := make(map[string]*DomainProfile)
	for rows.Next() {
		var domain string
		var configJSON []byte
		if err := rows.Scan(&domain, &configJSON); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}

		var profile DomainProfile
		if err := json.Unmarshal(configJSON, &profile); err != nil {
			r.logger.Warn("domain_registry: corrupt profile, skipping",
				zap.String("domain", domain),
				zap.Error(err),
			)
			continue // Non-fatal: skip corrupt row, use BaseProfile for this domain
		}

		newMap[strings.ToLower(domain)] = &profile
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}

	// Atomic swap — readers see either old map or new map, never partial.
	r.mu.Lock()
	r.profiles = newMap
	r.mu.Unlock()

	return nil
}
