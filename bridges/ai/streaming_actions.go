package ai

import (
	"context"

	"github.com/openai/openai-go/v3/responses"
)

type streamTurnActions struct {
	base  *streamingAdapterBase
	tools *streamToolRegistry
}

func newStreamTurnActions(base *streamingAdapterBase, tools *streamToolRegistry) streamTurnActions {
	return streamTurnActions{
		base:  base,
		tools: tools,
	}
}

func (a streamTurnActions) touchTyping() {
	if a.base != nil && a.base.touchTyping != nil {
		a.base.touchTyping()
	}
}

func (a streamTurnActions) signalToolStart() {
	a.touchTyping()
	if a.base != nil && a.base.typingSignals != nil {
		a.base.typingSignals.SignalToolStart()
	}
}

func (a streamTurnActions) textDelta(ctx context.Context, delta string, errText string, logMessage string) error {
	if a.base == nil {
		return nil
	}
	a.touchTyping()
	return a.base.oc.handleResponseOutputTextDelta(
		ctx,
		a.base.log,
		a.base.portal,
		a.base.state,
		a.base.meta,
		a.base.typingSignals,
		a.base.isHeartbeat,
		delta,
		errText,
		logMessage,
	)
}

func (a streamTurnActions) reasoningDelta(ctx context.Context, delta string, errText string, logMessage string) error {
	if a.base == nil {
		return nil
	}
	a.touchTyping()
	if a.base.typingSignals != nil {
		a.base.typingSignals.SignalReasoningDelta()
	}
	return a.base.oc.handleResponseReasoningTextDelta(
		ctx,
		a.base.log,
		a.base.portal,
		a.base.state,
		a.base.meta,
		a.base.isHeartbeat,
		delta,
		errText,
		logMessage,
	)
}

func (a streamTurnActions) refusalDelta(ctx context.Context, delta string) {
	if a.base == nil {
		return
	}
	a.touchTyping()
	a.base.oc.handleResponseRefusalDelta(ctx, a.base.portal, a.base.state, a.base.typingSignals, delta)
}

func (a streamTurnActions) refusalDone(ctx context.Context, refusal string) {
	if a.base == nil {
		return
	}
	a.base.oc.handleResponseRefusalDone(ctx, a.base.portal, a.base.state, refusal)
}

func (a streamTurnActions) responseOutputItemAdded(ctx context.Context, item responses.ResponseOutputItemUnion) {
	if a.base == nil {
		return
	}
	a.base.oc.handleResponseOutputItemAdded(ctx, a.base.portal, a.base.state, a.tools, item)
}

func (a streamTurnActions) responseOutputItemDone(ctx context.Context, item responses.ResponseOutputItemUnion) {
	if a.base == nil {
		return
	}
	a.base.oc.handleResponseOutputItemDone(ctx, a.base.portal, a.base.state, a.tools, item)
}

func (a streamTurnActions) customToolInputDelta(ctx context.Context, itemID string, item responses.ResponseOutputItemUnion, delta string) {
	if a.base == nil {
		return
	}
	a.base.oc.handleCustomToolInputDeltaFromOutputItem(ctx, a.base.portal, a.base.state, a.tools, itemID, item, delta)
}

func (a streamTurnActions) customToolInputDone(ctx context.Context, itemID string, item responses.ResponseOutputItemUnion, inputText string) {
	if a.base == nil {
		return
	}
	a.base.oc.handleCustomToolInputDoneFromOutputItem(ctx, a.base.portal, a.base.state, a.tools, itemID, item, inputText)
}

func (a streamTurnActions) mcpCallFailed(ctx context.Context, itemID string, item responses.ResponseOutputItemUnion) {
	if a.base == nil {
		return
	}
	a.base.oc.handleMCPCallFailedFromOutputItem(ctx, a.base.portal, a.base.state, a.tools, itemID, item)
}

func (a streamTurnActions) functionArgsDelta(ctx context.Context, itemID string, name string, delta string) {
	if a.base == nil {
		return
	}
	a.signalToolStart()
	a.base.oc.handleFunctionCallArgumentsDelta(ctx, a.base.portal, a.base.state, a.base.meta, a.tools, itemID, name, delta)
}

func (a streamTurnActions) functionArgsDone(ctx context.Context, itemID string, name string, arguments string, approvalFallbackForNonObject bool, logSuffix string) {
	if a.base == nil {
		return
	}
	a.signalToolStart()
	a.base.oc.handleFunctionCallArgumentsDone(ctx, a.base.log, a.base.portal, a.base.state, a.base.meta, a.tools, itemID, name, arguments, approvalFallbackForNonObject, logSuffix)
}

func (a streamTurnActions) providerToolInProgress(ctx context.Context, itemID string, toolName string, toolType ToolType) {
	if a.base == nil {
		return
	}
	a.signalToolStart()
	a.base.oc.handleProviderToolInProgress(ctx, a.base.portal, a.base.state, a.base.meta, a.tools, itemID, toolName, toolType)
}

func (a streamTurnActions) providerToolCompleted(ctx context.Context, itemID string, toolName string, toolType ToolType, failureText string) {
	if a.base == nil {
		return
	}
	a.touchTyping()
	a.base.oc.handleProviderToolCompleted(ctx, a.base.portal, a.base.state, a.tools, itemID, toolName, toolType, failureText)
}
