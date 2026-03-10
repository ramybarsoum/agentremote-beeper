package opencode

import (
	"context"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"

	"github.com/beeper/agentremote/pkg/bridgeadapter"
)

func (oc *OpenCodeClient) PreHandleMatrixReaction(_ context.Context, msg *bridgev2.MatrixReaction) (bridgev2.MatrixReactionPreResponse, error) {
	return bridgeadapter.PreHandleApprovalReaction(msg)
}

func (oc *OpenCodeClient) HandleMatrixReaction(ctx context.Context, msg *bridgev2.MatrixReaction) (*database.Reaction, error) {
	if oc == nil || msg == nil || msg.Event == nil || oc.bridge == nil {
		return &database.Reaction{}, nil
	}
	if bridgeadapter.IsMatrixBotUser(ctx, oc.UserLogin.Bridge, msg.Event.Sender) {
		return &database.Reaction{}, nil
	}
	rc := bridgeadapter.ExtractReactionContext(msg)
	oc.bridge.HandleApprovalPromptReaction(ctx, msg, rc.TargetEventID, rc.Emoji)
	return &database.Reaction{}, nil
}

func (oc *OpenCodeClient) HandleMatrixReactionRemove(_ context.Context, _ *bridgev2.MatrixReactionRemove) error {
	return nil
}
