package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/CodeWithRJ006/warden/audit"
	"github.com/CodeWithRJ006/warden/policy"
	"github.com/CodeWithRJ006/warden/pii"
)

type ExecuteRequest struct {
	Actor  string  `json:"actor"`
	Tool   string  `json:"tool"`
	Amount float64 `json:"amount,omitempty"`
}

type ExecuteResponse struct {
	Decision policy.Decision `json:"decision"`
	Reason   string          `json:"reason"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// HandleExecute returns the HTTP handler for /v1/tools/execute.
func HandleExecute(engine policy.Engine, piiProc pii.Processor, logger audit.Logger) http.HandlerFunc {
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

		// PII Processing (Mocked for Day 2)
		// Assuming we will process some payload arguments in the future.
		_, err := piiProc.Process(r.Context(), req.Tool)
		if err != nil {
			logEvent(r.Context(), logger, req, "ERROR", "PII processing failed")
			writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		// Policy Evaluation
		policyReq := policy.RequestCtx{
			Role:   req.Actor,
			Action: req.Tool,
			Amount: req.Amount,
		}

		result, err := engine.Evaluate(r.Context(), policyReq)
		if err != nil {
			// Write audit event even on error
			logEvent(r.Context(), logger, req, "ERROR", err.Error())
			writeJSONError(w, http.StatusInternalServerError, "Policy evaluation failed")
			return
		}

		// Audit Event
		logEvent(r.Context(), logger, req, string(result.Decision), result.Reason)

		// Response
		resp := ExecuteResponse{
			Decision: result.Decision,
			Reason:   result.Reason,
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
