package storage

import (
	"context"
	"github.com/redis/go-redis/v9"
)

// RedisAdapter implements mcpv2.RedisPort — atomic counters + Pub/Sub wakeup only (RULE 8-F:46).
type RedisAdapter struct{ client *redis.Client }

func NewRedisAdapter(c *redis.Client) *RedisAdapter { return &RedisAdapter{client: c} }

func (r *RedisAdapter) IncrTokens(ctx context.Context, key string, delta int64) (int64, error) {
	return r.client.IncrBy(ctx, key, delta).Result()
}

func (r *RedisAdapter) Publish(ctx context.Context, channel string, payload []byte) error {
	return r.client.Publish(ctx, channel, payload).Err()
}