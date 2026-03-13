package sdk

import (
	"context"
	"time"

	"sync/atomic"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/id"

	"github.com/beeper/agentremote"
	"github.com/beeper/agentremote/pkg/shared/citations"
	"github.com/beeper/agentremote/pkg/shared/streamui"
	"github.com/beeper/agentremote/turns"
)

// Stream is a writer for streaming response chunks back to Beeper.
// It wraps streamui.Emitter and turns.StreamSession to emit the AI SDK
// UIMessage protocol.
type Stream struct {
	ctx     context.Context
	conv    *Conversation
	emitter *streamui.Emitter
	state   *streamui.UIState
	session *turns.StreamSession
	turnID  string
	started bool
	ended   bool
}

func newStream(ctx context.Context, conv *Conversation) *Stream {
	turnID := uuid.NewString()
	state := &streamui.UIState{TurnID: turnID}
	state.InitMaps()

	s := &Stream{
		ctx:    ctx,
		conv:   conv,
		state:  state,
		turnID: turnID,
	}

	s.emitter = &streamui.Emitter{
		State: state,
		Emit: func(ctx context.Context, portal *bridgev2.Portal, part map[string]any) {
			if s.session != nil {
				s.session.EmitPart(ctx, part)
			}
		},
	}

	// Create stream session with minimal params.
	if conv.portal != nil {
		var seq int
		logger := zerolog.Nop()
		s.session = turns.NewStreamSession(turns.StreamSessionParams{
			TurnID:  turnID,
			NextSeq: func() int { seq++; return seq },
			GetRoomID: func() id.RoomID {
				return conv.portal.MXID
			},
			GetStreamTarget: func() turns.StreamTarget {
				return turns.StreamTarget{}
			},
			GetSuppressSend: func() bool { return false },
			RuntimeFallbackFlag: &atomic.Bool{},
			GetEphemeralSender: func(ctx context.Context) (bridgev2.EphemeralSendingMatrixAPI, bool) {
				return nil, false
			},
			SendDebouncedEdit: func(ctx context.Context, force bool) error { return nil },
			Logger:            &logger,
		})
	}

	return s
}

func (s *Stream) ensureStarted() {
	if s.started || s.ended {
		return
	}
	s.started = true
	s.emitter.EmitUIStart(s.ctx, s.conv.portal, nil)
}

// WriteText sends a text chunk.
func (s *Stream) WriteText(text string) {
	s.ensureStarted()
	s.emitter.EmitUITextDelta(s.ctx, s.conv.portal, text)
}

// WriteReasoning sends a reasoning/thinking chunk.
func (s *Stream) WriteReasoning(text string) {
	s.ensureStarted()
	s.emitter.EmitUIReasoningDelta(s.ctx, s.conv.portal, text)
}

// ToolStart begins a tool call.
func (s *Stream) ToolStart(toolName, toolCallID string, providerExecuted bool) {
	s.ensureStarted()
	s.emitter.EnsureUIToolInputStart(s.ctx, s.conv.portal, toolCallID, toolName, providerExecuted, toolName, nil)
}

// ToolInputDelta sends a streaming tool input argument chunk.
func (s *Stream) ToolInputDelta(toolCallID, delta string) {
	s.ensureStarted()
	s.emitter.EmitUIToolInputDelta(s.ctx, s.conv.portal, toolCallID, "", delta, false)
}

// ToolInputAvailable sends the complete tool input.
func (s *Stream) ToolInputAvailable(toolCallID string, input any) {
	s.ensureStarted()
	s.emitter.EmitUIToolInputAvailable(s.ctx, s.conv.portal, toolCallID, "", input, false)
}

// ToolInputError reports an error in tool input parsing.
func (s *Stream) ToolInputError(toolCallID string, input any, errorText string) {
	s.ensureStarted()
	s.emitter.EmitUIToolInputError(s.ctx, s.conv.portal, toolCallID, "", input, errorText, false)
}

