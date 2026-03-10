package openclaw

import (
	"testing"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/id"

	"github.com/beeper/agentremote/pkg/shared/streamui"
)

func TestApplyStreamPlaceholderResultWithoutEventIDFallsBackToDebounced(t *testing.T) {
	oc := &OpenClawClient{
		streamStates: map[string]*openClawStreamState{
			"turn-1": {turnID: "turn-1", placeholderPending: true},
		},
	}

	msgID := networkid.MessageID("openclaw:msg-1")
	oc.applyStreamPlaceholderResult("turn-1", msgID, bridgev2.EventHandlingResult{Success: true})

	state := oc.streamStates["turn-1"]
	if state == nil {
		t.Fatal("expected stream state")
	}
	if state.placeholderPending {
		t.Fatal("expected placeholderPending to be cleared")
	}
	if state.networkMessageID != msgID {
		t.Fatalf("expected network message id %q, got %q", msgID, state.networkMessageID)
	}
	if state.initialEventID != "" {
		t.Fatalf("expected empty initial event id, got %q", state.initialEventID)
	}
	if state.targetEventID != "" {
		t.Fatalf("expected empty target event id, got %q", state.targetEventID)
	}
	if !state.streamFallbackToDebounced.Load() {
		t.Fatal("expected stream to fall back to debounced edits without an event id")
	}
}

func TestApplyStreamPlaceholderResultWithEventIDKeepsEphemeralStreaming(t *testing.T) {
	oc := &OpenClawClient{
		streamStates: map[string]*openClawStreamState{
			"turn-2": {turnID: "turn-2", placeholderPending: true},
		},
	}

	msgID := networkid.MessageID("openclaw:msg-2")
	eventID := id.EventID("$event-2")
	oc.applyStreamPlaceholderResult("turn-2", msgID, bridgev2.EventHandlingResult{
		Success: true,
		EventID: eventID,
	})

	state := oc.streamStates["turn-2"]
	if state == nil {
		t.Fatal("expected stream state")
	}
	if state.placeholderPending {
		t.Fatal("expected placeholderPending to be cleared")
	}
	if state.networkMessageID != msgID {
		t.Fatalf("expected network message id %q, got %q", msgID, state.networkMessageID)
	}
	if state.initialEventID != eventID {
		t.Fatalf("expected initial event id %q, got %q", eventID, state.initialEventID)
	}
	if state.targetEventID != eventID.String() {
		t.Fatalf("expected target event id %q, got %q", eventID.String(), state.targetEventID)
	}
	if state.streamFallbackToDebounced.Load() {
		t.Fatal("expected ephemeral streaming to remain enabled")
	}
}

func TestApplyStreamPlaceholderResultFailureAllowsRetry(t *testing.T) {
	oc := &OpenClawClient{
		streamStates: map[string]*openClawStreamState{
			"turn-3": {turnID: "turn-3", placeholderPending: true},
		},
	}

	oc.applyStreamPlaceholderResult("turn-3", networkid.MessageID("openclaw:msg-3"), bridgev2.EventHandlingResult{})

	state := oc.streamStates["turn-3"]
	if state == nil {
		t.Fatal("expected stream state")
	}
	if state.placeholderPending {
		t.Fatal("expected placeholderPending to be cleared after failure")
	}
	if state.networkMessageID != "" {
		t.Fatalf("expected network message id to remain empty, got %q", state.networkMessageID)
	}
	if state.streamFallbackToDebounced.Load() {
		t.Fatal("expected no fallback when placeholder send fails")
	}
}

func TestBuildStreamDBMetadataIncludesToolCalls(t *testing.T) {
	oc := &OpenClawClient{}
	state := &openClawStreamState{
		turnID:     "turn-4",
		agentID:    "main",
		sessionID:  "sess-1",
		sessionKey: "agent:main:matrix-dm",
		role:       "assistant",
	}
	state.visible.WriteString("running")
	streamui.ApplyChunk(&state.ui, map[string]any{
		"type": "reasoning-start",
		"id":   "reasoning-1",
	})
	streamui.ApplyChunk(&state.ui, map[string]any{
		"type":  "reasoning-delta",
		"id":    "reasoning-1",
		"delta": "thinking",
	})
	streamui.ApplyChunk(&state.ui, map[string]any{
		"type": "reasoning-end",
		"id":   "reasoning-1",
	})
	streamui.ApplyChunk(&state.ui, map[string]any{
		"type":       "tool-input-available",
		"toolCallId": "call-1",
		"toolName":   "bash",
		"input":      map[string]any{"cmd": "pwd"},
	})
	streamui.ApplyChunk(&state.ui, map[string]any{
		"type":       "tool-output-available",
		"toolCallId": "call-1",
		"output":     map[string]any{"stdout": "/tmp"},
	})
	streamui.ApplyChunk(&state.ui, map[string]any{
		"type":      "file",
		"url":       "mxc://example.org/out",
		"mediaType": "image/png",
	})

	meta := oc.buildStreamDBMetadata(state)
	if meta == nil {
		t.Fatal("expected metadata")
	}
	if meta.ThinkingContent != "thinking" {
		t.Fatalf("unexpected thinking content: %q", meta.ThinkingContent)
	}
	if len(meta.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %#v", meta.ToolCalls)
	}
	call := meta.ToolCalls[0]
	if call.CallID != "call-1" || call.ToolName != "bash" || call.ToolType != "openclaw" {
		t.Fatalf("unexpected tool call metadata: %#v", call)
	}
	if call.Status != "output-available" || call.ResultStatus != "completed" {
		t.Fatalf("unexpected tool call status: %#v", call)
	}
	if call.Input["cmd"] != "pwd" {
		t.Fatalf("unexpected tool input: %#v", call.Input)
	}
	if call.Output["stdout"] != "/tmp" {
		t.Fatalf("unexpected tool output: %#v", call.Output)
	}
	if len(meta.GeneratedFiles) != 1 {
		t.Fatalf("expected 1 generated file, got %#v", meta.GeneratedFiles)
	}
	if meta.GeneratedFiles[0].URL != "mxc://example.org/out" || meta.GeneratedFiles[0].MimeType != "image/png" {
		t.Fatalf("unexpected generated files: %#v", meta.GeneratedFiles)
	}
}
