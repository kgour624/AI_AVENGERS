// Package usage implements C5: the cost & usage analytics product.
//
// WHY a dedicated package:
//   Cost is computed in exactly one place (the ModelGateway, G5). This
//   package turns those per-call costs into a persisted, queryable surface:
//   spend grouped by tenant / project / expert / model / use-case, plus
//   monthly budgets and alert thresholds. Keeping it separate from
//   monitoring.CostMonitor (which reports an in-memory running total) means
//   the gateway only depends on a tiny Recorder interface, and the analytics
//   /budget SQL lives in one place.
//
// ATTRIBUTION: a call does not always know its project/tenant. Callers attach
// what they know to the context (WithAttribution); the gateway reads it back
// (AttributionFrom) and records it. Anything unknown is resolved best-effort
// from project_id / workflow_id at write time so per-tenant grouping is still
// correct for the chat path.
//
// SOLID: SRP (recording + aggregation + budgets). DIP (gateway depends on the
// Recorder interface, not on Postgres). Pattern: Null Object (nil *Service is
// a no-op recorder).
package usage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Use-case labels (kept short + stable — they are DB values and API keys).
const (
	UseCaseChat     = "chat"
	UseCaseSynthesis = "synthesis"
	UseCaseConsolidate = "memory_consolidate"
	UseCaseSummary  = "summary"
	UseCaseIndex    = "index"
	UseCaseWorkflow = "workflow"
)

// Attribution is the optional context of one LLM call. The zero value is
// valid — an unattributed call is still recorded (grouped by model).
type Attribution struct {
	TenantID   *uuid.UUID
	ProjectID  *uuid.UUID
	ExpertID   *uuid.UUID
	ChatID     *uuid.UUID
	WorkflowID *uuid.UUID
	AccountID  *uuid.UUID
	UseCase    string
}

// IsZero reports whether nothing is attributed.
func (a Attribution) IsZero() bool {
	return a.TenantID == nil && a.ProjectID == nil && a.ExpertID == nil &&
		a.ChatID == nil && a.WorkflowID == nil && a.AccountID == nil && a.UseCase == ""
}

// Event is one recorded LLM call.
type Event struct {
	Attribution
	Provider     string
	Tier         string
	Model        string
	InputTokens  int
	OutputTokens int
	CostUSD      float64
	OccurredAt   time.Time
}

// Recorder persists usage events. Defined by the consumer (the ModelGateway).
type Recorder interface {
	Record(ctx context.Context, e Event) error
}

// ---- context-carried attribution ----

type ctxKey struct{}

// WithAttribution returns a child context carrying a for downstream gateway
// calls. Later calls override earlier ones (innermost wins).
func WithAttribution(ctx context.Context, a Attribution) context.Context {
	return context.WithValue(ctx, ctxKey{}, a)
}

// AttributionFrom returns the attribution attached to ctx, if any.
func AttributionFrom(ctx context.Context) (Attribution, bool) {
	a, ok := ctx.Value(ctxKey{}).(Attribution)
	return a, ok
}

// ---- service ----

// Service records usage and answers analytics/budget queries.
type Service struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewService builds the usage service.
func NewService(db *pgxpool.Pool, logger *zap.Logger) *Service {
	return &Service{db: db, logger: logger}
}

// Enabled reports whether the service can operate. Nil-safe.
func (s *Service) Enabled() bool { return s != nil && s.db != nil }

