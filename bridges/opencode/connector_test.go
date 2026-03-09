package opencode

import (
	"testing"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/event"
)

func TestFillPortalBridgeInfoSetsAIRoomType(t *testing.T) {
	conn := &OpenCodeConnector{}
	portal := &bridgev2.Portal{Portal: &database.Portal{}}
	meta := portalMeta(portal)
	meta.IsOpenCodeRoom = true

	content := &event.BridgeEventContent{}
	conn.FillPortalBridgeInfo(portal, content)
	if content.BeeperRoomTypeV2 != "ai" {
		t.Fatalf("expected ai room type, got %q", content.BeeperRoomTypeV2)
	}

	meta.IsOpenCodeRoom = false
	conn.FillPortalBridgeInfo(portal, content)
	if content.BeeperRoomTypeV2 != "ai" {
		t.Fatalf("expected ai room type for non-opencode room, got %q", content.BeeperRoomTypeV2)
	}
}
