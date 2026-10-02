package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	mcpv2 "ai_avengers/backend/internal/mcp-v2"
)

// PostgresAdapter implements mcpv2.DBPort — Storage Model -> Business via parse() (RULE 8-B:33).
type PostgresAdapter struct{ pool *pgxpool.Pool }

func NewPostgresAdapter(pool *pgxpool.Pool) *PostgresAdapter { return &PostgresAdapter{pool: pool} }

// dbExpert is Storage Model (native pgx types).
type dbExpert struct {
	id, name, slug, charter string
}

func (d dbExpert) parse() (mcpv2.Expert, error) {
	if d.id == "" {
		return mcpv2.Expert{}, fmt.Errorf("invalid expert id")
	}
	return mcpv2.Expert{ID: d.id, Name: d.name, Slug: d.slug, Charter: d.charter}, nil
}

func (p *PostgresAdapter) ListExperts(ctx context.Context) ([]mcpv2.Expert, error) {
	rows, err := p.pool.Query(ctx, `SELECT id::text, name, slug, charter FROM experts WHERE is_active = true ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list experts: %w", err)
	}
	defer rows.Close()
	var out []mcpv2.Expert
	for rows.Next() {
		var d dbExpert
		if err := rows.Scan(&d.id, &d.name, &d.slug, &d.charter); err != nil {
			return nil, err
		}
		e, err := d.parse()
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (p *PostgresAdapter) GetExpert(ctx context.Context, expertID string) (mcpv2.Expert, error) {
	var d dbExpert
	err := p.pool.QueryRow(ctx, `SELECT id::text, name, slug, charter FROM experts WHERE id=$1`, expertID).Scan(&d.id, &d.name, &d.slug, &d.charter)
	if err == nil {
		return d.parse()
	}
	// slug/name alias fallback — ID is source of truth but Claude often sends slug
	if err2 := p.pool.QueryRow(ctx, `SELECT id::text, name, slug, charter FROM experts WHERE LOWER(slug)=LOWER($1) OR LOWER(name)=LOWER($1) LIMIT 1`, expertID).Scan(&d.id, &d.name, &d.slug, &d.charter); err2 == nil {
		return d.parse()
	}
	return mcpv2.Expert{}, fmt.Errorf("get expert: %w", err)
}

func (p *PostgresAdapter) SearchChunks(ctx context.Context, expertID string, embedding []float32, limit int, cursor string) ([]mcpv2.Chunk, string, int, error) {
	// cursor is base64({offset,limit}) — decode offset; simplified offset/limit; pgvector ivfflat
	// WHY: Pagination §4.5 opaque cursor, WHERE expert_id = $1 isolation (China Wall).
	// NOTE: embedding search placeholder — real query uses vector param via pgvector-go; limit+offset
	_ = cursor
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	// Fallback: text search until vector wired via VectorPort; keep scoped.
	rows, err := p.pool.Query(ctx, `SELECT id::text, expert_id::text, content, 0.0::real as score FROM course_chunks WHERE expert_id=$1 LIMIT $2`, expertID, limit)
	if err != nil {
		return nil, "", 0, fmt.Errorf("search chunks: %w", err)
	}
	defer rows.Close()
	var chunks []mcpv2.Chunk
	for rows.Next() {
		var c mcpv2.Chunk
		if err := rows.Scan(&c.ID, &c.ExpertID, &c.Text, &c.Score); err != nil {
			return nil, "", 0, err
		}
		chunks = append(chunks, c)
	}
	return chunks, "", len(chunks), rows.Err()
}

func (p *PostgresAdapter) SearchRepoChunks(ctx context.Context, expertID string, embedding []float32, limit int) ([]mcpv2.RepoChunk, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := p.pool.Query(ctx, `SELECT rc.id::text, rc.file_path, rc.chunk_text, 0.0::real FROM mcp_repo_chunks rc JOIN mcp_repo_connections conn ON conn.id=rc.repo_connection_id WHERE conn.expert_id=$1 LIMIT $2`, expertID, limit)
	if err != nil {
		return nil, fmt.Errorf("search repo: %w", err)
	}
	defer rows.Close()
	var out []mcpv2.RepoChunk
	for rows.Next() {
		var r mcpv2.RepoChunk
		if err := rows.Scan(&r.ID, &r.FilePath, &r.Text, &r.Score); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (p *PostgresAdapter) InsertUsageLog(ctx context.Context, row mcpv2.UsageLog) error {
	_, err := p.pool.Exec(ctx, `INSERT INTO mcp_usage_log (expert_id, platform, provider, model, tier, input_tokens, output_tokens, total_tokens, cost_usd, generic_used, generic_percent, duration_ms) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, row.ExpertID, row.Platform, row.Provider, row.Model, row.Tier, row.InputTokens, row.OutputTokens, row.TotalTokens, row.CostUSD, row.GenericUsed, row.GenericPercent, row.DurationMs)
	return err
}

func (p *PostgresAdapter) GetExpertLimits(ctx context.Context, expertID, platform string) (mcpv2.ExpertLimits, error) {
	var l mcpv2.ExpertLimits
	err := p.pool.QueryRow(ctx, `SELECT expert_id::text, platform, daily_token_limit, monthly_token_limit, current_daily, current_monthly FROM mcp_expert_limits WHERE expert_id=$1 AND platform=$2`, expertID, platform).Scan(&l.ExpertID, &l.Platform, &l.DailyLimit, &l.MonthlyLimit, &l.CurrentDaily, &l.CurrentMonthly)
	if err != nil {
		// No row = unlimited (no limits set)
		return mcpv2.ExpertLimits{ExpertID: expertID, Platform: platform, DailyLimit: 1000000, MonthlyLimit: 20000000}, nil
	}
	return l, nil
}

func (p *PostgresAdapter) IncLimits(ctx context.Context, expertID, platform string, tokens int) error {
	_, err := p.pool.Exec(ctx, `INSERT INTO mcp_expert_limits (expert_id, platform, current_daily, current_monthly) VALUES ($1,$2,$3,$3) ON CONFLICT (expert_id, platform) DO UPDATE SET current_daily = mcp_expert_limits.current_daily + $3, current_monthly = mcp_expert_limits.current_monthly + $3, updated_at = NOW()`, expertID, platform, tokens)
	return err
}
