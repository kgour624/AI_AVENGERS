package workflow

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// minCalibrationSamples: below this, a domain's ratings are noise — keep
// the historical default and surface the shortfall in the proposal so the
// admin knows why nothing moved. 20 is small enough to reach quickly on a
// single-admin app and large enough that one good/bad rating cannot swing
// the bar.
const minCalibrationSamples = 20

// clamp ranges keep proposed values inside the migration CHECK constraints
// and preserve strong > usable by construction (see proposeForDomain).
const (
	calUsableMin  = 0.30
	calUsableMax  = 0.60
	calStrongMin  = 0.55
	calStrongMax  = 0.85
	calDeltaLoose = 0.03 // high-acceptance domains loosen by this
	calDeltaTight = 0.03 // high-rejection domains tighten by this
)

// GateThresholdRow is one row of gate_thresholds (plus derived fields
// the admin UI needs to decide whether to apply).
type GateThresholdRow struct {
	Domain      string    `json:"domain"`
	Usable      float64   `json:"usable"`
	Strong      float64   `json:"strong"`
	Source      string    `json:"source"`
	SampleSize  int       `json:"sample_size"`
	Notes       string    `json:"notes"`
	UpdatedAt   time.Time `json:"updated_at"`
	// Effective is what GateSystem would actually use right now
	// (applied/manual row, else the package default). Populated by List.
	EffectiveUsable float64 `json:"effective_usable"`
	EffectiveStrong float64 `json:"effective_strong"`
	EffectiveSource string  `json:"effective_source"`
}

// domainRatingStats is the aggregate used by Propose.
type domainRatingStats struct {
	Domain      string
	Total       int
	Accepted    int // score>=4 OR feedback_type='accepted'
	Rejected    int // score<=2 OR feedback_type='rejected'
	AvgScore    float64
}

