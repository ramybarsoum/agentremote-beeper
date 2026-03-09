package connector

import (
	"context"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/event"

	"github.com/beeper/ai-bridge/pkg/bridgeadapter"
)

const aiBridgeProtocolID = "ai"

func applyAIBridgeInfo(portal *bridgev2.Portal, meta *PortalMetadata, content *event.BridgeEventContent) {
	if portal == nil {
		return
	}
	bridgeadapter.ApplyAIBridgeInfo(content, aiBridgeProtocolID, portal.RoomType, integrationPortalAIKind(meta))
}

func sendAIPortalInfo(ctx context.Context, portal *bridgev2.Portal, meta *PortalMetadata) bool {
	return bridgeadapter.SendAIRoomInfo(ctx, portal, integrationPortalAIKind(meta))
}
