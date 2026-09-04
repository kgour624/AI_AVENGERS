package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// L1Memory is the hot per-expert per-project memory stored in Redis.
// Contains last 5 key decisions and current context summary.
// TTL: 24 hours, refreshed on every access.
//
// WHY Redis for L1:
// Accessed on EVERY turn. Must be sub-millisecond.
// PostgreSQL would add 5-10ms per turn — unacceptable for hot path.
type L1Memory struct {
	ProjectID    uuid.UUID    `json:"project_id"`
	ExpertID     uuid.UUID    `json:"expert_id"`
	KeyDecisions []L1Decision `json:"key_decisions"`
	ContextSummary string     `json:"context_summary"`
	LastUpdated  time.Time    `json:"last_updated"`
}

// L1Decision is a single key decision stored in L1.
type L1Decision struct {
	Content     string    `json:"content"`
	Reasoning   string    `json:"reasoning"`
	TurnNumber  int       `json:"turn_number"`
	MemoryType  string    `json:"memory_type"`
	Importance  int       `json:"importance"`
	CreatedAt   time.Time `json:"created_at"`
}

const (
	l1TTL      = 24 * time.Hour
	l1MaxItems = 5
)

// L1Store handles Redis-backed L1 memory operations.
type L1Store struct {
	redis  *redis.Client
	logger *zap.Logger
}

// NewL1Store creates a new L1 store.
func NewL1Store(redisClient *redis.Client, logger *zap.Logger) *L1Store {
	return &L1Store{redis: redisClient, logger: logger}
}

// l1Key builds the Redis key for L1 memory.
// Pattern: l1:{project_id}:{expert_id}
func l1Key(projectID, expertID uuid.UUID) string {
	return fmt.Sprintf("l1:%s:%s", projectID, expertID)
}

// Get retrieves L1 memory. Returns empty memory if not found (cache miss).
func (s *L1Store) Get(ctx context.Context, projectID, expertID uuid.UUID) (*L1Memory, error) {
	key := l1Key(projectID, expertID)

	data, err := s.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// Cache miss — return empty memory
		return &L1Memory{
			ProjectID: projectID,
			ExpertID:  expertID,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}

	var mem L1Memory
	if err := json.Unmarshal(data, &mem); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	// Refresh TTL on access — keep hot memory alive
	s.redis.Expire(ctx, key, l1TTL)

	return &mem, nil
}

// Append adds a new decision to L1 memory.
// Keeps only the last l1MaxItems decisions (FIFO eviction).
func (s *L1Store) Append(ctx context.Context, projectID, expertID uuid.UUID, decision L1Decision) error {
	mem, err := s.Get(ctx, projectID, expertID)
	if err != nil {
		return err
	}

	decision.CreatedAt = time.Now()
	mem.KeyDecisions = append(mem.KeyDecisions, decision)
	mem.LastUpdated = time.Now()

	// Keep only last l1MaxItems
	// WHY FIFO: Most recent decisions are most relevant
	if len(mem.KeyDecisions) > l1MaxItems {
		mem.KeyDecisions = mem.KeyDecisions[len(mem.KeyDecisions)-l1MaxItems:]
	}

	return s.save(ctx, mem)
}

// UpdateSummary updates the context summary in L1.
func (s *L1Store) UpdateSummary(ctx context.Context, projectID, expertID uuid.UUID, summary string) error {
	mem, err := s.Get(ctx, projectID, expertID)
	if err != nil {
		return err
	}
	mem.ContextSummary = summary
	mem.LastUpdated = time.Now()
	return s.save(ctx, mem)
}

// Invalidate removes L1 memory (called when project is deleted).
func (s *L1Store) Invalidate(ctx context.Context, projectID, expertID uuid.UUID) error {
	return s.redis.Del(ctx, l1Key(projectID, expertID)).Err()
}

// save serializes and stores L1 memory in Redis.
func (s *L1Store) save(ctx context.Context, mem *L1Memory) error {
	data, err := json.Marshal(mem)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}
	return s.redis.Set(ctx, l1Key(mem.ProjectID, mem.ExpertID), data, l1TTL).Err()
}
