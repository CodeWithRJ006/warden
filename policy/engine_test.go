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
			name:         "finance-manager get payment",
			req:          RequestCtx{Role: "finance-manager", Action: "get_payment"},
			wantDecision: DecisionAllow,
		},
		{
			name:         "finance-manager refund under limit",
			req:          RequestCtx{Role: "finance-manager", Action: "create_refund", Amount: 50000.0},
			wantDecision: DecisionAllow,
		},
		{
			name:         "finance-manager refund over limit",
			req:          RequestCtx{Role: "finance-manager", Action: "create_refund", Amount: 150000.0},
			wantDecision: DecisionRequireApproval,
		},
		{
			name:         "finance-operator refund under limit",
			req:          RequestCtx{Role: "finance-operator", Action: "create_refund", Amount: 5000.0},
			wantDecision: DecisionAllow,
		},
		{
			name:         "finance-operator refund over limit",
			req:          RequestCtx{Role: "finance-operator", Action: "create_refund", Amount: 50000.0},
			wantDecision: DecisionRequireApproval,
		},
		{
			name:         "finance-operator fetch settlement denied",
			req:          RequestCtx{Role: "finance-operator", Action: "fetch_settlement"},
			wantDecision: DecisionDeny,
		},
		{
			name:         "support agent get payment",
			req:          RequestCtx{Role: "support-agent", Action: "get_payment"},
			wantDecision: DecisionAllow,
		},
		{
			name:         "support agent fetch settlement redaction",
			req:          RequestCtx{Role: "support-agent", Action: "fetch_settlement"},
			wantDecision: DecisionRedactAndAllow,
		},
		{
			name:         "support agent refund denied",
			req:          RequestCtx{Role: "support-agent", Action: "create_refund", Amount: 50.0},
			wantDecision: DecisionDeny,
		},
		{
			name:         "viewer get payment redaction",
			req:          RequestCtx{Role: "viewer", Action: "get_payment"},
			wantDecision: DecisionRedactAndAllow,
		},
		{
			name:         "viewer mutate action",
			req:          RequestCtx{Role: "viewer", Action: "create_refund"},
			wantDecision: DecisionDeny,
		},
		{
			name:         "missing fields",
			req:          RequestCtx{Role: "finance-operator"},
			wantDecision: "",
			wantErr:      true,
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
