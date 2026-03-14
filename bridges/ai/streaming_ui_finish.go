package ai

import (
	"context"

	"maunium.net/go/mautrix/bridgev2"

	"github.com/beeper/agentremote/bridges/ai/msgconv"
	"github.com/beeper/agentremote/turns"
)

func (oc *AIClient) emitUIFinish(ctx context.Context, portal *bridgev2.Portal, state *streamingState, meta *PortalMetadata) {
	if state == nil {
		return
	}
	finishReason := msgconv.MapFinishReason(state.finishReason)
	state.writer().Finish(ctx, finishReason, oc.buildUIMessageMetadata(state, meta, true))
	if session := state.turn.Session(); session != nil {
		session.End(ctx, mapTurnEndReason(finishReason))
	}
}

func mapTurnEndReason(reason string) turns.EndReason {
	switch reason {
	case "error":
		return turns.EndReasonError
	case "disconnect":
		return turns.EndReasonDisconnect
	case "stop", "length", "content-filter", "tool-calls", "other":
		return turns.EndReasonFinish
	default:
		return turns.EndReasonFinish
	}
}
