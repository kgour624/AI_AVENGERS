package monitoring

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// CostMonitor tracks LLM costs and alerts when budget thresholds are exceeded.
//
// WHY cost monitoring:
// OpenRouter charges per token. Without monitoring:
// - A bug could cause infinite LLM loops -> $1000s in minutes
// - No visibility into which experts cost most
// - No way to optimize spending
//
// Architecture doc: monthly_limit_usd = 1000, alert_threshold = 0.8
type CostMonitor struct {
	db           *pgxpool.Pool
	gateway      *gateway.ModelGateway
	monthlyLimit float64
	alertThreshold float64
	logger       *zap.Logger
}

// NewCostMonitor creates a new cost monitor.
func NewCostMonitor(
	db *pgxpool.Pool,
	gw *gateway.ModelGateway,
	monthlyLimit float64,
	alertThreshold float64,
	logger *zap.Logger,
) *CostMonitor {
	return &CostMonitor{
		db:             db,
		gateway:        gw,
		monthlyLimit:   monthlyLimit,
		alertThreshold: alertThreshold,
		logger:         logger,
	}
}

// WorkflowCostReport holds per-workflow cost data.
type WorkflowCostReport struct {
	WorkflowID      string  `json:"workflow_id"`
	CostSpentUSD    float64 `json:"cost_spent_usd"`
	CostBudgetUSD   float64 `json:"cost_budget_usd"`
	SoftLimitPct    float64 `json:"soft_limit_pct"`
	HardLimitPct    float64 `json:"hard_limit_pct"`
	SoftLimitHit    bool    `json:"soft_limit_hit"`
	HardLimitHit    bool    `json:"hard_limit_hit"`
	MessageCount    int     `json:"message_count"`
}

// GetWorkflowCost returns cost breakdown for a specific workflow.
// Uses SUM(messages.cost_usd) for accuracy in reporting.
func (m *CostMonitor) GetWorkflowCost(ctx context.Context, workflowID string) (*WorkflowCostReport, error) {
	var report WorkflowCostReport
	report.WorkflowID = workflowID

	// Get workflow budget fields
	err := m.db.QueryRow(ctx,
		`SELECT cost_spent_usd, cost_budget_usd, cost_soft_limit_pct, cost_hard_limit_pct
		 FROM workflows WHERE id = $1`,
		workflowID,
	).Scan(&report.CostSpentUSD, &report.CostBudgetUSD, &report.SoftLimitPct, &report.HardLimitPct)
	if err != nil {
		return nil, fmt.Errorf("get workflow cost: %w", err)
	}

	// Get message count for this workflow
	m.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM messages WHERE workflow_id = $1`,
		workflowID,
	).Scan(&report.MessageCount)

	// Check limits
	softThreshold := report.CostBudgetUSD * (report.SoftLimitPct / 100.0)
	hardThreshold := report.CostBudgetUSD * (report.HardLimitPct / 100.0)
	report.SoftLimitHit = report.CostSpentUSD >= softThreshold
	report.HardLimitHit = report.CostSpentUSD >= hardThreshold

	return &report, nil
}

// CheckWorkflowLimits checks if a workflow has hit its cost limits.
// Fast path: reads cost_spent_usd from workflows table (no aggregation).
// Called after every LLM call in a workflow context.
func (m *CostMonitor) CheckWorkflowLimits(ctx context.Context, workflowID string) (softHit, hardHit bool, err error) {
	var spent, budget, softPct, hardPct float64
	err = m.db.QueryRow(ctx,
		`SELECT cost_spent_usd, cost_budget_usd, cost_soft_limit_pct, cost_hard_limit_pct
		 FROM workflows WHERE id = $1`,
		workflowID,
	).Scan(&spent, &budget, &softPct, &hardPct)
	if err != nil {
		return false, false, fmt.Errorf("check workflow limits: %w", err)
	}
	softHit = spent >= budget*(softPct/100.0)
	hardHit = spent >= budget*(hardPct/100.0)
	return softHit, hardHit, nil
}

// CostReport holds cost breakdown.
type CostReport struct {
	TotalCostUSD     float64            `json:"total_cost_usd"`
	MonthlyLimitUSD  float64            `json:"monthly_limit_usd"`
	UsagePercent     float64            `json:"usage_percent"`
	AlertTriggered   bool               `json:"alert_triggered"`
	TotalLLMCalls    int64              `json:"total_llm_calls"`
	CostByExpert     []ExpertCostItem   `json:"cost_by_expert"`
	CostByDay        []DailyCostItem    `json:"cost_by_day"`
	GeneratedAt      time.Time          `json:"generated_at"`
}

// ExpertCostItem holds cost for a single expert.
type ExpertCostItem struct {
	ExpertName   string  `json:"expert_name"`
	Domain       string  `json:"domain"`
	TotalMessages int    `json:"total_messages"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	AvgCostUSD   float64 `json:"avg_cost_per_message"`
}