// ListGateThresholds returns every known domain's current state: package
// default as the baseline, any calibrated/applied/manual row, and the
// effective pair GateSystem would use. Domains with ratings but no row
// still appear (so admin sees them as "default").
func ListGateThresholds(ctx context.Context, db *pgxpool.Pool) ([]GateThresholdRow, error) {
	if db == nil {
		return nil, fmt.Errorf("list gate thresholds: nil db")
	}
	def := DefaultGateThresholds()

	// 1. All domains that currently have experts (active) OR a row.
	rows, err := db.Query(ctx, `
		WITH domains AS (
			SELECT DISTINCT lower(domain) AS domain FROM experts
			 WHERE deleted_at IS NULL AND domain <> ''
			UNION
			SELECT lower(domain) FROM gate_thresholds
		),
		rows AS (
			SELECT lower(domain) AS domain, usable, strong, source,
			       sample_size, notes, updated_at
			  FROM gate_thresholds
		)
		SELECT d.domain,
		       r.usable, r.strong, r.source, r.sample_size, r.notes, r.updated_at
		  FROM domains d
		  LEFT JOIN rows r ON r.domain = d.domain
		 ORDER BY d.domain`)
	if err != nil {
		return nil, fmt.Errorf("list gate thresholds: %w", err)
	}
	defer rows.Close()

	out := make([]GateThresholdRow, 0)
	for rows.Next() {
		var (
			domain                           string
			usable, strong                   *float64
			source                           *string
			sampleSize                       *int
			notes                            *string
			updatedAt                        *time.Time
		)
		if err := rows.Scan(&domain, &usable, &strong, &source, &sampleSize, &notes, &updatedAt); err != nil {
			continue
		}
		row := GateThresholdRow{
			Domain:          domain,
			EffectiveUsable: def.Usable,
			EffectiveStrong: def.Strong,
			EffectiveSource: "default",
		}
		if usable != nil && strong != nil && source != nil {
			row.Usable = *usable
			row.Strong = *strong
			row.Source = *source
			if sampleSize != nil {
				row.SampleSize = *sampleSize
			}
			if notes != nil {
				row.Notes = *notes
			}
			if updatedAt != nil {
				row.UpdatedAt = *updatedAt
			}
			if *source == "applied" || *source == "manual" {
				row.EffectiveUsable = *usable
				row.EffectiveStrong = *strong
				row.EffectiveSource = *source
			}
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// ProposeGateThresholds aggregates ratings per expert domain and writes a
// calibrated proposal row for each domain with enough samples. Existing
// applied/manual rows are left alone (proposal sits next to them via the
// same PK — we UPSERT only when source would stay 'calibrated' OR when
// no applied/manual row exists). Admin Apply is required to make the
// proposal live (P7: no blind auto-tune).
//
// Heuristic (simple, documented, reversible):
//   acceptance = accepted / total
//   rejection  = rejected / total
//   high acceptance (>=0.80) + low rejection (<0.15) → loosen usable/strong
//   high rejection  (>=0.30)                        → tighten usable/strong
//   otherwise                                       → keep defaults
func ProposeGateThresholds(ctx context.Context, db *pgxpool.Pool, logger *zap.Logger) ([]GateThresholdRow, error) {
	if db == nil {
		return nil, fmt.Errorf("propose gate thresholds: nil db")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	stats, err := loadDomainRatingStats(ctx, db)
	if err != nil {
		return nil, err
	}
	def := DefaultGateThresholds()
	proposed := make([]GateThresholdRow, 0, len(stats))
	for _, s := range stats {
		row, ok := proposeForDomain(s, def)
		if !ok {
			logger.Info("gate calibration: skip domain (insufficient sample)",
				zap.String("domain", s.Domain),
				zap.Int("sample", s.Total),
				zap.Int("min", minCalibrationSamples),
			)
			continue
		}
		// Never overwrite an applied/manual row — admin intent wins until
		// they re-apply. A calibrated proposal only lands when the domain
		// is fresh or already holding a previous proposal.
		if src, exists, _ := existingSource(ctx, db, s.Domain); exists && (src == "applied" || src == "manual") {
			logger.Info("gate calibration: skip domain (applied/manual holds)",
				zap.String("domain", s.Domain),
				zap.String("source", src),
			)
			continue
		}
		if err := upsertCalibrated(ctx, db, row); err != nil {
			logger.Warn("gate calibration: upsert failed",
				zap.String("domain", s.Domain),
				zap.Error(err),
			)
			continue
		}
		proposed = append(proposed, row)
		logger.Info("gate calibration: proposed",
			zap.String("domain", s.Domain),
			zap.Float64("usable", row.Usable),
			zap.Float64("strong", row.Strong),
			zap.Int("sample", row.SampleSize),
		)
	}
	return proposed, nil
}

func existingSource(ctx context.Context, db *pgxpool.Pool, domain string) (string, bool, error) {
	var src string
	err := db.QueryRow(ctx,
		`SELECT source FROM gate_thresholds WHERE lower(domain) = lower($1)`,
		domain,
	).Scan(&src)
	if err != nil {
		return "", false, err
	}
	return src, true, nil
}

// ApplyGateThreshold promotes a calibrated (or freshly written) row for
// domain to source='applied' so GateSystem starts using it. Returns the
// applied row. missing row → error.
func ApplyGateThreshold(ctx context.Context, db *pgxpool.Pool, domain string, adminID uuid.UUID) (GateThresholdRow, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return GateThresholdRow{}, fmt.Errorf("apply gate threshold: empty domain")
	}
	if db == nil {
		return GateThresholdRow{}, fmt.Errorf("apply gate threshold: nil db")
	}
	var row GateThresholdRow
	err := db.QueryRow(ctx, `
		UPDATE gate_thresholds
		   SET source = 'applied',
		       updated_at = NOW(),
		       updated_by = $2,
		       notes = CASE
		                 WHEN notes = '' THEN 'applied by admin'
		                 ELSE notes || ' | applied by admin'
		               END
		 WHERE lower(domain) = $1
		 RETURNING domain, usable, strong, source, sample_size, notes, updated_at`,
		domain, adminID,
	).Scan(&row.Domain, &row.Usable, &row.Strong, &row.Source, &row.SampleSize, &row.Notes, &row.UpdatedAt)
	if err != nil {
		return GateThresholdRow{}, fmt.Errorf("apply gate threshold %q: %w", domain, err)
	}
	row.EffectiveUsable = row.Usable
	row.EffectiveStrong = row.Strong
	row.EffectiveSource = row.Source
	return row, nil
}

// SetGateThreshold writes a manual override (admin hand-edit). Marks
// source='manual' so it is immediately live for GateSystem.
func SetGateThreshold(ctx context.Context, db *pgxpool.Pool, domain string, usable, strong float64, adminID uuid.UUID, notes string) (GateThresholdRow, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return GateThresholdRow{}, fmt.Errorf("set gate threshold: empty domain")
	}
	if strong <= usable {
		return GateThresholdRow{}, fmt.Errorf("set gate threshold: strong (%.3f) must be > usable (%.3f)", strong, usable)
	}
	if usable < calUsableMin || usable > calUsableMax || strong < calStrongMin || strong > calStrongMax {
		return GateThresholdRow{}, fmt.Errorf(
			"set gate threshold: usable in [%.2f,%.2f], strong in [%.2f,%.2f]",
			calUsableMin, calUsableMax, calStrongMin, calStrongMax,
		)
	}
	if notes == "" {
		notes = "manual admin override"
	}
	var row GateThresholdRow
	err := db.QueryRow(ctx, `
		INSERT INTO gate_thresholds (domain, usable, strong, source, sample_size, notes, updated_by)
		VALUES ($1, $2, $3, 'manual', 0, $4, $5)
		ON CONFLICT (domain) DO UPDATE SET
			usable     = EXCLUDED.usable,
			strong     = EXCLUDED.strong,
			source     = 'manual',
			notes      = EXCLUDED.notes,
			updated_by = EXCLUDED.updated_by,
			updated_at = NOW()
		RETURNING domain, usable, strong, source, sample_size, notes, updated_at`,
		domain, usable, strong, notes, adminID,
	).Scan(&row.Domain, &row.Usable, &row.Strong, &row.Source, &row.SampleSize, &row.Notes, &row.UpdatedAt)
	if err != nil {
		return GateThresholdRow{}, fmt.Errorf("set gate threshold %q: %w", domain, err)
	}
	row.EffectiveUsable = row.Usable
	row.EffectiveStrong = row.Strong
	row.EffectiveSource = row.Source
	return row, nil
}

// loadDomainRatingStats aggregates ratings joined to experts.domain.
func loadDomainRatingStats(ctx context.Context, db *pgxpool.Pool) ([]domainRatingStats, error) {
	rows, err := db.Query(ctx, `
		SELECT lower(e.domain) AS domain,
		       COUNT(r.id) AS total,
		       COUNT(*) FILTER (
		           WHERE r.score >= 4 OR r.feedback_type = 'accepted'
		       ) AS accepted,
		       COUNT(*) FILTER (
		           WHERE r.score <= 2 OR r.feedback_type = 'rejected'
		       ) AS rejected,
		       COALESCE(AVG(r.score), 0) AS avg_score
		  FROM experts e
		  JOIN ratings r ON r.expert_id = e.id
		 WHERE e.deleted_at IS NULL
		   AND e.domain <> ''
		 GROUP BY lower(e.domain)
		 ORDER BY lower(e.domain)`)
	if err != nil {
		return nil, fmt.Errorf("load domain rating stats: %w", err)
	}
	defer rows.Close()

	out := make([]domainRatingStats, 0)
	for rows.Next() {
		var s domainRatingStats
		if err := rows.Scan(&s.Domain, &s.Total, &s.Accepted, &s.Rejected, &s.AvgScore); err != nil {
			continue
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// proposeForDomain derives a calibrated pair from stats. ok=false when
// the sample is too small — caller must not write a row in that case.
func proposeForDomain(s domainRatingStats, def GateThresholds) (GateThresholdRow, bool) {
	if s.Total < minCalibrationSamples {
		return GateThresholdRow{}, false
	}
	acceptance := float64(s.Accepted) / float64(s.Total)
	rejection := float64(s.Rejected) / float64(s.Total)

	usable := def.Usable
	strong := def.Strong
	action := "hold"
	switch {
	case acceptance >= 0.80 && rejection < 0.15:
		// Over-refusing signal: clients like the answers → loosen.
		usable = clamp(usable-calDeltaLoose, calUsableMin, calUsableMax)
		strong = clamp(strong-calDeltaLoose, calStrongMin, calStrongMax)
		action = "loosen"
	case rejection >= 0.30:
		// Over-answering signal: clients reject a lot → tighten.
		usable = clamp(usable+calDeltaTight, calUsableMin, calUsableMax)
		strong = clamp(strong+calDeltaTight, calStrongMin, calStrongMax)
		action = "tighten"
	}
	// Keep the order invariant even after independent clamps.
	if strong <= usable {
		strong = math.Min(calStrongMax, usable+0.15)
	}

	notes := fmt.Sprintf(
		"calibrated action=%s n=%d accepted=%d rejected=%d avg=%.2f acceptance=%.2f rejection=%.2f base_usable=%.2f base_strong=%.2f",
		action, s.Total, s.Accepted, s.Rejected, s.AvgScore, acceptance, rejection, def.Usable, def.Strong,
	)
	return GateThresholdRow{
		Domain:     s.Domain,
		Usable:     round3(usable),
		Strong:     round3(strong),
		Source:     "calibrated",
		SampleSize: s.Total,
		Notes:      notes,
		UpdatedAt:  time.Now().UTC(),
	}, true
}

// upsertCalibrated writes a calibrated proposal. It refuses to overwrite
// an applied/manual row (admin intent wins until they re-apply). A fresh
// domain or a previous calibrated proposal is free to replace.
func upsertCalibrated(ctx context.Context, db *pgxpool.Pool, row GateThresholdRow) error {
	_, err := db.Exec(ctx, `
		INSERT INTO gate_thresholds (domain, usable, strong, source, sample_size, notes)
		VALUES ($1, $2, $3, 'calibrated', $4, $5)
		ON CONFLICT (domain) DO UPDATE SET
			usable      = EXCLUDED.usable,
			strong      = EXCLUDED.strong,
			source      = 'calibrated',
			sample_size = EXCLUDED.sample_size,
			notes       = EXCLUDED.notes,
			updated_at  = NOW(),
			updated_by  = NULL
		WHERE gate_thresholds.source = 'calibrated'`,
		row.Domain, row.Usable, row.Strong, row.SampleSize, row.Notes,
	)
	return err
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}
