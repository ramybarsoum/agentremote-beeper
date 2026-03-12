package opencode

import (
	"testing"
	"time"
)

func TestCurrentCanonicalUIMessageFallbackIncludesModelAndUsage(t *testing.T) {
	oc := &OpenCodeClient{}
	ui := oc.currentCanonicalUIMessage(&openCodeStreamState{
		turnID:           "turn-1",
		agentID:          "agent-1",
		modelID:          "gpt-4.1",
		finishReason:     "stop",
		promptTokens:     11,
		completionTokens: 7,
		reasoningTokens:  3,
		totalTokens:      21,
		startedAtMs:      1000,
		completedAtMs:    2000,
	})

	metadata, ok := ui["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("expected metadata map, got %T", ui["metadata"])
	}
	if metadata["model"] != "gpt-4.1" {
		t.Fatalf("expected model metadata, got %#v", metadata["model"])
	}
	usage, ok := metadata["usage"].(map[string]any)
	if !ok {
		t.Fatalf("expected usage metadata, got %T", metadata["usage"])
	}
	if usage["total_tokens"] != int64(21) {
		t.Fatalf("expected total_tokens 21, got %#v", usage["total_tokens"])
	}
}

func TestOpenCodeStreamEventTimestampPrefersStartedAndCompleted(t *testing.T) {
	state := &openCodeStreamState{
		startedAtMs:   time.Date(2026, time.March, 12, 11, 0, 0, 0, time.UTC).UnixMilli(),
		completedAtMs: time.Date(2026, time.March, 12, 11, 0, 7, 0, time.UTC).UnixMilli(),
	}
	if got := openCodeStreamEventTimestamp(state, false); got.UnixMilli() != state.startedAtMs {
		t.Fatalf("expected startedAtMs timestamp, got %d", got.UnixMilli())
	}
	if got := openCodeStreamEventTimestamp(state, true); got.UnixMilli() != state.completedAtMs {
		t.Fatalf("expected completedAtMs timestamp, got %d", got.UnixMilli())
	}
}

func TestOpenCodeNextStreamOrderMonotonic(t *testing.T) {
	state := &openCodeStreamState{}
	ts := time.UnixMilli(1_700_000_000_000)
	first := openCodeNextStreamOrder(state, ts)
	second := openCodeNextStreamOrder(state, ts)
	if second <= first {
		t.Fatalf("expected monotonic stream order, got %d then %d", first, second)
	}
}