// DailyCostItem holds cost for a single day.
type DailyCostItem struct {
	Date         string  `json:"date"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	MessageCount int     `json:"message_count"`
}

// GetReport generates a cost report.
//
// Mental execution:
// 1. Get total cost from model gateway (in-memory)
// 2. Get per-expert cost from messages table
// 3. Get daily cost breakdown
// 4. Check if alert threshold exceeded
// 5. Return report
func (m *CostMonitor) GetReport(ctx context.Context) (*CostReport, error) {
	gwStats := m.gateway.GetStats()
	totalCost, _ := gwStats["total_cost"].(float64)
	totalCalls, _ := gwStats["total_calls"].(int64)

	usagePercent := 0.0
	if m.monthlyLimit > 0 {
		usagePercent = (totalCost / m.monthlyLimit) * 100
	}

	// Per-expert cost from DB
	expertCosts, err := m.getExpertCosts(ctx)
	if err != nil {
		m.logger.Warn("get expert costs failed", zap.Error(err))
		expertCosts = []ExpertCostItem{}
	}

	// Daily cost breakdown
	dailyCosts, err := m.getDailyCosts(ctx)
	if err != nil {
		m.logger.Warn("get daily costs failed", zap.Error(err))
		dailyCosts = []DailyCostItem{}
	}

	alertTriggered := usagePercent >= m.alertThreshold*100
	if alertTriggered {
		m.logger.Warn("COST ALERT: monthly budget threshold exceeded",
			zap.Float64("usage_percent", usagePercent),
			zap.Float64("total_cost", totalCost),
			zap.Float64("monthly_limit", m.monthlyLimit),
		)
	}

	return &CostReport{
		TotalCostUSD:    totalCost,
		MonthlyLimitUSD: m.monthlyLimit,
		UsagePercent:    usagePercent,
		AlertTriggered:  alertTriggered,
		TotalLLMCalls:   totalCalls,
		CostByExpert:    expertCosts,
		CostByDay:       dailyCosts,
		GeneratedAt:     time.Now(),
	}, nil
}

// getExpertCosts returns cost breakdown per expert from messages table.
func (m *CostMonitor) getExpertCosts(ctx context.Context) ([]ExpertCostItem, error) {
	rows, err := m.db.Query(ctx, `
		SELECT
			e.name,
			e.domain,
			COUNT(msg.id) as total_messages,
			COALESCE(SUM(msg.cost_usd), 0) as total_cost,
			COALESCE(AVG(msg.cost_usd), 0) as avg_cost
		FROM experts e
		LEFT JOIN messages msg ON msg.expert_id = e.id
			AND msg.role = 'assistant'
			AND msg.created_at >= date_trunc('month', NOW())
		WHERE e.deleted_at IS NULL
		GROUP BY e.id, e.name, e.domain
		ORDER BY total_cost DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ExpertCostItem
	for rows.Next() {
		var item ExpertCostItem
		if err := rows.Scan(
			&item.ExpertName, &item.Domain,
			&item.TotalMessages, &item.TotalCostUSD, &item.AvgCostUSD,
		); err != nil {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

// getDailyCosts returns cost breakdown per day for the last 30 days.
func (m *CostMonitor) getDailyCosts(ctx context.Context) ([]DailyCostItem, error) {
	rows, err := m.db.Query(ctx, `
		SELECT
			DATE(created_at) as date,
			COALESCE(SUM(cost_usd), 0) as total_cost,
			COUNT(*) as message_count
		FROM messages
		WHERE role = 'assistant'
			AND created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []DailyCostItem
	for rows.Next() {
		var item DailyCostItem
		var date time.Time
		if err := rows.Scan(&date, &item.TotalCostUSD, &item.MessageCount); err != nil {
			continue
		}
		item.Date = date.Format("2006-01-02")
		items = append(items, item)
	}
	return items, nil
}

// CheckBudget checks if monthly budget is exceeded.
// Called periodically or before expensive LLM calls.
func (m *CostMonitor) CheckBudget() (bool, float64) {
	gwStats := m.gateway.GetStats()
	totalCost, _ := gwStats["total_cost"].(float64)
	usagePercent := 0.0
	if m.monthlyLimit > 0 {
		usagePercent = (totalCost / m.monthlyLimit) * 100
	}
	exceeded := usagePercent >= 100
	if exceeded {
		m.logger.Error("BUDGET EXCEEDED: blocking LLM calls",
			zap.Float64("total_cost", totalCost),
			zap.Float64("monthly_limit", m.monthlyLimit),
		)
	}
	return exceeded, usagePercent
}

// FormatCostSummary returns a human-readable cost summary.
func (m *CostMonitor) FormatCostSummary(report *CostReport) string {
	return fmt.Sprintf(
		"Cost: $%.4f / $%.2f (%.1f%%) | Calls: %d | Alert: %v",
		report.TotalCostUSD,
		report.MonthlyLimitUSD,
		report.UsagePercent,
		report.TotalLLMCalls,
		report.AlertTriggered,
	)
}
