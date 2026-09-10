package policy

import (
	"context"
	"errors"
)

// Decision represents the result of a policy evaluation.
type Decision string

const (
	DecisionAllow           Decision = "ALLOW"
	DecisionDeny            Decision = "DENY"
	DecisionRequireApproval Decision = "REQUIRE_APPROVAL"
	DecisionRedactAndAllow  Decision = "REDACT_AND_ALLOW"
)

// RequestCtx contains the context of the action being requested.
type RequestCtx struct {
	Role   string
	Action string
	Amount float64 // Optional, used for some policies
}

// Result contains the decision and the reason.
type Result struct {
	Decision Decision
	Reason   string
}

// Engine defines the interface for evaluating policies.
type Engine interface {
	Evaluate(ctx context.Context, req RequestCtx) (Result, error)
}

// HardcodedEngine implements Engine with static rules.
type HardcodedEngine struct{}

// NewHardcodedEngine creates a new HardcodedEngine.
func NewHardcodedEngine() *HardcodedEngine {
	return &HardcodedEngine{}
}

// Evaluate applies hardcoded rules to the incoming request.
func (e *HardcodedEngine) Evaluate(ctx context.Context, req RequestCtx) (Result, error) {
	if req.Role == "" || req.Action == "" {
		return Result{}, errors.New("role and action are required")
	}

	// Finance Manager rules (high privilege)
	if req.Role == "finance-manager" {
		switch req.Action {
		case "get_payment", "capture_payment", "fetch_settlement":
			return Result{Decision: DecisionAllow, Reason: "finance manager has full access to this action"}, nil
		case "create_refund":
			if req.Amount > 0 && req.Amount <= 100000 {
				return Result{Decision: DecisionAllow, Reason: "refund amount within manager limits"}, nil
			}
			if req.Amount > 100000 {
				return Result{Decision: DecisionRequireApproval, Reason: "refund amount exceeds manager limits"}, nil
			}
			return Result{Decision: DecisionDeny, Reason: "unquantified refund not allowed"}, nil
		}
	}

	// Finance Operator rules (medium privilege)
	if req.Role == "finance-operator" {
		switch req.Action {
		case "get_payment", "capture_payment":
			return Result{Decision: DecisionAllow, Reason: "finance operator can process payments"}, nil
		case "create_refund":
			if req.Amount > 0 && req.Amount <= 10000 {
				return Result{Decision: DecisionAllow, Reason: "refund amount within operator limits"}, nil
			}
			if req.Amount > 10000 {
				return Result{Decision: DecisionRequireApproval, Reason: "refund amount exceeds operator limits"}, nil
			}
			return Result{Decision: DecisionDeny, Reason: "unquantified refund not allowed"}, nil
		}
	}

	// Support agent rules (low privilege)
	if req.Role == "support-agent" {
		switch req.Action {
		case "get_payment":
			return Result{Decision: DecisionAllow, Reason: "support agents can view payments"}, nil
		case "fetch_settlement":
			return Result{Decision: DecisionRedactAndAllow, Reason: "support agents can view settlements but PII is redacted"}, nil
		}
	}

	// Viewer rules (read-only)
	if req.Role == "viewer" {
		if req.Action == "get_payment" || req.Action == "fetch_settlement" {
			return Result{Decision: DecisionRedactAndAllow, Reason: "viewers can view data with PII redacted"}, nil
		}
	}

	// Default policy: deny
	return Result{Decision: DecisionDeny, Reason: "no matching policy found or action not allowed for role"}, nil
}
