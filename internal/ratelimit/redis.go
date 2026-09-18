package ratelimit

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const incrementScript = `
local count = redis.call("INCR", KEYS[1])
if count == 1 then
	redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("PTTL", KEYS[1])
return {count, ttl}
`

type Limiter struct {
	redis  *redis.Client
	limit  int
	window time.Duration
}

func NewLimiter(redisClient *redis.Client, limit int, window time.Duration) *Limiter {
	return &Limiter{
		redis:  redisClient,
		limit:  limit,
		window: window,
	}
}

func NewRedisClient(addr, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if l.limit <= 0 || l.window <= 0 {
			c.Next()
			return
		}

		key := "rate_limit:" + c.ClientIP()
		count, ttl, err := l.increment(c.Request.Context(), key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "rate limiter unavailable"})
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(l.limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(max(l.limit-count, 0)))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))

		if count > l.limit {
			c.Header("Retry-After", strconv.FormatInt(int64(ttl.Seconds()), 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		c.Next()
	}
}

func (l *Limiter) Ping(ctx context.Context) error {
	return l.redis.Ping(ctx).Err()
}

func (l *Limiter) Close() error {
	return l.redis.Close()
}

func (l *Limiter) increment(ctx context.Context, key string) (int, time.Duration, error) {
	result, err := l.redis.Eval(ctx, incrementScript, []string{key}, l.window.Milliseconds()).Result()
	if err != nil {
		return 0, 0, err
	}

	values, ok := result.([]any)
	if !ok || len(values) != 2 {
		return 0, 0, errors.New("unexpected redis rate limit response")
	}

	count, ok := values[0].(int64)
	if !ok {
		return 0, 0, errors.New("unexpected redis rate limit count")
	}

	ttlMillis, ok := values[1].(int64)
	if !ok {
		return 0, 0, errors.New("unexpected redis rate limit ttl")
	}

	remainingTTL := time.Duration(ttlMillis) * time.Millisecond
	if remainingTTL <= 0 {
		remainingTTL = l.window
	}

	return int(count), remainingTTL, nil
}
