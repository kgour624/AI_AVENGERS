package llm

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// DBCandidateSink persists LLM suggested kachra to candidate_kachra (status=pending).
// I: []KachraSpan + chunkText -> P: UPSERT on unique(pattern,pattern_type,category) -> O: rows written.
// Future seed: human approves -> inserts into kachra_patterns -> bump_brain_version trigger -> 30s hot-reload.
type DBCandidateSink struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewDBCandidateSink(db *pgxpool.Pool, logger *zap.Logger) *DBCandidateSink {
	return &DBCandidateSink{db: db, logger: logger}
}

func (s *DBCandidateSink) Save(ctx context.Context, spans []KachraSpan, chunkText string) (int, error) {
	return s.SaveWithFileID(ctx, "", spans, chunkText)
}

func (s *DBCandidateSink) SaveWithFileID(ctx context.Context, fileJobID string, spans []KachraSpan, chunkText string) (int, error) {
	if len(spans) == 0 || s.db == nil {
		return 0, nil
	}
	snippet := truncate(chunkText, 500)
	var fileID *uuid.UUID
	if fileJobID != "" {
		if id, err := uuid.Parse(fileJobID); err == nil {
			fileID = &id
		}
	}
	n := 0
	for _, sp := range spans {
		pat := strings.TrimSpace(sp.Text)
		if pat == "" {
			continue
		}
		ptype := "PHRASE"
		cat := strings.ToLower(strings.TrimSpace(sp.Type))
		if !validKachraType[cat] {
			cat = "filler"
		}
		if cat == "" {
			cat = "filler"
		}
		conf := sp.Confidence
		if conf < 0 {
			conf = 0
		}
		if conf > 1 {
			conf = 1
		}
		_, err := s.db.Exec(ctx, `
			INSERT INTO candidate_kachra(pattern, pattern_type, category, context_snippet, file_id, confidence, status, hit_count)
			VALUES ($1,$2,$3,$4,$5,$6,'pending',1)
			ON CONFLICT (lower(trim(pattern)), pattern_type, category) DO UPDATE
			SET hit_count = candidate_kachra.hit_count+1,
			    confidence = GREATEST(candidate_kachra.confidence, EXCLUDED.confidence),
			    context_snippet = EXCLUDED.context_snippet,
			    file_id = COALESCE(EXCLUDED.file_id, candidate_kachra.file_id),
			    last_seen_at = NOW(),
			    updated_at = NOW()
		`, pat, ptype, cat, snippet, fileID, conf)
		if err != nil {
			if s.logger != nil {
				s.logger.Warn("candidate_kachra sink failed", zap.String("pattern", pat), zap.Error(err))
			}
			continue
		}
		n++
	}
	if s.logger != nil && n > 0 {
		s.logger.Info("kachra sink: saved candidates", zap.Int("count", n), zap.String("file_job_id", fileJobID))
	}
	return n, nil
}

// NoopSink is used when DB is nil (dev/test) — satisfies interface, writes nowhere.
type NoopSink struct{}

func (NoopSink) Save(_ context.Context, _ []KachraSpan, _ string) (int, error) { return 0, nil }
func (NoopSink) SaveWithFileID(_ context.Context, _ string, _ []KachraSpan, _ string) (int, error) {
	return 0, nil
}
