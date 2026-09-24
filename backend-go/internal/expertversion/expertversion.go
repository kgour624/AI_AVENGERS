// Package expertversion implements C2: expert charter/corpus versioning
// and capability drift detection.
//
// §3.1 P8: version the corpus/charter (and the model that embedded it) so
// a re-ingest that silently changes behaviour becomes visible, and a
// corpus/charter change can be flagged for re-eval (C3) before promotion.
//
// A Version is an immutable snapshot: hashes (corpus + charter), the
// declared capability snapshot, and the charter text — enough to pin a
// canonical version and roll the charter back without re-ingesting. Full
// corpus rollback needs the original transcripts and is deferred.
//
// All writes are best-effort at the call sites (an ingest must not fail
// because versioning did) — the same fail-open policy as C1/B6/B8.
package expertversion

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Version is one immutable expert snapshot.
type Version struct {
	ID                   uuid.UUID       `json:"id"`
	ExpertID             uuid.UUID       `json:"expert_id"`
	VersionNumber        int             `json:"version_number"`
	Source               string          `json:"source"`
	CorpusHash           string          `json:"corpus_hash"`
	CharterHash          string          `json:"charter_hash"`
	ChunkCount           int             `json:"chunk_count"`
	TopicCount           int             `json:"topic_count"`
	Topics               []string        `json:"topics"`
	ReasoningCharter     string          `json:"reasoning_charter"`
	ClarificationCharter json.RawMessage `json:"clarification_charter"`
	CapabilitySummary    json.RawMessage `json:"capability_summary"`
	Model                string          `json:"model,omitempty"`
	IsActive             bool            `json:"is_active"`
	Notes                string          `json:"notes"`
	CreatedAt            time.Time       `json:"created_at"`
}

// DriftEvent records a detected change between two versions.
type DriftEvent struct {
	ID            uuid.UUID       `json:"id"`
	ExpertID      uuid.UUID       `json:"expert_id"`
	FromVersionID *uuid.UUID      `json:"from_version_id,omitempty"`
	ToVersionID   *uuid.UUID      `json:"to_version_id,omitempty"`
	DriftType     string          `json:"drift_type"`
	DriftScore    float64         `json:"drift_score"`
	Details       json.RawMessage `json:"details"`
	Acknowledged  bool            `json:"acknowledged"`
	CreatedAt     time.Time       `json:"created_at"`
}

// liveState is the current on-disk state of an expert's corpus.
type liveState struct {
	CorpusHash string
	Topics     []string
	ChunkCount int
}

// Service snapshots and compares expert versions.
type Service struct {
	db        *pgxpool.Pool
	threshold float64
	logger    *zap.Logger
}

// NewService builds the versioning service. threshold is the topic-set
// (capability) Jaccard distance at/above which drift is reported.
func NewService(db *pgxpool.Pool, threshold float64, logger *zap.Logger) *Service {
	if threshold <= 0 || threshold > 1 {
		threshold = 0.30
	}
	return &Service{db: db, threshold: threshold, logger: logger}
}

// Enabled reports whether the service can operate.
func (s *Service) Enabled() bool {
	return s != nil && s.db != nil
}

