package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/CodeWithRJ006/warden/audit"
	"github.com/CodeWithRJ006/warden/middleware"
	"github.com/CodeWithRJ006/warden/pii"
	"github.com/CodeWithRJ006/warden/policy"
	"github.com/CodeWithRJ006/warden/tools"
)

type ExecuteRequest struct {
	Actor  string  `json:"actor"`
	Tool   string  `json:"tool"`
	Amount float64 `json:"amount,omitempty"`
}

type ExecuteResponse struct {
	Decision   policy.Decision `json:"decision"`
	Reason     string          `json:"reason"`
	ToolStatus string          `json:"tool_status,omitempty"`
	ToolResult string          `json:"tool_result,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// HandleExecute returns the HTTP handler for /v1/tools/execute.
func HandleExecute(engine policy.Engine, piiProc pii.Processor, executor tools.Executor, logger audit.Logger, limiter middleware.RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := generateRequestID()

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ExecuteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		// Validation
		if req.Actor == "" || req.Tool == "" {
			writeJSONError(w, http.StatusBadRequest, "Missing required fields: actor and tool")
			return
		}

		// Rate Limiting (10 req / minute per actor)
		allowed, err := limiter.Allow(r.Context(), req.Actor, 10, time.Minute)
		if err != nil || !allowed {
			logEvent(r.Context(), logger, reqID, req, audit.PolicyTrace{Decision: "DENY", Rule: "rate_limit"}, audit.ModelTrace{}, audit.ToolTrace{}, "RATE_LIMITED")
			writeJSONError(w, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		}

		// PII Processing
		sanitizedTool, err := piiProc.Process(r.Context(), req.Tool)
		if err != nil {
			logEvent(r.Context(), logger, reqID, req, audit.PolicyTrace{Decision: "ERROR", Rule: "pii_processing"}, audit.ModelTrace{}, audit.ToolTrace{}, "ERROR")
			writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		// Policy Evaluation
		policyReq := policy.RequestCtx{
			Role:   req.Actor,
			Action: sanitizedTool,
			Amount: req.Amount,
		}

		result, err := engine.Evaluate(r.Context(), policyReq)
		if err != nil {
			logEvent(r.Context(), logger, reqID, req, audit.PolicyTrace{Decision: "ERROR", Rule: "policy_engine"}, audit.ModelTrace{}, audit.ToolTrace{}, "ERROR")
			writeJSONError(w, http.StatusInternalServerError, "Policy evaluation failed")
			return
		}

		// Tool Execution (if allowed)
		var toolResp tools.ToolResponse
		toolExecuted := false
		if result.Decision == policy.DecisionAllow || result.Decision == policy.DecisionRedactAndAllow {
			toolReq := tools.ToolRequest{
				Name:   req.Tool,
				Amount: req.Amount,
			}
			toolResp, err = executor.Execute(r.Context(), toolReq)
			toolExecuted = true
			if err != nil {
				logEvent(r.Context(), logger, reqID, req, audit.PolicyTrace{Decision: string(result.Decision), Rule: result.Reason}, audit.ModelTrace{}, audit.ToolTrace{Executed: true, Status: "FAILED"}, "ERROR")
				writeJSONError(w, http.StatusInternalServerError, "Tool execution failed")
				return
			}
		}

		// Audit Event
		logEvent(r.Context(), logger, reqID, req, audit.PolicyTrace{Decision: string(result.Decision), Rule: result.Reason}, audit.ModelTrace{}, audit.ToolTrace{Executed: toolExecuted, Status: toolResp.Status}, string(result.Decision))

		// Response
		resp := ExecuteResponse{
			Decision:   result.Decision,
			Reason:     result.Reason,
			ToolStatus: toolResp.Status,
			ToolResult: toolResp.Result,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func generateRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "req_" + hex.EncodeToString(b)
}

func logEvent(ctx context.Context, logger audit.Logger, reqID string, req ExecuteRequest, policyTrace audit.PolicyTrace, modelTrace audit.ModelTrace, toolTrace audit.ToolTrace, finalDecision string) {
	dataClass := "clean"
	if strings.Contains(req.Tool, "[") {
		dataClass = "restricted"
	}

	event := audit.Event{
		RequestID:          reqID,
		Actor:              req.Actor,
		Tool:               req.Tool,
		PolicyVersion:      "v1.4",
		DataClassification: dataClass,
		Policy:             policyTrace,
		Model:              modelTrace,
		ToolExec:           toolTrace,
		FinalDecision:      finalDecision,
		Timestamp:          time.Now(),
	}
	_ = logger.Log(ctx, event)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}
