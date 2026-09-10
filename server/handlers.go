package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	Actor  string  `json:"-"`
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

type ModelRouter interface {
	Route(ctx context.Context, dataClass string) (string, error)
}

type HttpModelRouter struct {
	LocalURL    string
	ExternalURL string
}

func (r *HttpModelRouter) Route(ctx context.Context, dataClass string) (string, error) {
	client := http.Client{Timeout: 1 * time.Second}
	
	// Always try Local Model first
	_, err := client.Get(r.LocalURL)
	if err == nil {
		return "local_model", nil
	}

	// Local is down. If restricted, FAIL CLOSED.
	if dataClass == "restricted" {
		return "", fmt.Errorf("local model unavailable and data is restricted")
	}

	// Clean data. Try External Model fallback.
	_, err = client.Get(r.ExternalURL)
	if err == nil {
		return "external_model", nil
	}

	return "", fmt.Errorf("all models unavailable")
}

// HandleExecute returns the HTTP handler for /v1/tools/execute.
func HandleExecute(engine policy.Engine, piiProc pii.Processor, executor tools.Executor, logger audit.Logger, limiter middleware.RateLimiter, router ModelRouter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := generateRequestID()

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		actor := r.Header.Get("X-Warden-Actor")
		if actor == "" {
			writeJSONError(w, http.StatusUnauthorized, "Missing X-Warden-Actor header")
			return
		}

		var req ExecuteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}
		req.Actor = actor

		// Validation
		if req.Actor == "" || req.Tool == "" {
			writeJSONError(w, http.StatusBadRequest, "Missing required fields: actor and tool")
			return
		}

		// Rate Limiting
		allowed, err := limiter.Allow(r.Context(), req.Actor, 10, time.Minute)
		if err != nil || !allowed {
			logEvent(r.Context(), logger, reqID, req.Actor, req.Tool, "clean", audit.PolicyTrace{Decision: "DENY", Rule: "rate_limit"}, audit.ModelTrace{}, audit.ToolTrace{}, "RATE_LIMITED")
			writeJSONError(w, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		}

		// PII Processing
		sanitizedTool, err := piiProc.Process(r.Context(), req.Tool)
		if err != nil {
			logEvent(r.Context(), logger, reqID, req.Actor, req.Tool, "clean", audit.PolicyTrace{Decision: "ERROR", Rule: "pii_processing"}, audit.ModelTrace{}, audit.ToolTrace{}, "ERROR")
			writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		// Data Classification
		dataClass := "clean"
		if strings.Contains(sanitizedTool, "[") {
			dataClass = "restricted"
		}

		// Model Routing
		modelName, err := router.Route(r.Context(), dataClass)
		if err != nil {
			logEvent(r.Context(), logger, reqID, req.Actor, sanitizedTool, dataClass, audit.PolicyTrace{Decision: "DENY", Rule: "model_routing_failed"}, audit.ModelTrace{Invoked: false}, audit.ToolTrace{}, "DENY_MODEL_UNAVAILABLE")
			writeJSONError(w, http.StatusServiceUnavailable, "Model routing failed: "+err.Error())
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
			logEvent(r.Context(), logger, reqID, req.Actor, sanitizedTool, dataClass, audit.PolicyTrace{Decision: "ERROR", Rule: "policy_engine"}, audit.ModelTrace{Invoked: true, ModelName: modelName}, audit.ToolTrace{}, "ERROR")
			writeJSONError(w, http.StatusInternalServerError, "Policy evaluation failed")
			return
		}

		// Tool Execution
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
				logEvent(r.Context(), logger, reqID, req.Actor, sanitizedTool, dataClass, audit.PolicyTrace{Decision: string(result.Decision), Rule: result.Reason}, audit.ModelTrace{Invoked: true, ModelName: modelName}, audit.ToolTrace{Executed: true, Status: "FAILED"}, "ERROR")
				writeJSONError(w, http.StatusInternalServerError, "Tool execution failed")
				return
			}
		}

		// Audit Event
		logEvent(r.Context(), logger, reqID, req.Actor, sanitizedTool, dataClass, audit.PolicyTrace{Decision: string(result.Decision), Rule: result.Reason}, audit.ModelTrace{Invoked: true, ModelName: modelName}, audit.ToolTrace{Executed: toolExecuted, Status: toolResp.Status}, string(result.Decision))

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

func logEvent(ctx context.Context, logger audit.Logger, reqID string, actor string, tool string, dataClass string, policyTrace audit.PolicyTrace, modelTrace audit.ModelTrace, toolTrace audit.ToolTrace, finalDecision string) {
	event := audit.Event{
		RequestID:          reqID,
		Actor:              actor,
		Tool:               tool,
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

func HandleGetLogs(logger audit.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		events, err := logger.GetLogs(r.Context(), 50)
		if err != nil {
			fmt.Println("Error in GetLogs:", err)
			writeJSONError(w, 500, "Failed to get logs: " + err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(events)
	}
}

