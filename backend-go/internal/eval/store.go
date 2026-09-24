package eval

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists evaluation runs and baselines.
type Store struct {
	db *pgxpool.Pool
}

// NewStore builds a run store.
func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// SaveRun inserts a run and returns its id. isBaseline promotes it (the
// caller must clear the previous baseline first — see PromoteBaseline).
func (s *Store) SaveRun(ctx context.Context, run *RunSummary, isBaseline bool) (uuid.UUID, error) {
	results, err := json.Marshal(run.Results)
	if err != nil {
		return uuid.Nil, err
	}
	var id uuid.UUID
	err = s.db.QueryRow(ctx,
		`INSERT INTO eval_runs
			(suite, model, total, passed, vital_total, vital_failed, score, cost_usd, is_baseline, results, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		 RETURNING id`,
		run.Suite, nullStr(run.Model), run.Total, run.Passed, run.VitalTotal, run.VitalFailed,
		run.Score, run.CostUSD, isBaseline, results, run.CreatedAt,
	).Scan(&id)
	if err == nil {
		run.ID = id
	}
	return id, err
}

// PromoteBaseline clears any existing baseline for the suite, then marks
// the given run as the baseline (single baseline per suite).
func (s *Store) PromoteBaseline(ctx context.Context, suite string, runID uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx,
		`UPDATE eval_runs SET is_baseline=FALSE WHERE suite=$1 AND is_baseline`, suite,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE eval_runs SET is_baseline=TRUE WHERE id=$1 AND suite=$2`, runID, suite,
	); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// LatestBaseline returns the baseline run for a suite, or (nil, nil).
func (s *Store) LatestBaseline(ctx context.Context, suite string) (*RunSummary, error) {
	return s.queryOne(ctx,
		`SELECT id, suite, model, total, passed, vital_total, vital_failed, score, cost_usd, results, created_at
		 FROM eval_runs WHERE suite=$1 AND is_baseline LIMIT 1`,
		suite,
	)
}

// LatestRun returns the most recent run for a suite, or (nil, nil).
func (s *Store) LatestRun(ctx context.Context, suite string) (*RunSummary, error) {
	return s.queryOne(ctx,
		`SELECT id, suite, model, total, passed, vital_total, vital_failed, score, cost_usd, results, created_at
		 FROM eval_runs WHERE suite=$1 ORDER BY created_at DESC LIMIT 1`,
		suite,
	)
}

// ListSuites returns the distinct suites that have runs, newest first.
func (s *Store) ListSuites(ctx context.Context) ([]string, error) {
	rows, err := s.db.Query(ctx,
		`SELECT suite FROM eval_runs GROUP BY suite ORDER BY MAX(created_at) DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var suite string
		if err := rows.Scan(&suite); err != nil {
			continue
		}
		out = append(out, suite)
	}
	if out == nil {
		out = []string{}
	}
	return out, rows.Err()
}

func (s *Store) queryOne(ctx context.Context, sql string, arg interface{}) (*RunSummary, error) {
	var run RunSummary
	var model *string
	var results []byte
	err := s.db.QueryRow(ctx, sql, arg).Scan(
		&run.ID, &run.Suite, &model, &run.Total, &run.Passed,
		&run.VitalTotal, &run.VitalFailed, &run.Score, &run.CostUSD, &results, &run.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if model != nil {
		run.Model = *model
	}
	_ = json.Unmarshal(results, &run.Results)
	if run.Results == nil {
		run.Results = []CaseResult{}
	}
	return &run, nil
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