// Snapshot captures a new immutable version of the expert's current corpus
// + charter + declared capability, then detects drift against the prior
// active version (best-effort). The first version becomes active.
func (s *Service) Snapshot(ctx context.Context, expertID uuid.UUID, source, notes string) (*Version, error) {
	if !s.Enabled() || expertID == uuid.Nil {
		return nil, nil
	}
	if source == "" {
		source = "ingest"
	}

	live, err := s.liveState(ctx, expertID)
	if err != nil {
		return nil, err
	}

	var reasoning string
	var clarification, capability []byte
	var model string
	err = s.db.QueryRow(ctx,
		`SELECT COALESCE(reasoning_charter,''),
		        COALESCE(clarification_charter,'{}'),
		        COALESCE(capability_summary,'{}'),
		        COALESCE(model_tier,'')
		 FROM experts WHERE id=$1 AND deleted_at IS NULL`,
		expertID,
	).Scan(&reasoning, &clarification, &capability, &model)
	if err != nil {
		return nil, fmt.Errorf("expertversion: load expert: %w", err)
	}

	var nextNum int
	if err := s.db.QueryRow(ctx,
		`SELECT COALESCE(MAX(version_number),0)+1 FROM expert_versions WHERE expert_id=$1`,
		expertID,
	).Scan(&nextNum); err != nil {
		return nil, err
	}

	var hasActive bool
	_ = s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM expert_versions WHERE expert_id=$1 AND is_active)`,
		expertID,
	).Scan(&hasActive)

	var prev *Version
	if hasActive {
		prev, _ = s.activeVersion(ctx, expertID)
	}

	topicsJSON, _ := json.Marshal(live.Topics)
	charterHash := sha256Hex(reasoning)
	v := &Version{
		ExpertID:             expertID,
		VersionNumber:        nextNum,
		Source:               source,
		CorpusHash:           live.CorpusHash,
		CharterHash:          charterHash,
		ChunkCount:           live.ChunkCount,
		TopicCount:           len(live.Topics),
		Topics:               live.Topics,
		ReasoningCharter:     reasoning,
		ClarificationCharter: normalize(clarification),
		CapabilitySummary:    normalize(capability),
		Model:                model,
		IsActive:             !hasActive,
		Notes:                notes,
		CreatedAt:            time.Now().UTC(),
	}

	err = s.db.QueryRow(ctx,
		`INSERT INTO expert_versions
			(expert_id, version_number, source, corpus_hash, charter_hash,
			 chunk_count, topic_count, topics, reasoning_charter,
			 clarification_charter, capability_summary, model, is_active, notes, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		 RETURNING id`,
		v.ExpertID, v.VersionNumber, v.Source, v.CorpusHash, v.CharterHash,
		v.ChunkCount, v.TopicCount, string(topicsJSON), v.ReasoningCharter,
		string(v.ClarificationCharter), string(v.CapabilitySummary),
		nullStr(v.Model), v.IsActive, v.Notes, v.CreatedAt,
	).Scan(&v.ID)
	if err != nil {
		return nil, fmt.Errorf("expertversion: insert: %w", err)
	}

	// Drift vs the previous active version (same corpus/charter families).
	if prev != nil {
		s.recordDrift(ctx, expertID, prev, v)
	}
	return v, nil
}

// DetectDrift compares the expert's live state to its active version and
// records a drift event when anything changed. Does not create a version.
func (s *Service) DetectDrift(ctx context.Context, expertID uuid.UUID) (*DriftEvent, error) {
	if !s.Enabled() || expertID == uuid.Nil {
		return nil, nil
	}
	prev, err := s.activeVersion(ctx, expertID)
	if err != nil || prev == nil {
		return nil, err
	}
	live, err := s.liveState(ctx, expertID)
	if err != nil {
		return nil, err
	}
	// Charter hash from the current experts row, so a charter edited
	// outside ingest is detected too.
	var reasoning string
	_ = s.db.QueryRow(ctx, `SELECT COALESCE(reasoning_charter,'') FROM experts WHERE id=$1`, expertID).Scan(&reasoning)
	cur := &Version{
		CorpusHash:  live.CorpusHash,
		CharterHash: sha256Hex(reasoning),
		Topics:      live.Topics,
		ChunkCount:  live.ChunkCount,
	}

	driftType, score, details, drifted := classifyDrift(prev, cur, s.threshold)
	if !drifted {
		return nil, nil
	}
	return s.insertDrift(ctx, expertID, &prev.ID, nil, driftType, score, details)
}

// Pin makes a version the canonical (active) one and rolls the expert's
// charter back to that version's snapshot. Corpus rollback is deferred
// (needs the original transcripts) — this restores charter + capability
// declaration, the cheap and reversible behavioural elements.
func (s *Service) Pin(ctx context.Context, expertID, versionID uuid.UUID) (*Version, error) {
	if !s.Enabled() {
		return nil, nil
	}
	v, err := s.getVersion(ctx, expertID, versionID)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`UPDATE expert_versions SET is_active=FALSE WHERE expert_id=$1 AND is_active`,
		expertID,
	); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE expert_versions SET is_active=TRUE WHERE id=$1`,
		versionID,
	); err != nil {
		return nil, err
	}
	// Charter rollback (declared behaviour).
	if _, err := tx.Exec(ctx,
		`UPDATE experts
		 SET reasoning_charter=$1, clarification_charter=$2, updated_at=NOW()
		 WHERE id=$3 AND deleted_at IS NULL`,
		v.ReasoningCharter, string(v.ClarificationCharter), expertID,
	); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	v.IsActive = true
	return v, nil
}

