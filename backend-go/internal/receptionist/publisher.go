package receptionist

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
)

type Publisher struct {
    rdb    *redis.Client
    logger *zap.Logger
}

func NewPublisher(rdb *redis.Client, logger *zap.Logger) *Publisher {
    return &Publisher{rdb: rdb, logger: logger}
}

// NotifyOnly - Postgres likhne ke BAAD sirf ID bhejo, pura data nahi
func (p *Publisher) NotifyNewEvent(ctx context.Context, sessionID string, eventID int64) error {
    ch := fmt.Sprintf("session:%s:events", sessionID)
    msg, _ := json.Marshal(map[string]any{"event_id": eventID})
    // zap lowercase msg - steering rule
    p.logger.Info("publish notify", zap.String("channel", ch), zap.Int64("event_id", eventID))
    return p.rdb.Publish(ctx, ch, msg).Err()
}

func (p *Publisher) Subscribe(ctx context.Context, sessionID string) *redis.PubSub {
    ch := fmt.Sprintf("session:%s:events", sessionID)
    return p.rdb.Subscribe(ctx, ch)
}
