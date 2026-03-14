package sdk

import (
	"context"
	"strings"

	"maunium.net/go/mautrix/bridgev2"

	"github.com/beeper/agentremote/pkg/shared/citations"
	"github.com/beeper/agentremote/pkg/shared/streamui"
)

// ToolInputOptions controls how a tool input start is represented in the SDK UI stream.
type ToolInputOptions struct {
	ToolName         string
	ProviderExecuted bool
	DisplayTitle     string
}

// ToolOutputOptions controls how a tool output is represented in the SDK UI stream.
type ToolOutputOptions struct {
	ProviderExecuted bool
	Streaming        bool
}

// turnAccessor provides shared valid/portal checks for turn-scoped controllers.
type turnAccessor struct {
	turn *Turn
}

func (a *turnAccessor) valid() bool { return a != nil && a.turn != nil }

func (a *turnAccessor) portal() *bridgev2.Portal {
	if !a.valid() || a.turn.conv == nil {
		return nil
	}
	return a.turn.conv.portal
}

// TurnStream is the provider-facing streaming surface for a turn.
type TurnStream struct {
	turnAccessor
}

// ToolsController is the turn-owned tool streaming surface.
type ToolsController struct {
	turnAccessor
}

// Stream returns the turn's provider-facing streaming surface.
func (t *Turn) Stream() *TurnStream {
	if t == nil {
		return nil
	}
	return &TurnStream{turnAccessor{turn: t}}
}

// Emitter returns the underlying stream emitter as an escape hatch.
func (s *TurnStream) Emitter() *streamui.Emitter {
	if !s.valid() {
		return nil
	}
	return s.turn.emitter
}

// SetTransport configures a custom transport for streamed turn events.
func (s *TurnStream) SetTransport(hook func(turnID string, seq int, content map[string]any, txnID string) bool) {
	if !s.valid() {
		return
	}
	s.turn.streamHook = hook
}

// TextDelta emits a text delta.
func (s *TurnStream) TextDelta(text string) {
	if !s.valid() {
		return
	}
	s.turn.WriteText(text)
}

// ReasoningDelta emits a reasoning delta.
func (s *TurnStream) ReasoningDelta(text string) {
	if !s.valid() {
		return
	}
	s.turn.WriteReasoning(text)
}

// Error emits a UI error event for the turn.
func (s *TurnStream) Error(text string) {
	if !s.valid() {
		return
	}
	s.turn.ensureStarted()
	s.turn.emitter.EmitUIError(s.turn.turnCtx, s.portal(), text)
}

// TextEnd closes the current text stream part.
func (s *TurnStream) TextEnd() {
	if !s.valid() {
		return
	}
	s.turn.FinishText()
}

// ReasoningEnd closes the current reasoning stream part.
func (s *TurnStream) ReasoningEnd() {
	if !s.valid() {
		return
	}
	s.turn.FinishReasoning()
}

// Tools returns the turn's tool streaming controller.
func (t *Turn) Tools() *ToolsController {
	if t == nil {
		return nil
	}
	return &ToolsController{turnAccessor{turn: t}}
}

// EnsureInputStart ensures the tool input UI exists and optionally publishes input.
func (c *ToolsController) EnsureInputStart(toolCallID string, input any, opts ToolInputOptions) {
	if !c.valid() || strings.TrimSpace(toolCallID) == "" {
		return
	}
	c.turn.ensureStarted()
	toolName := strings.TrimSpace(opts.ToolName)
	displayTitle := strings.TrimSpace(opts.DisplayTitle)
	if displayTitle == "" {
		displayTitle = streamui.ToolDisplayTitle(toolName)
	}
	c.turn.emitter.EnsureUIToolInputStart(c.turn.turnCtx, c.portal(), toolCallID, toolName, opts.ProviderExecuted, displayTitle, nil)
	if input != nil {
		c.turn.emitter.EmitUIToolInputAvailable(c.turn.turnCtx, c.portal(), toolCallID, toolName, input, opts.ProviderExecuted)
	}
}

// InputDelta emits a tool input delta.
func (c *ToolsController) InputDelta(toolCallID, delta string, providerExecuted bool) {
	if !c.valid() {
		return
	}
	c.turn.ensureStarted()
	c.turn.emitter.EmitUIToolInputDelta(c.turn.turnCtx, c.portal(), toolCallID, "", delta, providerExecuted)
}

// Input emits a complete tool input payload.
func (c *ToolsController) Input(toolCallID, toolName string, input any, providerExecuted bool) {
	if !c.valid() {
		return
	}
	c.turn.ensureStarted()
	c.turn.emitter.EmitUIToolInputAvailable(c.turn.turnCtx, c.portal(), toolCallID, toolName, input, providerExecuted)
}

// Output emits a tool output payload.
func (c *ToolsController) Output(toolCallID string, output any, opts ToolOutputOptions) {
	if !c.valid() {
		return
	}
	c.turn.ensureStarted()
	c.turn.emitter.EmitUIToolOutputAvailable(c.turn.turnCtx, c.portal(), toolCallID, output, opts.ProviderExecuted, opts.Streaming)
}

