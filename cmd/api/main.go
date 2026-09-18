package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ErenKarakus1/File-Management-Platform/internal/auth"
	"github.com/ErenKarakus1/File-Management-Platform/internal/config"
	"github.com/ErenKarakus1/File-Management-Platform/internal/database"
	"github.com/ErenKarakus1/File-Management-Platform/internal/files"
	"github.com/ErenKarakus1/File-Management-Platform/internal/ratelimit"
	"github.com/ErenKarakus1/File-Management-Platform/internal/repository"
	"github.com/ErenKarakus1/File-Management-Platform/internal/server"
	"github.com/ErenKarakus1/File-Management-Platform/internal/workers"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	redisClient := ratelimit.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	rateLimiter := ratelimit.NewLimiter(redisClient, cfg.RateLimit, cfg.RateLimitWindow)
	if err := rateLimiter.Ping(ctx); err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer rateLimiter.Close()

	authService := auth.NewService(db, cfg.JWTSecret, cfg.TokenTTL)
	fileRepository := repository.NewFileRepository(db)
	fileService := files.NewService(fileRepository, cfg.FileStorageDir)
	fileHandler := files.NewHandler(fileService, cfg.MaxUploadBytes)
	fileWorker := workers.NewFileWorker(fileRepository, cfg.FileWorkerCount, cfg.FileDeleteWorkerCount, cfg.FileWorkerPoll)
	waitForFileWorkers := fileWorker.Start(ctx)
	router := server.New(authService, fileHandler, rateLimiter.Middleware())

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("api listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	waitForFileWorkers()
}
