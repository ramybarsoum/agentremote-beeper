package ai

import (
	"context"
	"strings"

	"github.com/openai/openai-go/v3"
	"maunium.net/go/mautrix/id"

	airuntime "github.com/beeper/agentremote/pkg/runtime"
)

func (a *agentLoopProviderBase) GetFollowUpMessages(context.Context) []openai.ChatCompletionMessageParamUnion {
	if a == nil || a.oc == nil || a.state == nil {
		return nil
	}
	return a.oc.getFollowUpMessages(a.state.roomID)
}

func (a *agentLoopProviderBase) ContinueAgentLoop(messages []openai.ChatCompletionMessageParamUnion) {
	if a == nil || len(messages) == 0 {
		return
	}
	a.messages = append(a.messages, messages...)
}

func (oc *AIClient) getFollowUpMessages(roomID id.RoomID) []openai.ChatCompletionMessageParamUnion {
	prompts, items := oc.takeAgentLoopFollowUpPrompts(roomID)
	if len(prompts) == 0 {
		return nil
	}
	for _, item := range items {
		oc.registerRoomRunPendingItem(roomID, item)
	}
	return buildSteeringUserMessages(prompts)
}

func (oc *AIClient) takeAgentLoopFollowUpPrompts(roomID id.RoomID) ([]string, []pendingQueueItem) {
	if oc == nil || roomID == "" {
		return nil, nil
	}
	candidate, snapshot := oc.takePendingQueueDispatchCandidate(roomID, true)
	if snapshot == nil {
		return nil, nil
	}
	behavior := airuntime.ResolveQueueBehavior(snapshot.mode)
	if !behavior.Followup {
		return nil, nil
	}
	if candidate == nil || len(candidate.items) == 0 {
		return nil, nil
	}
	if candidate.collect {
		for idx := range candidate.items {
			candidate.items[idx].prompt = strings.TrimSpace(candidate.items[idx].pending.MessageBody)
		}
		return []string{buildCollectPrompt("[Queued messages while agent was busy]", candidate.items, candidate.summaryPrompt)}, candidate.items
	}
	if candidate.summaryPrompt != "" && candidate.synthetic {
		return []string{candidate.summaryPrompt}, candidate.items
	}
	return []string{strings.TrimSpace(candidate.items[0].pending.MessageBody)}, candidate.items
}
