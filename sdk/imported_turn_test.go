package sdk

import (
	"testing"
	"time"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/networkid"
)

func TestConvertTranscriptTurnPreservesSenderAndToolMetadata(t *testing.T) {
	ts := time.Unix(1700000000, 0).UTC()
	sender := bridgeSender("agent")
	msg := ConvertTranscriptTurn(&ImportedTurn{
		ID:        "turn-1",
		Role:      "assistant",
		Text:      "hello",
		Reasoning: "thinking",
		Sender:    sender,
		Timestamp: ts,
		ToolCalls: []ImportedToolCall{{
			ID:     "call-1",
			Name:   "search",
			Input:  `{"q":"hello"}`,
			Output: `{"ok":true}`,
		}},
	}, "sdk")
	if msg == nil {
		t.Fatal("expected backfill message")
	}
	if msg.ID != networkid.MessageID("turn-1") {
		t.Fatalf("unexpected message id %q", msg.ID)
	}
	if msg.Sender != sender {
		t.Fatalf("unexpected sender: %#v", msg.Sender)
	}
	if !msg.Timestamp.Equal(ts) {
		t.Fatalf("unexpected timestamp: %v", msg.Timestamp)
	}
}

func bridgeSender(id string) bridgev2.EventSender {
	return bridgev2.EventSender{
		Sender:      networkid.UserID(id),
		SenderLogin: networkid.UserLoginID("login-1"),
	}
}
