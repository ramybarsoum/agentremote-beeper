package openclaw

import (
	"testing"

	"maunium.net/go/mautrix/event"
)

func TestParseOpenClawControlCommand(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    *openClawControlCommand
		wantOK  bool
		msgType event.MessageType
		evtType event.Type
	}{
		{
			name:   "reset",
			body:   "/reset",
			want:   &openClawControlCommand{Action: "reset"},
			wantOK: true,
		},
		{
			name:   "rename",
			body:   "/rename Support Inbox",
			want:   &openClawControlCommand{Action: "label", Value: "Support Inbox"},
			wantOK: true,
		},
		{
			name:   "clear label",
			body:   "/label clear",
			want:   &openClawControlCommand{Action: "label", Clear: true},
			wantOK: true,
		},
		{
			name:   "thinking value",
			body:   "/thinking high",
			want:   &openClawControlCommand{Action: "thinking", Value: "high"},
			wantOK: true,
		},
		{
			name:   "reasoning clear",
			body:   "/reasoning default",
			want:   &openClawControlCommand{Action: "reasoning", Clear: true},
			wantOK: true,
		},
		{
			name:   "non command",
			body:   "hello",
			wantOK: false,
		},
		{
			name:    "media ignored",
			body:    "/reset",
			msgType: event.MsgImage,
			wantOK:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseOpenClawControlCommand(tt.body, tt.msgType, tt.evtType)
			if ok != tt.wantOK {
				t.Fatalf("unexpected ok: got %v want %v", ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if got == nil {
				t.Fatal("expected command")
			}
			if *got != *tt.want {
				t.Fatalf("unexpected command: got %+v want %+v", *got, *tt.want)
			}
		})
	}
}
