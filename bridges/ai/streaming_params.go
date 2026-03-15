package ai

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"github.com/rs/zerolog"
	"maunium.net/go/mautrix/bridgev2"
)

// buildResponsesAPIParams creates common Responses API parameters for both streaming and non-streaming paths
func (oc *AIClient) buildResponsesAPIParams(ctx context.Context, portal *bridgev2.Portal, meta *PortalMetadata, messages []openai.ChatCompletionMessageParamUnion) responses.ResponseNewParams {
	log := zerolog.Ctx(ctx)
	input := oc.convertToResponsesInput(messages, meta)
	params := oc.buildResponsesAgentLoopParams(ctx, meta, input, false)
	if len(params.Tools) > 0 {
		log.Debug().Int("count", len(params.Tools)).Msg("Added streaming turn tools")
	}
	return params
}
