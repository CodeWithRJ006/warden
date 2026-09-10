package audit

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

// Event represents a structured audit log entry.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Actor     string    `json:"actor"`
	Action    string    `json:"action"`
	Amount    float64   `json:"amount,omitempty"`
	Decision  string    `json:"decision"`
	Reason    string    `json:"reason"`
	PolicyVer string    `json:"policy_version"`
}

// Logger defines the interface for recording audit events.
type Logger interface {
	Log(ctx context.Context, event Event) error
}

// StdoutLogger is a simple logger that writes to stdout.
type StdoutLogger struct{}

func NewStdoutLogger() *StdoutLogger {
	return &StdoutLogger{}
}

func (l *StdoutLogger) Log(ctx context.Context, event Event) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	log.Printf("[AUDIT] %s\n", string(b))
	return nil
}
