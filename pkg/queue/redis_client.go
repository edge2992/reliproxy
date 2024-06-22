package queue

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient interface {
	LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	BRPop(ctx context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd
}
