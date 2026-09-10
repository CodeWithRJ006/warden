package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	client *redis.Client
}

func NewRedisRateLimiter(redisURL string) (*RedisRateLimiter, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	// Quick ping to verify connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &RedisRateLimiter{client: client}, nil
}

var limitScript = redis.NewScript(`
local current
current = redis.call("incr",KEYS[1])
if tonumber(current) == 1 then
    redis.call("pexpire",KEYS[1],ARGV[1])
end
if tonumber(current) > tonumber(ARGV[2]) then
    return 0
end
return 1
`)

func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	redisKey := fmt.Sprintf("ratelimit:%s", key)
	res, err := limitScript.Run(ctx, r.client, []string{redisKey}, int64(window/time.Millisecond), limit).Result()
	if err != nil {
		return false, err
	}
	return res.(int64) == 1, nil
}