// Record appends one usage event. Best-effort: a monitoring write must never
// fail an LLM response the caller already paid for. Tenant is resolved from
// project_id (or workflow_id) when the caller did not supply it, so the chat
// path groups correctly per tenant without an extra call-site lookup.
func (s *Service) Record(ctx context.Context, e Event) error {
	if !s.Enabled() || e.CostUSD <= 0 {
		return nil
	}
	// Detached: the spend already happened even if ctx is cancelled.
	ctx = context.WithoutCancel(ctx)

	projectID := e.ProjectID
	if projectID == nil && e.WorkflowID != nil {
		projectID = s.projectForWorkflow(ctx, *e.WorkflowID)
	}
	tenantID := e.TenantID
	if tenantID == nil && projectID != nil {
		tenantID = s.tenantForProject(ctx, *projectID)
	}

	occurred := e.OccurredAt
	if occurred.IsZero() {
		occurred = time.Now().UTC()
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO usage_events
			(occurred_at, tenant_id, project_id, expert_id, chat_id, workflow_id,
			 account_id, provider, tier, model, use_case, input_tokens, output_tokens, cost_usd)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		occurred, tenantID, projectID, e.ExpertID, e.ChatID, e.WorkflowID,
		e.AccountID, nullStr(e.Provider), nullStr(e.Tier), nullStr(e.Model),
		nullStr(e.UseCase), e.InputTokens, e.OutputTokens, e.CostUSD,
	)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("usage record failed", zap.Error(err))
		}
		return fmt.Errorf("usage: record: %w", err)
	}
	return nil
}

// ---- analytics ----

// Filter selects and groups usage rows.
type Filter struct {
	From      time.Time
	To        time.Time
	GroupBy   string // tenant | project | expert | model | use_case
	TenantID  *uuid.UUID
	ProjectID *uuid.UUID
	ExpertID  *uuid.UUID
	UseCase   string
}

// Row is one aggregate bucket.
type Row struct {
	Key          string  `json:"key"`
	Label        string  `json:"label"`
	Calls        int64   `json:"calls"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	CostUSD      float64 `json:"cost_usd"`
}

// Summary returns usage rows grouped by Filter.GroupBy (default tenant).
func (s *Service) Summary(ctx context.Context, f Filter) ([]Row, error) {
	if !s.Enabled() {
		return []Row{}, nil
	}
	if f.From.IsZero() {
		f.From = MonthStart(time.Now())
	}
	if f.To.IsZero() {
		f.To = time.Now().UTC().Add(time.Minute)
	}
	keyExpr, labelExpr := groupExpr(f.GroupBy)

	q := `SELECT ` + keyExpr + ` AS key, ` + labelExpr + ` AS label,
		       COUNT(*),
		       COALESCE(SUM(u.input_tokens),0),
		       COALESCE(SUM(u.output_tokens),0),
		       COALESCE(SUM(u.cost_usd),0)
		FROM usage_events u
		LEFT JOIN tenants  t ON t.id = u.tenant_id
		LEFT JOIN projects p ON p.id = u.project_id
		LEFT JOIN experts  e ON e.id = u.expert_id
		WHERE u.occurred_at >= $1 AND u.occurred_at < $2
		  AND ($3::uuid IS NULL OR u.tenant_id  = $3)
		  AND ($4::uuid IS NULL OR u.project_id = $4)
		  AND ($5::uuid IS NULL OR u.expert_id  = $5)
		  AND ($6 = ''    OR u.use_case   = $6)
		GROUP BY 1, 2
		ORDER BY 6 DESC
		LIMIT 200`

	rows, err := s.db.Query(ctx, q, f.From, f.To, f.TenantID, f.ProjectID, f.ExpertID, f.UseCase)
	if err != nil {
		return nil, fmt.Errorf("usage summary: %w", err)
	}
	defer rows.Close()

	out := []Row{}
	for rows.Next() {
		var r Row
		if err := rows.Scan(&r.Key, &r.Label, &r.Calls, &r.InputTokens, &r.OutputTokens, &r.CostUSD); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// groupExpr maps a group-by key to (keyExpr, labelExpr) SQL. Whitelisted so
// the value can never be injected; unknown keys fall back to tenant.
func groupExpr(groupBy string) (keyExpr, labelExpr string) {
	switch strings.ToLower(strings.TrimSpace(groupBy)) {
	case "project":
		return "COALESCE(u.project_id::text,'(none)')", "COALESCE(p.name,'(none)')"
	case "expert":
		return "COALESCE(u.expert_id::text,'(none)')", "COALESCE(e.name,'(none)')"
	case "model":
		return "COALESCE(NULLIF(u.model,''),'(unknown)')", "COALESCE(NULLIF(u.model,''),'(unknown)')"
	case "use_case":
		return "COALESCE(NULLIF(u.use_case,''),'(none)')", "COALESCE(NULLIF(u.use_case,''),'(none)')"
	default: // tenant
		return "COALESCE(u.tenant_id::text,'(global)')", "COALESCE(t.name,'(global)')"
	}
}

// MonthSpend returns this calendar month's spend for a tenant (nil = global,
// i.e. platform-wide).
func (s *Service) MonthSpend(ctx context.Context, tenantID *uuid.UUID) (float64, error) {
	if !s.Enabled() {
		return 0, nil
	}
	var spend float64
	err := s.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(cost_usd),0) FROM usage_events
		 WHERE occurred_at >= date_trunc('month', NOW())
		   AND ($1::uuid IS NULL OR tenant_id = $1)`,
		tenantID,
	).Scan(&spend)
	if err != nil {
		return 0, fmt.Errorf("usage month spend: %w", err)
	}
	return spend, nil
}

