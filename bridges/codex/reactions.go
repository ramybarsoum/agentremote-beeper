package codex

import (
	"context"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"

	"github.com/beeper/agentremote/pkg/bridgeadapter"
)

func (cc *CodexClient) PreHandleMatrixReaction(_ context.Context, msg *bridgev2.MatrixReaction) (bridgev2.MatrixReactionPreResponse, error) {
	return bridgeadapter.PreHandleApprovalReaction(msg)
}

func (cc *CodexClient) HandleMatrixReaction(ctx context.Context, msg *bridgev2.MatrixReaction) (*database.Reaction, error) {
	if cc == nil || msg == nil || msg.Event == nil || msg.Portal == nil {
		return &database.Reaction{}, nil
	}
	if bridgeadapter.IsMatrixBotUser(ctx, cc.UserLogin.Bridge, msg.Event.Sender) {
		return &database.Reaction{}, nil
	}
	rc := bridgeadapter.ExtractReactionContext(msg)
	cc.approvalPrompts.HandleReaction(ctx, msg, rc.TargetEventID, rc.Emoji)
	return &database.Reaction{}, nil
}

func (cc *CodexClient) HandleMatrixReactionRemove(_ context.Context, _ *bridgev2.MatrixReactionRemove) error {
	return nil
}
