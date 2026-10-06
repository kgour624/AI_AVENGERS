package brain

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Store owns the versioned Trie snapshot (RWMutex + copy-on-write).
// Heavy reads (Search) hold RLock; write (Reload) builds new trie then swaps under Lock.
// Mirrors DTS "global highlight via versioned Trie hot-reload" design.
type Store struct {
	mu      sync.RWMutex
	trie    *Trie
	version int64
	db      *pgxpool.Pool
	logger  *zap.Logger
}

func NewStore(db *pgxpool.Pool, logger *zap.Logger) *Store {
	return &Store{trie: NewTrie(), db: db, logger: logger}
}

// Snapshot returns current trie + version under RLock. Caller must not mutate.
func (s *Store) Snapshot() (*Trie, int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.trie, s.version
}

// Version returns current brain version.
func (s *Store) Version() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

// Reload rebuilds trie from kachra_patterns where is_active=true.
// Called on startup and every 30s hot-reload tick (or on trigger).
func (s *Store) Reload(ctx context.Context) error {
	rows, err := s.db.Query(ctx, `SELECT id::text, pattern, pattern_type FROM kachra_patterns WHERE is_active=true ORDER BY pattern`)
	if err != nil {
		return err
	}
	defer rows.Close()
	nt := NewTrie()
	for rows.Next() {
		var id, pat, ptype string
		if err := rows.Scan(&id, &pat, &ptype); err != nil {
			continue
		}
		nt.Add(pat, id, ptype)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	nt.Build()
	// Fetch version from brain_version
	var v int64
	_ = s.db.QueryRow(ctx, `SELECT version FROM brain_version WHERE id=1`).Scan(&v)
	s.mu.Lock()
	s.trie = nt
	s.version = v
	s.mu.Unlock()
	s.logger.Info("vacuum brain reloaded", zap.Int("patterns", nt.Size()), zap.Int64("version", v))
	return nil
}

// Search is convenience: snapshot + search (RLock only).
func (s *Store) Search(text string) []Match {
	trie, _ := s.Snapshot()
	return trie.Search(text)
}

// Start launches 30s hot-reload ticker until ctx done.
func (s *Store) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := s.Reload(ctx); err != nil {
				s.logger.Warn("vacuum brain hot-reload failed", zap.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}
}