// ToolRequestApproval sends a tool approval prompt and blocks until the user responds.
func (s *Stream) ToolRequestApproval(toolCallID, toolName string) (ToolApprovalResponse, error) {
	s.ensureStarted()
	client := s.conv.client
	if client == nil || client.approvalFlow == nil || s.conv.portal == nil {
		return ToolApprovalResponse{}, nil
	}

	approvalID := "sdk-" + uuid.NewString()
	ttl := 10 * time.Minute

	_, created := client.approvalFlow.Register(approvalID, ttl, &pendingSDKApprovalData{
		RoomID:     s.conv.portal.MXID,
		TurnID:     s.turnID,
		ToolCallID: toolCallID,
		ToolName:   toolName,
	})
	if !created {
		return ToolApprovalResponse{}, nil
	}

	// Emit UI events for the approval request.
	s.emitter.EmitUIToolApprovalRequest(s.ctx, s.conv.portal, approvalID, toolCallID)

	// Send the approval prompt message.
	presentation := agentremote.ApprovalPromptPresentation{
		Title:       toolName,
		AllowAlways: true,
	}
	client.approvalFlow.SendPrompt(s.ctx, s.conv.portal, agentremote.SendPromptParams{
		ApprovalPromptMessageParams: agentremote.ApprovalPromptMessageParams{
			ApprovalID:   approvalID,
			ToolCallID:   toolCallID,
			ToolName:     toolName,
			TurnID:       s.turnID,
			Presentation: presentation,
			ExpiresAt:    time.Now().Add(ttl),
		},
		RoomID:    s.conv.portal.MXID,
		OwnerMXID: client.userLogin.UserMXID,
	})

	// Block until user decision.
	decision, ok := client.approvalFlow.Wait(s.ctx, approvalID)
	if !ok {
		reason := agentremote.ApprovalReasonTimeout
		if s.ctx.Err() != nil {
			reason = agentremote.ApprovalReasonCancelled
		}
		client.approvalFlow.FinishResolved(approvalID, agentremote.ApprovalDecisionPayload{
			ApprovalID: approvalID,
			Reason:     reason,
		})
		s.emitter.EmitUIToolApprovalResponse(s.ctx, s.conv.portal, approvalID, toolCallID, false, reason)
		return ToolApprovalResponse{Reason: reason}, nil
	}

	s.emitter.EmitUIToolApprovalResponse(s.ctx, s.conv.portal, approvalID, toolCallID, decision.Approved, decision.Reason)
	client.approvalFlow.FinishResolved(approvalID, decision)
	return ToolApprovalResponse{
		Approved: decision.Approved,
		Always:   decision.Always,
		Reason:   decision.Reason,
	}, nil
}

// ToolOutput sends the tool execution result.
func (s *Stream) ToolOutput(toolCallID string, output any) {
	s.ensureStarted()
	s.emitter.EmitUIToolOutputAvailable(s.ctx, s.conv.portal, toolCallID, output, false, false)
}

// ToolOutputError reports a tool execution error.
func (s *Stream) ToolOutputError(toolCallID, errorText string) {
	s.ensureStarted()
	s.emitter.EmitUIToolOutputError(s.ctx, s.conv.portal, toolCallID, errorText, false)
}

// ToolDenied reports that the tool execution was denied by the user.
func (s *Stream) ToolDenied(toolCallID string) {
	s.ensureStarted()
	s.emitter.EmitUIToolOutputDenied(s.ctx, s.conv.portal, toolCallID)
}

// AddSourceURL adds a source citation URL.
func (s *Stream) AddSourceURL(url, title string) {
	s.ensureStarted()
	s.emitter.EmitUISourceURL(s.ctx, s.conv.portal, citations.SourceCitation{
		URL:   url,
		Title: title,
	})
}

// AddSourceDocument adds a source document citation.
func (s *Stream) AddSourceDocument(docID, title, mediaType, filename string) {
	s.ensureStarted()
	s.emitter.EmitUISourceDocument(s.ctx, s.conv.portal, citations.SourceDocument{
		ID:        docID,
		Title:     title,
		MediaType: mediaType,
		Filename:  filename,
	})
}

// AddFile adds a generated file reference.
func (s *Stream) AddFile(url, mediaType string) {
	s.ensureStarted()
	s.emitter.EmitUIFile(s.ctx, s.conv.portal, url, mediaType)
}

// StepStart begins a visual step grouping.
func (s *Stream) StepStart() {
	s.ensureStarted()
	s.emitter.EmitUIStepStart(s.ctx, s.conv.portal)
}

// StepFinish ends a visual step grouping.
func (s *Stream) StepFinish() {
	s.ensureStarted()
	s.emitter.EmitUIStepFinish(s.ctx, s.conv.portal)
}

// SetMetadata sets message metadata (model, timing, usage).
func (s *Stream) SetMetadata(metadata map[string]any) {
	s.ensureStarted()
	s.emitter.EmitUIMessageMetadata(s.ctx, s.conv.portal, metadata)
}

// End finishes the stream with a reason.
func (s *Stream) End(finishReason string) {
	if s.ended {
		return
	}
	s.ensureStarted()
	s.ended = true
	s.emitter.EmitUIFinish(s.ctx, s.conv.portal, finishReason, nil)
	if s.session != nil {
		s.session.End(s.ctx, turns.EndReasonFinish)
	}
}

// EndWithError finishes the stream with an error.
func (s *Stream) EndWithError(errText string) {
	if s.ended {
		return
	}
	s.ensureStarted()
	s.ended = true
	s.emitter.EmitUIError(s.ctx, s.conv.portal, errText)
	s.emitter.EmitUIFinish(s.ctx, s.conv.portal, "error", nil)
	if s.session != nil {
		s.session.End(s.ctx, turns.EndReasonError)
	}
}

// Abort aborts the stream.
func (s *Stream) Abort(reason string) {
	if s.ended {
		return
	}
	s.ensureStarted()
	s.ended = true
	s.emitter.EmitUIAbort(s.ctx, s.conv.portal, reason)
	if s.session != nil {
		s.session.End(s.ctx, turns.EndReasonDisconnect)
	}
}

// Emitter returns the underlying streamui.Emitter for escape hatch access.
func (s *Stream) Emitter() *streamui.Emitter { return s.emitter }

// UIState returns the underlying streamui.UIState.
func (s *Stream) UIState() *streamui.UIState { return s.state }

// Session returns the underlying turns.StreamSession.
func (s *Stream) Session() *turns.StreamSession { return s.session }
