// Command eval runs the C3 golden-set evaluation harness against a live
// server, persists the run, and exits non-zero on vital failure / regression.
//
// Usage (from backend-go, server running, migration 029 applied):
//
//	EVAL_BASE_URL=http://localhost:8080/api/v1 \
//	EVAL_TOKEN=<access_token> \
//	EVAL_CHAT_ID=<chat-uuid> \
//	DATABASE_URL=... \
//	go run ./cmd/eval [-suite chat] [-suite adversarial] [-baseline] [-tolerance 0.05]
//
// Exit codes:
//
//	0  — all suites clean (no vital failure, no regression beyond tolerance)
//	1  — configuration / runtime error
//	2  — vital failure or score regression (CI merge gate)
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/eval"
)

type suiteList []string

func (s *suiteList) String() string { return strings.Join(*s, ",") }
func (s *suiteList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func main() {
	var suites suiteList
	var promoteBaseline bool
	var tolerance float64
	var model string
	flag.Var(&suites, "suite", "golden set name (repeatable; default = all embedded)")
	flag.BoolVar(&promoteBaseline, "baseline", false, "promote this run as the suite baseline after a clean pass")
	flag.Float64Var(&tolerance, "tolerance", 0.05, "max allowed score drop vs baseline before regression")
	flag.StringVar(&model, "model", os.Getenv("LLM_MODEL_STRONG"), "model label recorded on the run")
	flag.Parse()

	baseURL := strings.TrimSpace(os.Getenv("EVAL_BASE_URL"))
	token := strings.TrimSpace(os.Getenv("EVAL_TOKEN"))
	chatIDRaw := strings.TrimSpace(os.Getenv("EVAL_CHAT_ID"))
	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if baseURL == "" || token == "" || chatIDRaw == "" || dbURL == "" {
		fmt.Fprintln(os.Stderr, "required env: EVAL_BASE_URL, EVAL_TOKEN, EVAL_CHAT_ID, DATABASE_URL")
		os.Exit(1)
	}
	chatID, err := uuid.Parse(chatIDRaw)
	if err != nil {
		fmt.Fprintln(os.Stderr, "EVAL_CHAT_ID must be a UUID:", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "db connect:", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger, _ := zap.NewProduction()
	defer logger.Sync() //nolint:errcheck

	if len(suites) == 0 {
		names, err := eval.GoldenSetNames()
		if err != nil {
			fmt.Fprintln(os.Stderr, "list golden sets:", err)
			os.Exit(1)
		}
		suites = names
	}
	if model == "" {
		model = "unknown"
	}

	answerer := &eval.HTTPAnswerer{
		BaseURL: baseURL,
		Token:   token,
		ChatID:  chatID,
		Resolve: makeResolver(pool),
		Client:  nil, // default 5m timeout inside Answer
	}
	store := eval.NewStore(pool)
	runner := eval.NewRunner(model, logger)

	failed := false
	for _, name := range suites {
		set, err := eval.LoadGoldenSet(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load suite %s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("== suite %s (%d cases) ==\n", name, len(set.Cases))
		run := runner.Run(ctx, set, answerer)

		baseline, bErr := store.LatestBaseline(ctx, name)
		if bErr != nil {
			fmt.Fprintf(os.Stderr, "baseline lookup %s: %v\n", name, bErr)
			os.Exit(1)
		}
		delta := eval.ScoreDelta(baseline, run)
		regressed := eval.Regression(baseline, run, tolerance)

		if _, err := store.SaveRun(ctx, run, false); err != nil {
			fmt.Fprintf(os.Stderr, "save run %s: %v\n", name, err)
			os.Exit(1)
		}

		fmt.Printf("  total=%d passed=%d vital_failed=%d score=%.3f cost_usd=%.4f delta=%+.3f\n",
			run.Total, run.Passed, run.VitalFailed, run.Score, run.CostUSD, delta)
		for _, r := range run.Results {
			status := "PASS"
			if r.Skipped {
				status = "SKIP"
			} else if r.Error != "" {
				status = "ERR "
			} else if !r.Passed {
				status = "FAIL"
			}
			vital := " "
			if r.Vital {
				vital = "V"
			}
			fmt.Printf("  [%s%s] %s", status, vital, r.CaseID)
			if r.Error != "" {
				fmt.Printf(" err=%s", r.Error)
			}
			fmt.Println()
		}

		if run.HasVitalFailure() || regressed {
			failed = true
			fmt.Printf("  RESULT: BLOCK (vital_failure=%v regression=%v)\n", run.HasVitalFailure(), regressed)
			continue
		}
		fmt.Println("  RESULT: OK")
		if promoteBaseline {
			if err := store.PromoteBaseline(ctx, name, run.ID); err != nil {
				fmt.Fprintf(os.Stderr, "promote baseline %s: %v\n", name, err)
				os.Exit(1)
			}
			fmt.Println("  baseline promoted")
		}
	}

	if failed {
		os.Exit(2)
	}
}

// makeResolver returns a slug→id lookup that skips missing experts
// (ErrSkip) so environment gaps never look like quality regressions.
func makeResolver(pool *pgxpool.Pool) eval.ExpertResolver {
	return func(ctx context.Context, slug string) (uuid.UUID, error) {
		slug = strings.TrimSpace(strings.ToLower(slug))
		if slug == "" {
			return uuid.Nil, eval.ErrSkip
		}
		var id uuid.UUID
		err := pool.QueryRow(ctx,
			`SELECT id FROM experts WHERE LOWER(slug)=$1 AND deleted_at IS NULL LIMIT 1`,
			slug,
		).Scan(&id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return uuid.Nil, eval.ErrSkip
			}
			return uuid.Nil, err
		}
		return id, nil
	}
}
