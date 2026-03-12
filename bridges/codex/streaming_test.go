package codex

import (
	"context"
	"testing"
	"time"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/id"
)

func TestCodex_StreamChunks_BasicOrderingAndSeq(t *testing.T) {
	ctx := context.Background()

	var gotParts []map[string]any
	var gotSeq []int
	cc := &CodexClient{
		streamEventHook: func(turnID string, seq int, content map[string]any, txnID string) {
			_ = turnID
			_ = txnID
			gotSeq = append(gotSeq, seq)
			partAny, _ := content["part"].(map[string]any)
			gotParts = append(gotParts, partAny)
		},
	}
	portal := &bridgev2.Portal{Portal: &database.Portal{MXID: id.RoomID("!room:example.com")}}
	state := &streamingState{
		turnID:           "turn_local_1",
		initialEventID:   id.EventID("$event"),
		networkMessageID: networkid.MessageID("codex:test"),
	}

	cc.emitUIStart(ctx, portal, state, "gpt-5.1-codex")
	cc.uiEmitter(state).EmitUIStepStart(ctx, portal)
	cc.uiEmitter(state).EmitUITextDelta(ctx, portal, "hi")
	cc.emitUIFinish(ctx, portal, state, "gpt-5.1-codex", "completed")

	if len(gotParts) < 5 {
		t.Fatalf("expected >=5 parts, got %d", len(gotParts))
	}
	if gotSeq[0] != 1 {
		t.Fatalf("expected first seq=1, got %d", gotSeq[0])
	}
	for i := 1; i < len(gotSeq); i++ {
		if gotSeq[i] <= gotSeq[i-1] {
			t.Fatalf("seq not monotonic at %d: %v", i, gotSeq)
		}
	}

	if gotParts[0]["type"] != "start" {
		t.Fatalf("expected first part type=start, got %#v", gotParts[0]["type"])
	}
	if gotParts[1]["type"] != "start-step" {
		t.Fatalf("expected second part type=start-step, got %#v", gotParts[1]["type"])
	}
	// text-start then text-delta should be present before finish.
	seenTextStart := false
	seenTextDelta := false
	seenFinish := false
	for _, p := range gotParts {
		switch p["type"] {
		case "text-start":
			seenTextStart = true
		case "text-delta":
			seenTextDelta = true
		case "finish":
			seenFinish = true
		}
	}
	if !seenTextStart || !seenTextDelta {
		t.Fatalf("expected text-start and text-delta, got parts=%v", gotParts)
	}
	if !seenFinish {
		t.Fatalf("expected finish part, got parts=%v", gotParts)
	}
}

func TestCodexStreamEventTimestampPrefersStartedAndCompleted(t *testing.T) {
	state := &streamingState{
		startedAtMs:   time.Date(2026, time.March, 12, 10, 0, 0, 0, time.UTC).UnixMilli(),
		completedAtMs: time.Date(2026, time.March, 12, 10, 0, 5, 0, time.UTC).UnixMilli(),
	}
	if got := codexStreamEventTimestamp(state, false); got.UnixMilli() != state.startedAtMs {
		t.Fatalf("expected startedAtMs timestamp, got %d", got.UnixMilli())
	}
	if got := codexStreamEventTimestamp(state, true); got.UnixMilli() != state.completedAtMs {
		t.Fatalf("expected completedAtMs timestamp, got %d", got.UnixMilli())
	}
}

func TestCodexNextLiveStreamOrderMonotonic(t *testing.T) {
	state := &streamingState{}
	ts := time.UnixMilli(1_700_000_000_000)
	first := codexNextLiveStreamOrder(state, ts)
	second := codexNextLiveStreamOrder(state, ts)
	if second <= first {
		t.Fatalf("expected monotonic stream order, got %d then %d", first, second)
	}
}
