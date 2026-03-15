package ai

import (
	"context"
	"strings"
	"testing"

	bridgesdk "github.com/beeper/agentremote/sdk"
)

func TestGenerateStreamRejectsUnsupportedResponsesPromptContext(t *testing.T) {
	provider := &OpenAIProvider{}
	params := GenerateParams{
		Context: PromptContext{
			PromptContext: bridgesdk.UserPromptContext(bridgesdk.PromptBlock{
				Type:        bridgesdk.PromptBlockAudio,
				AudioB64:    "YXVkaW8=",
				AudioFormat: "mp3",
				MimeType:    "audio/mpeg",
			}),
		},
	}

	events, err := provider.GenerateStream(context.Background(), params)
	if err == nil {
		t.Fatal("expected unsupported prompt context error")
	}
	if events != nil {
		t.Fatal("expected nil event channel on validation failure")
	}
	if !strings.Contains(err.Error(), "responses API does not support prompt context block types required by this request") {
		t.Fatalf("unexpected error: %v", err)
	}
}
