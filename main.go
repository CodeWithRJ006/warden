package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/CodeWithRJ006/warden/audit"
	"github.com/CodeWithRJ006/warden/middleware"
	"github.com/CodeWithRJ006/warden/pii"
	"github.com/CodeWithRJ006/warden/policy"
	"github.com/CodeWithRJ006/warden/server"
	"github.com/CodeWithRJ006/warden/tools"
)

func main() {
	engine := policy.NewHardcodedEngine()
	var vault pii.Vault
	var logger audit.Logger

	ctx := context.Background()
	dbURL := os.Getenv("DB_URL")
	if dbURL != "" {
		pgVault, err := pii.NewPostgresVault(ctx, dbURL)
		if err != nil {
			log.Fatalf("Failed to connect to DB for vault: %v", err)
		}
		vault = pgVault
		
		pgLogger, err := audit.NewPostgresLogger(ctx, dbURL)
		if err != nil {
			log.Fatalf("Failed to connect to DB for logger: %v", err)
		}
		logger = pgLogger
		log.Println("Using Postgres for Vault and Audit Logging")
	} else {
		vault = pii.NewMemoryVault()
		logger = audit.NewStdoutLogger()
		log.Println("WARNING: Using in-memory vault and stdout logger")
	}

	piiProc := pii.NewTokenizer(vault)
	executor := tools.NewMockExecutor()
	var limiter middleware.RateLimiter
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		rl, err := middleware.NewRedisRateLimiter(redisURL)
		if err != nil {
			log.Fatalf("Failed to connect to Redis: %v", err)
		}
		limiter = rl
		log.Println("Using Redis for distributed rate limiting")
	} else {
		limiter = middleware.NewMemoryRateLimiter()
		log.Println("WARNING: Using in-memory rate limiter (not suitable for multiple instances)")
	}

	http.HandleFunc("/v1/tools/execute", server.HandleExecute(engine, piiProc, executor, logger, limiter))

	log.Println("Warden Gateway starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
