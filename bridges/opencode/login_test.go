package opencode

import "testing"

func TestGetLoginFlowsIncludesRemoteAndManaged(t *testing.T) {
	connector := &OpenCodeConnector{}
	flows := connector.GetLoginFlows()
	if len(flows) != 2 {
		t.Fatalf("expected 2 login flows, got %d", len(flows))
	}
	if flows[0].ID != FlowOpenCodeRemote {
		t.Fatalf("expected first flow to be remote, got %q", flows[0].ID)
	}
	if flows[1].ID != FlowOpenCodeManaged {
		t.Fatalf("expected second flow to be managed, got %q", flows[1].ID)
	}
}
