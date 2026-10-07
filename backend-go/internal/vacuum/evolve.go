package vacuum

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/observability"
)

// AutoPromote — Phase 3 auto-evolution: confidence>=0.90 + reason non-empty + hit_count>=3 + status pending
// I: candidate_kachra rows -> P: UPSERT into kachra_patterns + approved -> trigger bump_brain_version -> O: promoted count
// Human-in-loop kam, but safe: only high-confidence, multi-hit patterns.
func (s *Service) AutoPromote(ctx context.Context) (int64, error) {
	if s.db == nil {
		return 0, nil
	}
	// heal: use short tx with SKIP LOCKED to avoid concurrent double-promote
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `
		SELECT id::text, pattern, pattern_type, category FROM candidate_kachra
		WHERE status='pending' AND hit_count>=3 AND confidence>=0.90 AND length(trim(pattern))>0
		ORDER BY hit_count DESC LIMIT 100 FOR UPDATE SKIP LOCKED`)
	if err != nil {
		return 0, err
	}
	type row struct{ id, pat, ptype, cat string }
	var candidates []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.pat, &r.ptype, &r.cat); err != nil {
			continue
		}
		if strings.TrimSpace(r.pat) == "" {
			continue
		}
		candidates = append(candidates, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(candidates) == 0 {
		_ = tx.Commit(ctx)
		return 0, nil
	}
	promoted := int64(0)
	for _, c := range candidates {
		_, _ = tx.Exec(ctx, `INSERT INTO kachra_patterns(pattern, pattern_type, category) VALUES ($1,$2,$3) ON CONFLICT (lower(trim(pattern)), pattern_type, category) DO UPDATE SET hit_count=kachra_patterns.hit_count+1, updated_at=NOW()`, c.pat, c.ptype, c.cat)
		tag, _ := tx.Exec(ctx, `UPDATE candidate_kachra SET status='approved', updated_at=NOW() WHERE id=$1::uuid AND status='pending'`, c.id)
		if tag.RowsAffected() > 0 {
			promoted++
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	if promoted > 0 {
		observability.Global.IncAutoPromoted(promoted)
		go func() {
			bg, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.brain.Reload(bg); err != nil && s.logger != nil {
				s.logger.Warn("auto-promote brain reload failed", zap.Error(err))
			}
		}()
	}
	return promoted, nil
}

// DriftCheck — eval/drift: compare recent vacuum_eval avg vs baseline.
// Returns driftDetected bool (avg eval_score drop > 15% over last 24h window).
func (s *Service) DriftCheck(ctx context.Context) (bool, float64, float64, error) {
	if s.db == nil {
		return false, 0, 0, nil
	}
	var recentAvg, olderAvg float64
	_ = s.db.QueryRow(ctx, `SELECT COALESCE(AVG(eval_score),0) FROM file_jobs WHERE status='done' AND updated_at > NOW() - interval '24 hours' AND eval_score IS NOT NULL`).Scan(&recentAvg)
	_ = s.db.QueryRow(ctx, `SELECT COALESCE(AVG(eval_score),0) FROM file_jobs WHERE status='done' AND updated_at BETWEEN NOW()-interval '7 days' AND NOW()-interval '24 hours' AND eval_score IS NOT NULL`).Scan(&olderAvg)
	if olderAvg == 0 {
		return false, recentAvg, olderAvg, nil
	}
	drop := (olderAvg - recentAvg) / olderAvg
	drift := drop > 0.15
	if drift {
		observability.Global.IncDriftFail()
	}
	return drift, recentAvg, olderAvg, nil
}
