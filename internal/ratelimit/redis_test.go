package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestIncrementUsesFixedWindow(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	limiter := NewLimiter(client, 2, time.Minute)

	count, ttl, err := limiter.increment(context.Background(), "rate_limit:test")
	if err != nil {
		t.Fatalf("first increment: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
	if ttl <= 0 || ttl > time.Minute {
		t.Fatalf("unexpected ttl %s", ttl)
	}

	server.FastForward(10 * time.Second)

	count, ttl, err = limiter.increment(context.Background(), "rate_limit:test")
	if err != nil {
		t.Fatalf("second increment: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}
	if ttl > 51*time.Second {
		t.Fatalf("expected fixed window ttl to keep counting down, got %s", ttl)
	}
}

func TestMiddlewareRejectsOverLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	limiter := NewLimiter(client, 1, time.Minute)

	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	first := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	router.ServeHTTP(first, req)
	if first.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d", first.Code)
	}

	second := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	router.ServeHTTP(second, req)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request 429, got %d", second.Code)
	}
	if second.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}
