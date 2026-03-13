package agentremote

import "testing"

func TestNewApprovalFlowInit(t *testing.T) {
	flow := NewApprovalFlow[map[string]any](ApprovalFlowConfig[map[string]any]{})
	if flow == nil {
		t.Fatal("expected approval flow")
	}
}

func TestNewRuntimeInitializesServices(t *testing.T) {
	runtime := NewRuntime(RuntimeConfig{AgentID: " agent "})
	if runtime == nil {
		t.Fatal("expected runtime")
	}
	if runtime.AgentID != "agent" {
		t.Fatalf("expected trimmed agent id, got %q", runtime.AgentID)
	}
	if runtime.Turns == nil {
		t.Fatal("expected turn manager")
	}
	if runtime.Approvals == nil {
		t.Fatal("expected approval manager")
	}
}
