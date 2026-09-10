package policy

import (
	"context"
	"testing"
)

func TestHardcodedEngine_Evaluate(t *testing.T) {
	engine := NewHardcodedEngine()
	ctx := context.Background()

	tests := []struct {
		name         string
		req          RequestCtx
		wantDecision Decision
		wantErr      bool
	}{
		{
			name:         "admin bypass",
			req:          RequestCtx{Role: "admin", Action: "capture_payment"},
			wantDecision: DecisionAllow,
		},
		{
			name:         "support agent get payment",
			req:          RequestCtx{Role: "support-agent", Action: "get_payment"},
			wantDecision: DecisionAllow,
		},
		{
			name:         "support agent refund under limit",
			req:          RequestCtx{Role: "support-agent", Action: "create_refund", Amount: 50.0},
			wantDecision: DecisionAllow,
		},
		{
			name:         "support agent refund over limit",
			req:          RequestCtx{Role: "support-agent", Action: "create_refund", Amount: 150.0},
			wantDecision: DecisionRequireApproval,
		},
		{
			name:         "support agent unauthorized action",
			req:          RequestCtx{Role: "support-agent", Action: "capture_payment"},
			wantDecision: DecisionDeny,
		},
		{
			name:         "viewer get payment",
			req:          RequestCtx{Role: "viewer", Action: "get_payment"},
			wantDecision: DecisionAllow,
		},
		{
			name:         "viewer mutate action",
			req:          RequestCtx{Role: "viewer", Action: "create_refund"},
			wantDecision: DecisionDeny,
		},
		{
			name:         "missing fields",
			req:          RequestCtx{Role: "admin"},
			wantDecision: "",
			wantErr:      true,
		},
		{
			name:         "support agent fetch settlement redaction",
			req:          RequestCtx{Role: "support-agent", Action: "fetch_settlement"},
			wantDecision: DecisionRedactAndAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Evaluate(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("HardcodedEngine.Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Decision != tt.wantDecision {
				t.Errorf("HardcodedEngine.Evaluate() = %v, want %v", got.Decision, tt.wantDecision)
			}
		})
	}
}