// ---- budgets ----

// Budget is a monthly spend limit + alert threshold.
type Budget struct {
	TenantID       *uuid.UUID `json:"tenant_id,omitempty"`
	MonthlyLimitUSD float64   `json:"monthly_limit_usd"`
	AlertThreshold  float64   `json:"alert_threshold"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// BudgetStatus is a budget resolved against actual spend.
type BudgetStatus struct {
	TenantID       *uuid.UUID `json:"tenant_id,omitempty"`
	SpendUSD       float64    `json:"spend_usd"`
	LimitUSD       float64    `json:"limit_usd"`
	Percent        float64    `json:"percent"`
	AlertThreshold float64    `json:"alert_threshold"`
	Alert          bool       `json:"alert"`
	Breached       bool       `json:"breached"`
}

// GetBudget returns the tenant's budget, falling back to the global row.
// Returns (nil, nil) when neither exists.
func (s *Service) GetBudget(ctx context.Context, tenantID *uuid.UUID) (*Budget, error) {
	if !s.Enabled() {
		return nil, nil
	}
	if tenantID != nil {
		if b, err := s.scanBudget(ctx,
			`SELECT tenant_id, monthly_limit_usd, alert_threshold, updated_at
			 FROM usage_budgets WHERE tenant_id = $1`, *tenantID); err != nil {
			return nil, err
		} else if b != nil {
			return b, nil
		}
	}
	return s.scanBudget(ctx,
		`SELECT tenant_id, monthly_limit_usd, alert_threshold, updated_at
		 FROM usage_budgets WHERE tenant_id IS NULL`, nil)
}

func (s *Service) scanBudget(ctx context.Context, sql string, arg interface{}) (*Budget, error) {
	var b Budget
	var args []interface{}
	if arg != nil {
		args = append(args, arg)
	}
	err := s.db.QueryRow(ctx, sql, args...).Scan(&b.TenantID, &b.MonthlyLimitUSD, &b.AlertThreshold, &b.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

// SetBudget upserts a budget. tenantID nil = the platform default row.
func (s *Service) SetBudget(ctx context.Context, tenantID *uuid.UUID, limit, threshold float64) error {
	if !s.Enabled() {
		return fmt.Errorf("usage service not enabled")
	}
	if limit < 0 {
		return fmt.Errorf("monthly_limit_usd must be >= 0")
	}
	if threshold <= 0 || threshold > 1 {
		threshold = 0.8
	}
	_, err := s.db.Exec(ctx,
		`INSERT INTO usage_budgets (tenant_id, monthly_limit_usd, alert_threshold, updated_at)
		 VALUES ($1,$2,$3,NOW())
		 ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)))
		 DO UPDATE SET monthly_limit_usd = EXCLUDED.monthly_limit_usd,
		               alert_threshold   = EXCLUDED.alert_threshold,
		               updated_at        = NOW()`,
		tenantID, limit, threshold,
	)
	if err != nil {
		return fmt.Errorf("set usage budget: %w", err)
	}
	return nil
}

// BudgetStatus returns the resolved status for a tenant (nil = global).
func (s *Service) BudgetStatus(ctx context.Context, tenantID *uuid.UUID) (*BudgetStatus, error) {
	if !s.Enabled() {
		return nil, nil
	}
	b, err := s.GetBudget(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	spend, err := s.MonthSpend(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return statusFor(tenantID, spend, b.MonthlyLimitUSD, b.AlertThreshold), nil
}

// Alerts returns a status for every configured budget whose spend has reached
// its alert threshold (or breach). Empty when nothing is configured.
func (s *Service) Alerts(ctx context.Context) ([]BudgetStatus, error) {
	if !s.Enabled() {
		return []BudgetStatus{}, nil
	}
	rows, err := s.db.Query(ctx,
		`SELECT tenant_id, monthly_limit_usd, alert_threshold FROM usage_budgets`)
	if err != nil {
		return nil, fmt.Errorf("usage alerts: %w", err)
	}
	defer rows.Close()

	var budgets []Budget
	for rows.Next() {
		var b Budget
		if err := rows.Scan(&b.TenantID, &b.MonthlyLimitUSD, &b.AlertThreshold); err != nil {
			return nil, err
		}
		budgets = append(budgets, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := []BudgetStatus{}
	for _, b := range budgets {
		spend, err := s.MonthSpend(ctx, b.TenantID)
		if err != nil {
			return nil, err
		}
		st := statusFor(b.TenantID, spend, b.MonthlyLimitUSD, b.AlertThreshold)
		if st.Alert || st.Breached {
			out = append(out, *st)
		}
	}
	return out, nil
}

func statusFor(tenantID *uuid.UUID, spend, limit, threshold float64) *BudgetStatus {
	breached, alert := BudgetBreached(spend, limit, threshold)
	return &BudgetStatus{
		TenantID:       tenantID,
		SpendUSD:       spend,
		LimitUSD:       limit,
		Percent:        UsagePercent(spend, limit),
		AlertThreshold: threshold,
		Alert:          alert,
		Breached:       breached,
	}
}

// ---- pure helpers (unit-tested) ----

// MonthStart returns the first instant of now's calendar month, UTC.
func MonthStart(now time.Time) time.Time {
	u := now.UTC()
	return time.Date(u.Year(), u.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// UsagePercent returns spend/limit*100, or 0 when limit is non-positive.
func UsagePercent(spend, limit float64) float64 {
	if limit <= 0 {
		return 0
	}
	return spend / limit * 100
}

// BudgetBreached returns (breached, alert) for a spend against a limit.
// A non-positive limit disables both. Alert fires at/above threshold, breach
// at/above 100%.
func BudgetBreached(spend, limit, threshold float64) (breached, alert bool) {
	if limit <= 0 {
		return false, false
	}
	if threshold <= 0 || threshold > 1 {
		threshold = 0.8
	}
	breached = spend >= limit
	alert = breached || spend >= limit*threshold
	return breached, alert
}

func (s *Service) projectForWorkflow(ctx context.Context, workflowID uuid.UUID) *uuid.UUID {
	var pid uuid.UUID
	if err := s.db.QueryRow(ctx,
		`SELECT project_id FROM workflows WHERE id=$1`, workflowID,
	).Scan(&pid); err != nil || pid == uuid.Nil {
		return nil
	}
	return &pid
}

func (s *Service) tenantForProject(ctx context.Context, projectID uuid.UUID) *uuid.UUID {
	var tid *uuid.UUID
	if err := s.db.QueryRow(ctx,
		`SELECT tenant_id FROM projects WHERE id=$1`, projectID,
	).Scan(&tid); err != nil {
		return nil
	}
	return tid
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
