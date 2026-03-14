package sdk

import (
	"context"

	"maunium.net/go/mautrix/bridgev2"

	"github.com/beeper/agentremote/pkg/shared/citations"
	"github.com/beeper/agentremote/pkg/shared/streamui"
)

// SemanticStream applies SDK-owned semantic stream operations onto a UI state.
// Bridges can use this without constructing a full Turn.
type SemanticStream struct {
	State   *streamui.UIState
	Emitter *streamui.Emitter
	Portal  *bridgev2.Portal
}

func (s *SemanticStream) valid() bool {
	return s != nil && s.State != nil && s.Emitter != nil
}

func (s *SemanticStream) MessageMetadata(ctx context.Context, metadata map[string]any) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIMessageMetadata(ctx, s.Portal, metadata)
}

func (s *SemanticStream) Start(ctx context.Context, metadata map[string]any) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIStart(ctx, s.Portal, metadata)
}

func (s *SemanticStream) StepStart(ctx context.Context) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIStepStart(ctx, s.Portal)
}

func (s *SemanticStream) StepFinish(ctx context.Context) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIStepFinish(ctx, s.Portal)
}

func (s *SemanticStream) TextDelta(ctx context.Context, delta string) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUITextDelta(ctx, s.Portal, delta)
}

func (s *SemanticStream) ReasoningDelta(ctx context.Context, delta string) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIReasoningDelta(ctx, s.Portal, delta)
}

func (s *SemanticStream) Error(ctx context.Context, errText string) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIError(ctx, s.Portal, errText)
}

func (s *SemanticStream) Abort(ctx context.Context, reason string) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIAbort(ctx, s.Portal, reason)
}

func (s *SemanticStream) ToolInputStart(ctx context.Context, toolCallID, toolName string, providerExecuted bool, displayTitle string) {
	if !s.valid() {
		return
	}
	s.Emitter.EnsureUIToolInputStart(ctx, s.Portal, toolCallID, toolName, providerExecuted, displayTitle, nil)
}

func (s *SemanticStream) ToolInputDelta(ctx context.Context, toolCallID, toolName, delta string, providerExecuted bool) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIToolInputDelta(ctx, s.Portal, toolCallID, toolName, delta, providerExecuted)
}

func (s *SemanticStream) ToolInputAvailable(ctx context.Context, toolCallID, toolName string, input any, providerExecuted bool) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIToolInputAvailable(ctx, s.Portal, toolCallID, toolName, input, providerExecuted)
}

func (s *SemanticStream) ToolInputError(ctx context.Context, toolCallID, toolName, rawInput, errText string, providerExecuted bool) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIToolInputError(ctx, s.Portal, toolCallID, toolName, rawInput, errText, providerExecuted)
}

func (s *SemanticStream) ToolOutputAvailable(ctx context.Context, toolCallID string, output any, providerExecuted, streaming bool) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIToolOutputAvailable(ctx, s.Portal, toolCallID, output, providerExecuted, streaming)
}

func (s *SemanticStream) ToolOutputError(ctx context.Context, toolCallID, errText string, providerExecuted bool) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIToolOutputError(ctx, s.Portal, toolCallID, errText, providerExecuted)
}

func (s *SemanticStream) ToolOutputDenied(ctx context.Context, toolCallID string) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIToolOutputDenied(ctx, s.Portal, toolCallID)
}

func (s *SemanticStream) ToolApprovalRequest(ctx context.Context, approvalID, toolCallID string) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIToolApprovalRequest(ctx, s.Portal, approvalID, toolCallID)
}

func (s *SemanticStream) ToolApprovalResponse(ctx context.Context, approvalID, toolCallID string, approved bool, reason string) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIToolApprovalResponse(ctx, s.Portal, approvalID, toolCallID, approved, reason)
	streamui.RecordApprovalResponse(s.State, approvalID, toolCallID, approved, reason)
}

func (s *SemanticStream) File(ctx context.Context, url, mediaType string) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUIFile(ctx, s.Portal, url, mediaType)
}

func (s *SemanticStream) SourceURL(ctx context.Context, citation citations.SourceCitation) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUISourceURL(ctx, s.Portal, citation)
}

func (s *SemanticStream) SourceDocument(ctx context.Context, document citations.SourceDocument) {
	if !s.valid() {
		return
	}
	s.Emitter.EmitUISourceDocument(ctx, s.Portal, document)
}
