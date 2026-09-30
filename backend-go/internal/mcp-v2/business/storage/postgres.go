// Package storage — Storage under Business domain (Ultimate Go §7).
// Business defines Storer interface; this postgres adapter implements it.
// Implements business.Storer — no import of App (mcpv2), only Business (imports down only).
package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"ai_avengers/backend/internal/mcp-v2/business"
)

// PostgresAdapter implements business.Storer — Storage Model -> Business via parse() ( §7 Three Models + parse).
type PostgresAdapter struct{ pool *pgxpool.Pool }

func NewPostgresAdapter(pool *pgxpool.Pool) *PostgresAdapter { return &PostgresAdapter{pool: pool} }

// dbExpert is Storage Model (native pgx types) — business.Expert is core.
type dbExpert struct {
	id, name, slug, charter string
}

func (d dbExpert) parse() (business.Expert, error) {
	if d.id == "" {
		return business.Expert{}, fmt.Errorf("invalid expert id")
	}
	return business.Expert{ID: d.id, Name: d.name, Slug: d.slug, Charter: d.charter}, nil
}

func (p *PostgresAdapter) ListExperts(ctx context.Context) ([]business.Expert, error) {
	rows, err := p.pool.Query(ctx, `SELECT id::text, name, slug, charter FROM experts WHERE is_active = true ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list experts: %w", err)
	}
	defer rows.Close()
	var out []business.Expert
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

func (p *PostgresAdapter) GetExpert(ctx context.Context, expertID string) (business.Expert, error) {
	var d dbExpert
	err := p.pool.QueryRow(ctx, `SELECT id::text, name, slug, charter FROM experts WHERE id=$1`, expertID).Scan(&d.id, &d.name, &d.slug, &d.charter)
	if err != nil {
		return business.Expert{}, fmt.Errorf("get expert: %w", err)
	}
	return d.parse()
}

// pgvector cosine — real embedding search (production end-to-end)
func (p *PostgresAdapter) SearchChunks(ctx context.Context, expertID string, embedding []float32, limit int, cursor string) ([]business.Chunk, string, int, error) {
	_ = cursor
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	if len(embedding) == 0 {
		rows, err := p.pool.Query(ctx, `SELECT id::text, expert_id::text, content, 0.0::real as score FROM course_chunks WHERE expert_id=$1 LIMIT $2`, expertID, limit)
		if err != nil {
			return nil, "", 0, fmt.Errorf("search chunks fallback: %w", err)
		}
		defer rows.Close()
		var chunks []business.Chunk
		for rows.Next() {
			var c business.Chunk
			if err := rows.Scan(&c.ID, &c.ExpertID, &c.Text, &c.Score); err != nil {
				return nil, "", 0, err
			}
			chunks = append(chunks, c)
		}
		return chunks, "", len(chunks), rows.Err()
	}
	vecStr := pgVectorString(embedding)
	rows, err := p.pool.Query(ctx, `SELECT id::text, expert_id::text, content, 1 - (embedding <=> $2::vector) as score FROM course_chunks WHERE expert_id=$1 ORDER BY embedding <=> $2::vector LIMIT $3`, expertID, vecStr, limit)
	if err != nil {
		return nil, "", 0, fmt.Errorf("search chunks vector: %w", err)
	}
	defer rows.Close()
	var chunks []business.Chunk
	for rows.Next() {
		var c business.Chunk
		if err := rows.Scan(&c.ID, &c.ExpertID, &c.Text, &c.Score); err != nil {
			return nil, "", 0, err
		}
		chunks = append(chunks, c)
	}
	return chunks, "", len(chunks), rows.Err()
}

func pgVectorString(v []float32) string {
	b := make([]byte, 0, len(v)*8)
	b = append(b, '[')
	for i, f := range v {
		if i > 0 {
			b = append(b, ',')
		}
		b = fmt.Appendf(b, "%f", f)
	}
	b = append(b, ']')
	return string(b)
}

func (p *PostgresAdapter) SearchRepoChunks(ctx context.Context, expertID string, embedding []float32, limit int) ([]business.RepoChunk, error) {
	if limit <= 0 {
		limit = 10
	}
	if len(embedding) == 0 {
		rows, err := p.pool.Query(ctx, `SELECT rc.id::text, rc.file_path, rc.chunk_text, 0.0::real FROM mcp_repo_chunks rc JOIN mcp_repo_connections conn ON conn.id=rc.repo_connection_id WHERE conn.expert_id=$1 LIMIT $2`, expertID, limit)
		if err != nil {
			return nil, fmt.Errorf("search repo fallback: %w", err)
		}
		defer rows.Close()
		var out []business.RepoChunk
		for rows.Next() {
			var r business.RepoChunk
			if err := rows.Scan(&r.ID, &r.FilePath, &r.Text, &r.Score); err != nil {
				return nil, err
			}
			out = append(out, r)
		}
		return out, rows.Err()
	}
	vecStr := pgVectorString(embedding)
	rows, err := p.pool.Query(ctx, `SELECT rc.id::text, rc.file_path, rc.chunk_text, 1 - (rc.embedding <=> $2::vector) as score FROM mcp_repo_chunks rc JOIN mcp_repo_connections conn ON conn.id=rc.repo_connection_id WHERE conn.expert_id=$1 ORDER BY rc.embedding <=> $2::vector LIMIT $3`, expertID, vecStr, limit)
	if err != nil {
		return nil, fmt.Errorf("search repo vector: %w", err)
	}
	defer rows.Close()
	var out []business.RepoChunk
	for rows.Next() {
		var r business.RepoChunk
		if err := rows.Scan(&r.ID, &r.FilePath, &r.Text, &r.Score); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (p *PostgresAdapter) InsertUsageLog(ctx context.Context, row business.UsageLog) error {
	_, err := p.pool.Exec(ctx, `INSERT INTO mcp_usage_log (expert_id, platform, provider, model, tier, input_tokens, output_tokens, total_tokens, cost_usd, generic_used, generic_percent, duration_ms) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, row.ExpertID, row.Platform, row.Provider, row.Model, row.Tier, row.InputTokens, row.OutputTokens, row.TotalTokens, row.CostUSD, row.GenericUsed, row.GenericPercent, row.DurationMs)
	return err
}

func (p *PostgresAdapter) GetExpertLimits(ctx context.Context, expertID, platform string) (business.ExpertLimits, error) {
	var l business.ExpertLimits
	err := p.pool.QueryRow(ctx, `SELECT expert_id::text, platform, daily_token_limit, monthly_token_limit, current_daily, current_monthly FROM mcp_expert_limits WHERE expert_id=$1 AND platform=$2`, expertID, platform).Scan(&l.ExpertID, &l.Platform, &l.DailyLimit, &l.MonthlyLimit, &l.CurrentDaily, &l.CurrentMonthly)
	if err != nil {
		return business.ExpertLimits{ExpertID: expertID, Platform: platform, DailyLimit: 1000000, MonthlyLimit: 20000000}, nil
	}
	return l, nil
}

func (p *PostgresAdapter) IncLimits(ctx context.Context, expertID, platform string, tokens int) error {
	_, err := p.pool.Exec(ctx, `INSERT INTO mcp_expert_limits (expert_id, platform, current_daily, current_monthly) VALUES ($1,$2,$3,$3) ON CONFLICT (expert_id, platform) DO UPDATE SET current_daily = mcp_expert_limits.current_daily + $3, current_monthly = mcp_expert_limits.current_monthly + $3, updated_at = NOW()`, expertID, platform, tokens)
	return err
}