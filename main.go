package main

import (
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
	vault := pii.NewMemoryVault()
	piiProc := pii.NewTokenizer(vault)
	logger := audit.NewStdoutLogger()
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
