package pii

import "context"

// Processor defines the interface for detecting and tokenizing PII.
type Processor interface {
	Process(ctx context.Context, input string) (string, error)
}

// MockProcessor is a stub implementation for Day 2.
type MockProcessor struct{}

func NewMockProcessor() *MockProcessor {
	return &MockProcessor{}
}

func (m *MockProcessor) Process(ctx context.Context, input string) (string, error) {
	// Day 2 mock: return the input unchanged.
	// We will wire this to the actual Presidio API in Day 3.
	return input, nil
}
