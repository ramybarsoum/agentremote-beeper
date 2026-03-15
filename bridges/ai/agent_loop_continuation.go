package ai

import "github.com/openai/openai-go/v3"

func (oc *AIClient) buildChatAgentLoopContinuationMessages(
	state *streamingState,
	currentMessages []openai.ChatCompletionMessageParamUnion,
	assistantMsg openai.ChatCompletionAssistantMessageParam,
	steeringPrompts []string,
) []openai.ChatCompletionMessageParamUnion {
	currentMessages = append(currentMessages, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistantMsg})
	for _, output := range state.pendingFunctionOutputs {
		currentMessages = append(currentMessages, openai.ToolMessage(output.output, output.callID))
	}
	currentMessages = append(currentMessages, buildSteeringUserMessages(steeringPrompts)...)
	return currentMessages
}
