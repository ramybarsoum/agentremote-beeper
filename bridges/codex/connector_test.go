package codex

import (
	"testing"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/event"
)

func TestFillPortalBridgeInfoSetsAIRoomType(t *testing.T) {
	conn := &CodexConnector{}
	content := &event.BridgeEventContent{}

	conn.FillPortalBridgeInfo(&bridgev2.Portal{}, content)
	if content.BeeperRoomTypeV2 != "ai" {
		t.Fatalf("expected ai room type, got %q", content.BeeperRoomTypeV2)
	}
}
