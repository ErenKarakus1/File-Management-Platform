package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr              string
	DatabaseURL           string
	JWTSecret             string
	TokenTTL              time.Duration
	FileStorageDir        string
	MaxUploadBytes        int64
	FileWorkerCount       int
	FileDeleteWorkerCount int
	FileWorkerPoll        time.Duration
	RedisAddr             string
	RedisPassword         string
	RedisDB               int
	RateLimit             int
	RateLimitWindow       time.Duration
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		HTTPAddr:              getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/file_management?sslmode=disable"),
		JWTSecret:             getEnv("JWT_SECRET", "dev-secret-change-me"),
		TokenTTL:              getDurationEnv("TOKEN_TTL", 24*time.Hour),
		FileStorageDir:        getEnv("FILE_STORAGE_DIR", "storage"),
		MaxUploadBytes:        getInt64Env("MAX_UPLOAD_BYTES", 50<<20),
		FileWorkerCount:       getIntEnv("FILE_WORKER_COUNT", 2),
		FileDeleteWorkerCount: getIntEnv("FILE_DELETE_WORKER_COUNT", 1),
		FileWorkerPoll:        getDurationEnv("FILE_WORKER_POLL", time.Second),
		RedisAddr:             getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:         getEnv("REDIS_PASSWORD", ""),
		RedisDB:               getIntEnv("REDIS_DB", 0),
		RateLimit:             getIntEnv("RATE_LIMIT", 120),
		RateLimitWindow:       getDurationEnv("RATE_LIMIT_WINDOW", time.Minute),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}

func getInt64Env(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func getIntEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
