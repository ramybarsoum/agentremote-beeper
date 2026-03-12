package ai

import (
	"context"
	"testing"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/id"

	"github.com/beeper/agentremote"
)

func TestEmitUIToolApprovalRequestWithoutApprovalFlow(t *testing.T) {
	owner := id.UserID("@owner:example.com")
	portal := &bridgev2.Portal{Portal: &database.Portal{MXID: id.RoomID("!room:example.com")}}
	oc := &AIClient{
		UserLogin: &bridgev2.UserLogin{
			UserLogin: &database.UserLogin{
				UserMXID: owner,
			},
		},
	}

	ok := oc.emitUIToolApprovalRequest(
		context.Background(),
		portal,
		nil,
		"approval-1",
		"tool-call-1",
		"tool",
		agentremote.ApprovalPromptPresentation{Title: "Prompt"},
		"",
		60,
	)
	if ok {
		t.Fatalf("expected approval prompt emission to fail without an approval flow")
	}
}
