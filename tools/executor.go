package tools

import (
	"context"
	"errors"
	"fmt"
)

var ErrUnauthorized = errors.New("tool execution unauthorized")
var ErrUnknownTool = errors.New("unknown tool")

type ToolRequest struct {
	Name   string
	Amount float64
	// In production, we'd have generic Args here as well.
}

type ToolResponse struct {
	Status string `json:"status"`
	Result string `json:"result"`
}

type Executor interface {
	Execute(ctx context.Context, req ToolRequest) (ToolResponse, error)
}

type MockExecutor struct{}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{}
}

func (m *MockExecutor) Execute(ctx context.Context, req ToolRequest) (ToolResponse, error) {
	switch req.Name {
	case "get_payment":
		return ToolResponse{Status: "success", Result: "payment details..."}, nil
	case "create_order":
		return ToolResponse{Status: "success", Result: fmt.Sprintf("order created for amount: %.2f", req.Amount)}, nil
	case "capture_payment":
		return ToolResponse{Status: "success", Result: "payment captured"}, nil
	case "create_refund":
		return ToolResponse{Status: "success", Result: fmt.Sprintf("refund processed for amount: %.2f", req.Amount)}, nil
	case "fetch_settlement":
		return ToolResponse{Status: "success", Result: "settlement details..."}, nil
	default:
		return ToolResponse{}, ErrUnknownTool
	}
}
