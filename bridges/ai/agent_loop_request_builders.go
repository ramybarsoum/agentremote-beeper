package ai

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

type agentLoopRequestSettings struct {
	model           string
	maxTokens       int
	temperature     float64
	systemPrompt    string
	reasoningEffort string
}

func (oc *AIClient) buildAgentLoopRequestSettings(meta *PortalMetadata) agentLoopRequestSettings {
	return agentLoopRequestSettings{
		model:           oc.effectiveModelForAPI(meta),
		maxTokens:       oc.effectiveMaxTokens(meta),
		temperature:     oc.effectiveTemperature(meta),
		systemPrompt:    oc.effectivePrompt(meta),
		reasoningEffort: oc.effectiveReasoningEffort(meta),
	}
}

func (oc *AIClient) buildChatCompletionsAgentLoopParams(
	ctx context.Context,
	meta *PortalMetadata,
	messages []openai.ChatCompletionMessageParamUnion,
) openai.ChatCompletionNewParams {
	settings := oc.buildAgentLoopRequestSettings(meta)
	params := openai.ChatCompletionNewParams{
		Model:    settings.model,
		Messages: messages,
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: param.NewOpt(true),
		},
		Tools: oc.selectedChatStreamingTools(ctx, meta),
	}
	if settings.maxTokens > 0 {
		params.MaxCompletionTokens = openai.Int(int64(settings.maxTokens))
	}
	if settings.temperature > 0 {
		params.Temperature = openai.Float(settings.temperature)
	}
	return params
}

func (oc *AIClient) buildResponsesAgentLoopParams(
	ctx context.Context,
	meta *PortalMetadata,
	input responses.ResponseInputParam,
	allowResolvedBossAgent bool,
) responses.ResponseNewParams {
	settings := oc.buildAgentLoopRequestSettings(meta)
	params := responses.ResponseNewParams{
		Model:           shared.ResponsesModel(settings.model),
		MaxOutputTokens: openai.Int(int64(settings.maxTokens)),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: input,
		},
		Tools: oc.selectedResponsesStreamingTools(ctx, meta, allowResolvedBossAgent),
	}
	if settings.systemPrompt != "" {
		params.Instructions = openai.String(settings.systemPrompt)
	}
	if settings.reasoningEffort != "" {
		params.Reasoning = shared.ReasoningParam{
			Effort: shared.ReasoningEffort(settings.reasoningEffort),
		}
	}
	logToolParamDuplicates(&oc.log, params.Tools)
	return params
}
