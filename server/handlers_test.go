package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CodeWithRJ006/warden/audit"
	"github.com/CodeWithRJ006/warden/middleware"
	"github.com/CodeWithRJ006/warden/pii"
	"github.com/CodeWithRJ006/warden/policy"
	"github.com/CodeWithRJ006/warden/tools"
)

type mockLogger struct {
	events []audit.Event
}

type MockRouter struct{}

func (m *mockLogger) Log(ctx context.Context, event audit.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockLogger) GetLogs(ctx context.Context, limit int) ([]audit.Event, error) {
	return m.events, nil
}

func TestHandleExecute(t *testing.T) {
	engine := policy.NewHardcodedEngine()
	vault := pii.NewMemoryVault()
	piiProc := pii.NewTokenizer(vault, "")
	logger := &mockLogger{}
	executor := tools.NewMockExecutor()
	limiter := middleware.NewMemoryRateLimiter()
	handler := HandleExecute(engine, piiProc, executor, logger, limiter, &MockRouter{})

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		expectedDec    policy.Decision
	}{
		{
			name: "valid request - allowed",
			body: map[string]interface{}{
				"actor":  "finance-operator",
				"tool":   "create_refund",
				"amount": 50.0,
			},
			expectedStatus: http.StatusOK,
			expectedDec:    policy.DecisionAllow,
		},
		{
			name: "valid request - require approval",
			body: map[string]interface{}{
				"actor":  "finance-operator",
				"tool":   "create_refund",
				"amount": 50000.0,
			},
			expectedStatus: http.StatusOK,
			expectedDec:    policy.DecisionRequireApproval,
		},
		{
			name: "missing actor",
			body: map[string]interface{}{
				"tool": "create_refund",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "malformed JSON",
			body: nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			} else {
				bodyBytes = []byte("{invalid json")
			}

			req, _ := http.NewRequest(http.MethodPost, "/v1/tools/execute", bytes.NewBuffer(bodyBytes))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp ExecuteResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.Decision != tt.expectedDec {
					t.Errorf("handler returned wrong decision: got %v want %v", resp.Decision, tt.expectedDec)
				}
			}
		})
	}

	// Verify that audit logs were generated for valid requests
	if len(logger.events) == 0 {
		t.Errorf("expected audit events to be generated, got 0")
	}
}


func (m *MockRouter) Route(ctx context.Context, dataClass string) (string, error) { return "local_model", nil }

