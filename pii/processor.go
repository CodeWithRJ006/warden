package pii

import "context"

// Processor defines the interface for detecting and tokenizing PII.
type Processor interface {
	Process(ctx context.Context, input string) (string, error)
}