// ListVersions returns versions for an expert, newest first.
func (s *Service) ListVersions(ctx context.Context, expertID uuid.UUID, limit int) ([]Version, error) {
	if !s.Enabled() {
		return []Version{}, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.Query(ctx,
		`SELECT id, expert_id, version_number, source, corpus_hash, charter_hash,
		        chunk_count, topic_count, topics, reasoning_charter,
		        clarification_charter, capability_summary, COALESCE(model,''),
		        is_active, notes, created_at
		 FROM expert_versions
		 WHERE expert_id=$1
		 ORDER BY version_number DESC
		 LIMIT $2`,
		expertID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVersions(rows)
}

// ListDrift returns drift events for an expert, newest first.
func (s *Service) ListDrift(ctx context.Context, expertID uuid.UUID, limit int) ([]DriftEvent, error) {
	if !s.Enabled() {
		return []DriftEvent{}, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.Query(ctx,
		`SELECT id, expert_id, from_version_id, to_version_id,
		        drift_type, drift_score, details, acknowledged, created_at
		 FROM expert_drift_events
		 WHERE expert_id=$1
		 ORDER BY created_at DESC
		 LIMIT $2`,
		expertID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DriftEvent
	for rows.Next() {
		var d DriftEvent
		var details []byte
		if err := rows.Scan(&d.ID, &d.ExpertID, &d.FromVersionID, &d.ToVersionID,
			&d.DriftType, &d.DriftScore, &details, &d.Acknowledged, &d.CreatedAt); err != nil {
			continue
		}
		d.Details = normalize(details)
		out = append(out, d)
	}
	if out == nil {
		out = []DriftEvent{}
	}
	return out, rows.Err()
}

// AcknowledgeDrift marks one drift event as reviewed.
func (s *Service) AcknowledgeDrift(ctx context.Context, expertID, driftID uuid.UUID) error {
	if !s.Enabled() {
		return nil
	}
	_, err := s.db.Exec(ctx,
		`UPDATE expert_drift_events SET acknowledged=TRUE WHERE id=$1 AND expert_id=$2`,
		driftID, expertID,
	)
	return err
}

// ---- internals ----

func (s *Service) liveState(ctx context.Context, expertID uuid.UUID) (liveState, error) {
	var st liveState
	rows, err := s.db.Query(ctx,
		`SELECT COALESCE(chunk_hash,''), COALESCE(topic,'') FROM course_chunks WHERE expert_id=$1`,
		expertID,
	)
	if err != nil {
		return st, err
	}
	defer rows.Close()
	var hashes []string
	seen := map[string]struct{}{}
	for rows.Next() {
		var h, topic string
		if err := rows.Scan(&h, &topic); err != nil {
			continue
		}
		if h != "" {
			hashes = append(hashes, h)
		}
		if topic != "" {
			seen[topic] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return st, err
	}
	st.ChunkCount = len(hashes)
	sort.Strings(hashes)
	st.CorpusHash = sha256Hex(strings.Join(hashes, "\n"))
	topics := make([]string, 0, len(seen))
	for t := range seen {
		topics = append(topics, t)
	}
	sort.Strings(topics)
	st.Topics = topics
	return st, nil
}

func (s *Service) activeVersion(ctx context.Context, expertID uuid.UUID) (*Version, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, expert_id, version_number, source, corpus_hash, charter_hash,
		        chunk_count, topic_count, topics, reasoning_charter,
		        clarification_charter, capability_summary, COALESCE(model,''),
		        is_active, notes, created_at
		 FROM expert_versions
		 WHERE expert_id=$1 AND is_active
		 LIMIT 1`,
		expertID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	vs, err := scanVersions(rows)
	if err != nil || len(vs) == 0 {
		return nil, err
	}
	return &vs[0], nil
}

func (s *Service) getVersion(ctx context.Context, expertID, versionID uuid.UUID) (*Version, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, expert_id, version_number, source, corpus_hash, charter_hash,
		        chunk_count, topic_count, topics, reasoning_charter,
		        clarification_charter, capability_summary, COALESCE(model,''),
		        is_active, notes, created_at
		 FROM expert_versions
		 WHERE id=$1 AND expert_id=$2`,
		versionID, expertID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	vs, err := scanVersions(rows)
	if err != nil || len(vs) == 0 {
		return nil, err
	}
	return &vs[0], nil
}

// recordDrift computes and stores a drift event between two versions.
func (s *Service) recordDrift(ctx context.Context, expertID uuid.UUID, prev, cur *Version) {
	driftType, score, details, drifted := classifyDrift(prev, cur, s.threshold)
	if !drifted {
		return
	}
	if _, err := s.insertDrift(ctx, expertID, &prev.ID, &cur.ID, driftType, score, details); err != nil {
		s.logger.Warn("expert drift record failed",
			zap.String("expert_id", expertID.String()),
			zap.Error(err),
		)
	}
}

func (s *Service) insertDrift(
	ctx context.Context,
	expertID uuid.UUID,
	fromID, toID *uuid.UUID,
	driftType string,
	score float64,
	details map[string]interface{},
) (*DriftEvent, error) {
	b, _ := json.Marshal(details)
	var d DriftEvent
	err := s.db.QueryRow(ctx,
		`INSERT INTO expert_drift_events
			(expert_id, from_version_id, to_version_id, drift_type, drift_score, details)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 RETURNING id, created_at`,
		expertID, fromID, toID, driftType, score, string(b),
	).Scan(&d.ID, &d.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("expertversion: insert drift: %w", err)
	}
	d.ExpertID = expertID
	d.FromVersionID = fromID
	d.ToVersionID = toID
	d.DriftType = driftType
	d.DriftScore = score
	d.Details = json.RawMessage(b)
	s.logger.Info("expert drift detected",
		zap.String("expert_id", expertID.String()),
		zap.String("drift_type", driftType),
		zap.Float64("drift_score", score),
	)
	return &d, nil
}

// classifyDrift is pure (unit-tested): it decides whether the change from
// prev to cur counts as drift and, if so, its dominant type + score.
// Priority: capability (topic-set distance ≥ threshold) > charter > corpus.
func classifyDrift(prev, cur *Version, threshold float64) (string, float64, map[string]interface{}, bool) {
	if prev == nil || cur == nil {
		return "", 0, nil, false
	}
	topicDrift := jaccardDistance(prev.Topics, cur.Topics)
	corpusChanged := prev.CorpusHash != "" && prev.CorpusHash != cur.CorpusHash
	charterChanged := prev.CharterHash != "" && prev.CharterHash != cur.CharterHash

	details := map[string]interface{}{
		"corpus_changed":  corpusChanged,
		"charter_changed": charterChanged,
		"topic_drift":     topicDrift,
		"prev_topics":     len(prev.Topics),
		"cur_topics":      len(cur.Topics),
		"prev_chunks":     prev.ChunkCount,
		"cur_chunks":      cur.ChunkCount,
	}

	switch {
	case topicDrift >= threshold:
		return "capability", topicDrift, details, true
	case charterChanged:
		return "charter", 1.0, details, true
	case corpusChanged:
		return "corpus", 1.0, details, true
	default:
		return "", 0, details, false
	}
}

// jaccardDistance returns 1 - |A∩B|/|A∪B| over two string sets (0 = equal).
// Pure — unit-tested.
func jaccardDistance(a, b []string) float64 {
	sa := map[string]struct{}{}
	for _, x := range a {
		if x != "" {
			sa[x] = struct{}{}
		}
	}
	sb := map[string]struct{}{}
	for _, x := range b {
		if x != "" {
			sb[x] = struct{}{}
		}
	}
	if len(sa) == 0 && len(sb) == 0 {
		return 0
	}
	inter := 0
	for k := range sa {
		if _, ok := sb[k]; ok {
			inter++
		}
	}
	union := len(sa) + len(sb) - inter
	if union == 0 {
		return 0
	}
	return 1 - float64(inter)/float64(union)
}

// sha256Hex returns hex(sha256(text)). Pure — unit-tested.
func sha256Hex(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func scanVersions(rows interface {
	Next() bool
	Scan(...interface{}) error
	Err() error
}) ([]Version, error) {
	var out []Version
	for rows.Next() {
		var v Version
		var topics, clar, cap []byte
		if err := rows.Scan(
			&v.ID, &v.ExpertID, &v.VersionNumber, &v.Source, &v.CorpusHash, &v.CharterHash,
			&v.ChunkCount, &v.TopicCount, &topics, &v.ReasoningCharter,
			&clar, &cap, &v.Model, &v.IsActive, &v.Notes, &v.CreatedAt,
		); err != nil {
			continue
		}
		_ = json.Unmarshal(topics, &v.Topics)
		if v.Topics == nil {
			v.Topics = []string{}
		}
		v.ClarificationCharter = normalize(clar)
		v.CapabilitySummary = normalize(cap)
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func normalize(b []byte) json.RawMessage {
	if len(b) == 0 || string(b) == "null" {
		return json.RawMessage("{}")
	}
	return json.RawMessage(b)
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
