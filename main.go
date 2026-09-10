package main

import (
	"log"
	"net/http"

	"github.com/CodeWithRJ006/warden/audit"
	"github.com/CodeWithRJ006/warden/pii"
	"github.com/CodeWithRJ006/warden/policy"
	"github.com/CodeWithRJ006/warden/server"
)

func main() {
	engine := policy.NewHardcodedEngine()
	piiProc := pii.NewMockProcessor()
	logger := audit.NewStdoutLogger()

	http.HandleFunc("/v1/tools/execute", server.HandleExecute(engine, piiProc, logger))

	log.Println("Warden Gateway starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
