package ai

import (
	"strings"

	"github.com/openai/openai-go/v3"
	"maunium.net/go/mautrix/id"
)

func (oc *AIClient) getSteeringMessages(roomID id.RoomID) []string {
	if oc == nil || roomID == "" {
		return nil
	}
	steerItems := oc.drainSteerQueue(roomID)
	if len(steerItems) == 0 {
		return nil
	}

	messages := make([]string, 0, len(steerItems))
	for _, item := range steerItems {
		if item.pending.Type != pendingTypeText {
			continue
		}
		prompt := strings.TrimSpace(item.prompt)
		if prompt == "" {
			prompt = item.pending.MessageBody
		}
		prompt = strings.TrimSpace(prompt)
		if prompt == "" {
			continue
		}
		messages = append(messages, prompt)
	}
	return messages
}

func buildSteeringUserMessages(prompts []string) []openai.ChatCompletionMessageParamUnion {
	if len(prompts) == 0 {
		return nil
	}
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(prompts))
	for _, prompt := range prompts {
		prompt = strings.TrimSpace(prompt)
		if prompt == "" {
			continue
		}
		messages = append(messages, openai.UserMessage(prompt))
	}
	return messages
}

func (s *streamingState) addPendingSteeringPrompts(prompts []string) {
	if s == nil || len(prompts) == 0 {
		return
	}
	s.pendingSteeringPrompts = append(s.pendingSteeringPrompts, prompts...)
}

func (s *streamingState) consumePendingSteeringPrompts() []string {
	if s == nil || len(s.pendingSteeringPrompts) == 0 {
		return nil
	}
	prompts := append([]string(nil), s.pendingSteeringPrompts...)
	s.pendingSteeringPrompts = nil
	return prompts
}
