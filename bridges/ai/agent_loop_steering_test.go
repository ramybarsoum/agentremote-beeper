package ai

import (
	"context"
	"testing"

	"github.com/openai/openai-go/v3"
	"maunium.net/go/mautrix/id"

	airuntime "github.com/beeper/agentremote/pkg/runtime"
)

func TestGetSteeringMessages_FiltersAndDrainsQueue(t *testing.T) {
	roomID := id.RoomID("!room:example.com")
	oc := &AIClient{
		connector: &OpenAIConnector{},
		activeRoomRuns: map[id.RoomID]*roomRunState{
			roomID: {
				steerQueue: []pendingQueueItem{
					{
						pending: pendingMessage{Type: pendingTypeText, MessageBody: "fallback"},
						prompt:  "  explicit steer  ",
					},
					{
						pending: pendingMessage{Type: pendingTypeText, MessageBody: "body only"},
					},
					{
						pending: pendingMessage{Type: pendingTypeImage, MessageBody: "ignored"},
						prompt:  "ignored",
					},
					{
						pending: pendingMessage{Type: pendingTypeText, MessageBody: "   "},
						prompt:  "   ",
					},
				},
			},
		},
	}

	got := oc.getSteeringMessages(roomID)
	if len(got) != 2 {
		t.Fatalf("expected 2 steering messages, got %d: %#v", len(got), got)
	}
	if got[0] != "explicit steer" {
		t.Fatalf("expected first steering prompt to prefer explicit prompt, got %q", got[0])
	}
	if got[1] != "body only" {
		t.Fatalf("expected second steering prompt to fallback to message body, got %q", got[1])
	}

	if again := oc.getSteeringMessages(roomID); len(again) != 0 {
		t.Fatalf("expected steering queue to be drained, got %#v", again)
	}
}

func TestBuildChatAgentLoopContinuationMessages_OrdersAssistantToolResultsAndSteering(t *testing.T) {
	roomID := id.RoomID("!room:example.com")
	oc := &AIClient{
		connector: &OpenAIConnector{},
		activeRoomRuns: map[id.RoomID]*roomRunState{
			roomID: {
				steerQueue: []pendingQueueItem{
					{
						pending: pendingMessage{Type: pendingTypeText, MessageBody: "steer now"},
					},
				},
			},
		},
	}
	state := &streamingState{
		roomID: roomID,
		pendingFunctionOutputs: []functionCallOutput{{
			callID: "call_1",
			output: "tool output",
		}},
	}

	got := oc.buildChatAgentLoopContinuationMessages(
		state,
		[]openai.ChatCompletionMessageParamUnion{openai.UserMessage("before")},
		openai.ChatCompletionAssistantMessageParam{},
		[]string{"steer now"},
	)

	if len(got) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(got))
	}
	if got[1].OfAssistant == nil {
		t.Fatalf("expected assistant continuation message at index 1")
	}
	if got[2].OfTool == nil || got[2].OfTool.ToolCallID != "call_1" {
		t.Fatalf("expected tool result message at index 2, got %#v", got[2])
	}
	if got[3].OfUser == nil || got[3].OfUser.Content.OfString.Value != "steer now" {
		t.Fatalf("expected steering user message at index 3, got %#v", got[3])
	}
}

func TestTakeAgentLoopFollowUpPrompts_ConsumesSingleQueuedTextMessage(t *testing.T) {
	roomID := id.RoomID("!room:example.com")
	oc := &AIClient{
		pendingQueues: map[id.RoomID]*pendingQueue{
			roomID: {
				mode: airuntime.QueueModeFollowup,
				items: []pendingQueueItem{
					{pending: pendingMessage{Type: pendingTypeText, MessageBody: "follow up"}},
				},
			},
		},
	}

	prompts, items := oc.takeAgentLoopFollowUpPrompts(roomID)
	if len(prompts) != 1 || prompts[0] != "follow up" {
		t.Fatalf("unexpected follow-up prompts: %#v", prompts)
	}
	if len(items) != 1 || items[0].pending.MessageBody != "follow up" {
		t.Fatalf("unexpected consumed follow-up items: %#v", items)
	}
	if snapshot := oc.getQueueSnapshot(roomID); snapshot != nil {
		t.Fatalf("expected queue to be drained, got %#v", snapshot.items)
	}
}

func TestTakeAgentLoopFollowUpPrompts_LeavesNonTextQueueItemsForBacklogProcessing(t *testing.T) {
	roomID := id.RoomID("!room:example.com")
	oc := &AIClient{
		pendingQueues: map[id.RoomID]*pendingQueue{
			roomID: {
				mode: airuntime.QueueModeFollowup,
				items: []pendingQueueItem{
					{pending: pendingMessage{Type: pendingTypeImage, MessageBody: "image"}},
				},
			},
		},
	}

	prompts, items := oc.takeAgentLoopFollowUpPrompts(roomID)
	if len(prompts) != 0 || len(items) != 0 {
		t.Fatalf("expected non-text follow-up to stay queued, got prompts=%#v items=%#v", prompts, items)
	}
	if snapshot := oc.getQueueSnapshot(roomID); snapshot == nil || len(snapshot.items) != 1 {
		t.Fatalf("expected non-text queue item to remain queued, got %#v", snapshot)
	}
}

func TestBuildContinuationParams_UsesPendingSteeringPromptsBeforeDrainingQueue(t *testing.T) {
	roomID := id.RoomID("!room:example.com")
	oc := &AIClient{
		connector: &OpenAIConnector{},
		activeRoomRuns: map[id.RoomID]*roomRunState{
			roomID: {
				steerQueue: []pendingQueueItem{
					{pending: pendingMessage{Type: pendingTypeText, MessageBody: "queue steer"}},
				},
			},
		},
	}
	state := &streamingState{roomID: roomID}
	state.addPendingSteeringPrompts([]string{"pending steer"})

	params := oc.buildContinuationParams(context.Background(), state, nil, nil, nil)
	if len(params.Input.OfInputItemList) == 0 {
		t.Fatal("expected continuation input to include stored steering prompt")
	}
	if pending := state.consumePendingSteeringPrompts(); len(pending) != 0 {
		t.Fatalf("expected pending steering prompts to be consumed, got %#v", pending)
	}
	if snapshot := oc.getRoomRun(roomID); snapshot == nil || len(snapshot.steerQueue) != 1 {
		t.Fatalf("expected queued steering item to remain available, got %#v", snapshot)
	}
}
