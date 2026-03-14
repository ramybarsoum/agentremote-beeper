package ai

import (
	"context"
	"testing"

	"github.com/beeper/agentremote"
	bridgesdk "github.com/beeper/agentremote/sdk"
)

func TestRequestTurnApprovalWithoutApprovalFlowReturnsHandle(t *testing.T) {
	oc := &AIClient{}

	handle := oc.requestTurnApproval(context.Background(), nil, nil, nil, bridgesdk.ApprovalRequest{
		ApprovalID:   "approval-1",
		ToolCallID:   "tool-call-1",
		ToolName:     "tool",
		TTL:          60,
		Presentation: &agentremote.ApprovalPromptPresentation{Title: "Prompt"},
	})
	if handle == nil {
		t.Fatal("expected approval handle")
	}
	if handle.ID() != "approval-1" {
		t.Fatalf("expected approval id to round-trip, got %q", handle.ID())
	}
	if handle.ToolCallID() != "tool-call-1" {
		t.Fatalf("expected tool call id to round-trip, got %q", handle.ToolCallID())
	}

	resp, err := handle.Wait(context.Background())
	if err != nil {
		t.Fatalf("unexpected wait error: %v", err)
	}
	if resp.Approved {
		t.Fatal("expected approval to be denied without an approval flow")
	}
	if resp.Reason != agentremote.ApprovalReasonTimeout {
		t.Fatalf("expected timeout reason without approval flow, got %q", resp.Reason)
	}
}
