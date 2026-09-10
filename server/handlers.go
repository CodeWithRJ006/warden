package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
			logEvent(r.Context(), logger, req, "RATE_LIMITED", "exceeded rate limit")
			writeJSONError(w, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		}

		// PII Processing (Mocked/Regex)
		sanitizedTool, err := piiProc.Process(r.Context(), req.Tool)
		if err != nil {
			logEvent(r.Context(), logger, req, "ERROR", "PII processing failed")
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
			logEvent(r.Context(), logger, req, "ERROR", err.Error())
			writeJSONError(w, http.StatusInternalServerError, "Policy evaluation failed")
			return
		}

		// Tool Execution (if allowed)
		var toolResp tools.ToolResponse
		if result.Decision == policy.DecisionAllow || result.Decision == policy.DecisionRedactAndAllow {
			toolReq := tools.ToolRequest{
				Name:   req.Tool,
				Amount: req.Amount,
			}
			toolResp, err = executor.Execute(r.Context(), toolReq)
			if err != nil {
				logEvent(r.Context(), logger, req, string(result.Decision), fmt.Sprintf("tool execution failed: %v", err))
				writeJSONError(w, http.StatusInternalServerError, "Tool execution failed")
				return
			}
		}

		// Audit Event
		logEvent(r.Context(), logger, req, string(result.Decision), result.Reason)

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

func logEvent(ctx context.Context, logger audit.Logger, req ExecuteRequest, decision, reason string) {
	logger.Log(ctx, audit.Event{
		Timestamp: time.Now(),
		Actor:     req.Actor,
		Action:    req.Tool,
		Amount:    req.Amount,
		Decision:  decision,
		Reason:    reason,
		PolicyVer: "v1.0.0-hardcoded",
	})
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}