// OutputError emits a tool error payload.
func (c *ToolsController) OutputError(toolCallID, errText string, providerExecuted bool) {
	if !c.valid() {
		return
	}
	c.turn.ensureStarted()
	c.turn.emitter.EmitUIToolOutputError(c.turn.turnCtx, c.portal(), toolCallID, errText, providerExecuted)
}

// Denied emits a denied tool result.
func (c *ToolsController) Denied(toolCallID string) {
	if !c.valid() {
		return
	}
	c.turn.ToolDenied(toolCallID)
}

// Backward-compatible TurnStream tool helpers delegate to the turn-owned tools controller.
func (s *TurnStream) EnsureToolInputStart(toolCallID string, input any, opts ToolInputOptions) {
	if !s.valid() {
		return
	}
	s.turn.Tools().EnsureInputStart(toolCallID, input, opts)
}

func (s *TurnStream) ToolInputDelta(toolCallID, delta string, providerExecuted bool) {
	if !s.valid() {
		return
	}
	s.turn.Tools().InputDelta(toolCallID, delta, providerExecuted)
}

func (s *TurnStream) ToolInput(toolCallID, toolName string, input any, providerExecuted bool) {
	if !s.valid() {
		return
	}
	s.turn.Tools().Input(toolCallID, toolName, input, providerExecuted)
}

func (s *TurnStream) ToolOutput(toolCallID string, output any, opts ToolOutputOptions) {
	if !s.valid() {
		return
	}
	s.turn.Tools().Output(toolCallID, output, opts)
}

func (s *TurnStream) ToolOutputError(toolCallID, errText string, providerExecuted bool) {
	if !s.valid() {
		return
	}
	s.turn.Tools().OutputError(toolCallID, errText, providerExecuted)
}

func (s *TurnStream) ToolDenied(toolCallID string) {
	if !s.valid() {
		return
	}
	s.turn.Tools().Denied(toolCallID)
}

// SourceURL emits a source URL citation.
func (s *TurnStream) SourceURL(url, title string) {
	if !s.valid() {
		return
	}
	s.turn.AddSourceURL(url, title)
}

// SourceCitation emits a source URL citation from a structured citation object.
func (s *TurnStream) SourceCitation(citation citations.SourceCitation) {
	if !s.valid() {
		return
	}
	s.turn.AddSourceURL(citation.URL, citation.Title)
}

// SourceDocument emits a source document citation.
func (s *TurnStream) SourceDocument(document citations.SourceDocument) {
	if !s.valid() {
		return
	}
	s.turn.AddSourceDocument(document.ID, document.Title, document.MediaType, document.Filename)
}

// File emits a generated file part.
func (s *TurnStream) File(url, mediaType string) {
	if !s.valid() {
		return
	}
	s.turn.AddFile(url, mediaType)
}

// GeneratedFile emits a generated file part from a structured file object.
func (s *TurnStream) GeneratedFile(file citations.GeneratedFilePart) {
	if !s.valid() {
		return
	}
	s.turn.AddFile(file.URL, file.MediaType)
}

// StepStart begins a visual step group.
func (s *TurnStream) StepStart() {
	if !s.valid() {
		return
	}
	s.turn.StepStart()
}

// StepFinish ends a visual step group.
func (s *TurnStream) StepFinish() {
	if !s.valid() {
		return
	}
	s.turn.StepFinish()
}

// Metadata merges message metadata for the turn.
func (s *TurnStream) Metadata(metadata map[string]any) {
	if !s.valid() {
		return
	}
	s.turn.SetMetadata(metadata)
}

// ApprovalController is the turn-owned approval surface.
type ApprovalController struct {
	turnAccessor
}

// Approvals returns the turn's approval controller.
func (t *Turn) Approvals() *ApprovalController {
	if t == nil {
		return nil
	}
	return &ApprovalController{turnAccessor{turn: t}}
}

// SetHandler configures a provider-specific approval handler for this turn.
func (a *ApprovalController) SetHandler(handler func(ctx context.Context, turn *Turn, req ApprovalRequest) ApprovalHandle) {
	if !a.valid() {
		return
	}
	a.turn.approvalRequester = handler
}

// Request creates a new approval request.
func (a *ApprovalController) Request(req ApprovalRequest) ApprovalHandle {
	if !a.valid() {
		return nil
	}
	return a.turn.RequestApproval(req)
}

// EmitRequest emits the approval-request UI state for a provider-managed approval.
func (a *ApprovalController) EmitRequest(approvalID, toolCallID string) {
	if !a.valid() {
		return
	}
	a.turn.ensureStarted()
	a.turn.emitter.EmitUIToolApprovalRequest(a.turn.turnCtx, a.portal(), approvalID, toolCallID)
}

// Respond emits the approval-response UI state for a provider-managed approval.
func (a *ApprovalController) Respond(approvalID, toolCallID string, approved bool, reason string) {
	if !a.valid() {
		return
	}
	a.turn.ensureStarted()
	a.turn.emitter.EmitUIToolApprovalResponse(a.turn.turnCtx, a.portal(), approvalID, toolCallID, approved, reason)
}
