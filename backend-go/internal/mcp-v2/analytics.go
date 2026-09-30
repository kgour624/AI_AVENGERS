package mcpv2

import (
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Analytics exposes v_mcp_daily_cost + rental + generic views ( §3.13 ).
type Analytics struct{ pool *pgxpool.Pool }

func NewAnalytics(pool *pgxpool.Pool) *Analytics { return &Analytics{pool: pool} }

type DailyCostRow struct {
	Day          time.Time `json:"day"`
	ExpertID     string    `json:"expertId"`
	Platform     string    `json:"platform"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	Tier         string    `json:"tier"`
	Calls        int       `json:"calls"`
	TotalTokens  int64     `json:"totalTokens"`
	GenericCalls int       `json:"genericCalls"`
	CostUSD      float64   `json:"costUsd"`
}

// GetDailyCost reads v_mcp_daily_cost with filters.
func (a *Analytics) GetDailyCost(ctx context.Context, expertID, platform string, from, to time.Time) ([]DailyCostRow, error) {
	rows, err := a.pool.Query(ctx, `SELECT day, expert_id::text, platform, provider, model, tier, calls, total_tokens, generic_calls, cost_usd FROM v_mcp_daily_cost WHERE ($1='' OR expert_id::text=$1) AND ($2='' OR platform=$2) AND day BETWEEN $3 AND $4 ORDER BY day DESC`, expertID, platform, from, to)
	if err != nil {
		return nil, fmt.Errorf("daily cost: %w", err)
	}
	defer rows.Close()
	var out []DailyCostRow
	for rows.Next() {
		var r DailyCostRow
		if err := rows.Scan(&r.Day, &r.ExpertID, &r.Platform, &r.Provider, &r.Model, &r.Tier, &r.Calls, &r.TotalTokens, &r.GenericCalls, &r.CostUSD); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ExportCSV writes daily cost to CSV writer (Admin export).
func (a *Analytics) ExportCSV(rows []DailyCostRow, w *csv.Writer) error {
	if err := w.Write([]string{"day","expert_id","platform","provider","model","tier","calls","total_tokens","generic_calls","cost_usd"}); err != nil {
		return err
	}
	for _, r := range rows {
		if err := w.Write([]string{r.Day.Format("2006-01-02"), r.ExpertID, r.Platform, r.Provider, r.Model, r.Tier, fmt.Sprint(r.Calls), fmt.Sprint(r.TotalTokens), fmt.Sprint(r.GenericCalls), fmt.Sprintf("%.6f", r.CostUSD)}); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}