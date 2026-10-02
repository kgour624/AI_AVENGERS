package mcpv2

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisAdapter implements RedisPort via go-redis — atomic counters + Pub/Sub (Ultimate Go §F).
type RedisAdapter struct{ client *redis.Client }

func NewRedisAdapter(c *redis.Client) *RedisAdapter { return &RedisAdapter{client: c} }

func (r *RedisAdapter) IncrTokens(ctx context.Context, key string, delta int64) (int64, error) {
	return r.client.IncrBy(ctx, key, delta).Result()
}
func (r *RedisAdapter) Publish(ctx context.Context, channel string, payload []byte) error {
	return r.client.Publish(ctx, channel, payload).Err()
}
func (r *RedisAdapter) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}
func (r *RedisAdapter) Set(ctx context.Context, key string, value string, ttlSeconds int) error {
	return r.client.Set(ctx, key, value, time.Duration(ttlSeconds)*time.Second).Err()
}
func (r *RedisAdapter) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
