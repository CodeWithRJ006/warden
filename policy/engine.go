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

// HardcodedEngine implements Engine with static rules for Day 1.
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

	// Rule 1: Admins can do anything
	if req.Role == "admin" {
		return Result{Decision: DecisionAllow, Reason: "admin bypass"}, nil
	}

	// Support agent rules
	if req.Role == "support-agent" {
		switch req.Action {
		case "get_payment":
			return Result{Decision: DecisionAllow, Reason: "support agents can view payments"}, nil
		case "create_refund":
			// Rule 2: Support agents can refund up to $100 without approval
			if req.Amount > 0 && req.Amount <= 100 {
				return Result{Decision: DecisionAllow, Reason: "refund amount within support limits"}, nil
			}
			// Rule 3: Support agents require approval for refunds over $100
			if req.Amount > 100 {
				return Result{Decision: DecisionRequireApproval, Reason: "refund amount exceeds support limits"}, nil
			}
			// Default deny for refunds without amount
			return Result{Decision: DecisionDeny, Reason: "support agents cannot issue unquantified refunds"}, nil
		case "fetch_settlement":
			return Result{Decision: DecisionRedactAndAllow, Reason: "support agents can view settlements but PII is redacted"}, nil
		default:
			// Default deny
			return Result{Decision: DecisionDeny, Reason: "action not allowed for support agent"}, nil
		}
	}

	// Rule 4: Read-only users can only get data
	if req.Role == "viewer" {
		if req.Action == "get_payment" || req.Action == "fetch_settlement" {
			return Result{Decision: DecisionAllow, Reason: "viewers can view data"}, nil
		}
		return Result{Decision: DecisionDeny, Reason: "viewers cannot modify data"}, nil
	}

	// Default policy: deny
	return Result{Decision: DecisionDeny, Reason: "no matching policy found"}, nil
}
