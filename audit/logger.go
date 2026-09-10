package audit

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

type PolicyTrace struct {
	Decision string `json:"decision"`
	Rule     string `json:"rule"`
}

type ModelTrace struct {
	Invoked   bool   `json:"invoked"`
	ModelName string `json:"model_name,omitempty"`
	Status    string `json:"status,omitempty"`
}

type ToolTrace struct {
	Executed bool   `json:"executed"`
	Status   string `json:"status,omitempty"`
}

// Event represents a structured audit log entry (Decision Trace).
type Event struct {
	RequestID          string      `json:"request_id"`
	Actor              string      `json:"actor"`
	Tool               string      `json:"requested_tool"`
	PolicyVersion      string      `json:"policy_version"`
	DataClassification string      `json:"data_classification"`
	Policy             PolicyTrace `json:"policy"`
	Model              ModelTrace  `json:"model"`
	ToolExec           ToolTrace   `json:"tool"`
	FinalDecision      string      `json:"final_decision"`
	Timestamp          time.Time   `json:"timestamp"`
}

// Logger defines the interface for recording audit events.
type Logger interface {
	Log(ctx context.Context, event Event) error
	GetLogs(ctx context.Context, limit int) ([]Event, error)
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


func (l *StdoutLogger) GetLogs(ctx context.Context, limit int) ([]Event, error) { return []Event{}, nil }

