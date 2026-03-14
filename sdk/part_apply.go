package sdk

import (
	"strings"

	"github.com/beeper/agentremote/pkg/shared/citations"
)

// PartApplyOptions controls provider-specific edge cases when applying
// streamed UI/tool parts to a turn.
type PartApplyOptions struct {
	ResetMetadataOnStartMarkers     bool
	ResetMetadataOnEmptyMessageMeta bool
	ResetMetadataOnEmptyTextDelta   bool
	ResetMetadataOnAbort            bool
	ResetMetadataOnDataParts        bool
	HandleTerminalEvents            bool
	DefaultFinishReason             string
}

// ApplyStreamPart maps a canonical stream part onto a turn. It returns true when
// the part type is recognized and applied.
func ApplyStreamPart(turn *Turn, part map[string]any, opts PartApplyOptions) bool {
	if turn == nil || len(part) == 0 {
		return false
	}
	partType := strings.TrimSpace(partString(part, "type"))
	if partType == "" {
		return false
	}
	writer := turn.Writer()
	tools := writer.Tools()
	approvals := turn.Approvals()
	ctx := turn.Context()
	switch partType {
	case "start", "message-metadata":
		metadata, _ := part["messageMetadata"].(map[string]any)
		if len(metadata) > 0 {
			writer.MessageMetadata(ctx, metadata)
		} else if opts.ResetMetadataOnEmptyMessageMeta {
			writer.MessageMetadata(ctx, nil)
		}
	case "start-step":
		writer.StepStart(ctx)
	case "finish-step":
		writer.StepFinish(ctx)
	case "text-start", "reasoning-start":
		if opts.ResetMetadataOnStartMarkers {
			writer.MessageMetadata(ctx, nil)
		}
	case "text-delta":
		if delta := partString(part, "delta"); delta != "" {
			writer.TextDelta(ctx, delta)
		} else if opts.ResetMetadataOnEmptyTextDelta {
			writer.MessageMetadata(ctx, nil)
		}
	case "text-end":
		writer.FinishText(ctx)
	case "reasoning-delta":
		if delta := partString(part, "delta"); delta != "" {
			writer.ReasoningDelta(ctx, delta)
		} else if opts.ResetMetadataOnEmptyTextDelta {
			writer.MessageMetadata(ctx, nil)
		}
	case "reasoning-end":
		writer.FinishReasoning(ctx)
	case "tool-input-start":
		tools.EnsureInputStart(ctx, partString(part, "toolCallId"), nil, ToolInputOptions{
			ToolName:         partString(part, "toolName"),
			ProviderExecuted: partBool(part, "providerExecuted"),
		})
	case "tool-input-delta":
		tools.InputDelta(ctx, partString(part, "toolCallId"), "", partString(part, "inputTextDelta"), partBool(part, "providerExecuted"))
	case "tool-input-available":
		tools.Input(ctx, partString(part, "toolCallId"), partString(part, "toolName"), part["input"], partBool(part, "providerExecuted"))
	case "tool-output-available":
		tools.Output(ctx, partString(part, "toolCallId"), part["output"], ToolOutputOptions{
			ProviderExecuted: partBool(part, "providerExecuted"),
		})
	case "tool-output-error":
		tools.OutputError(ctx, partString(part, "toolCallId"), partString(part, "errorText"), partBool(part, "providerExecuted"))
	case "tool-output-denied":
		tools.Denied(ctx, partString(part, "toolCallId"))
	case "tool-approval-request":
		approvals.EmitRequest(ctx, partString(part, "approvalId"), partString(part, "toolCallId"))
	case "tool-approval-response":
		approvals.Respond(ctx, partString(part, "approvalId"), partString(part, "toolCallId"), partBool(part, "approved"), partString(part, "reason"))
	case "file":
		writer.File(ctx, partString(part, "url"), partString(part, "mediaType"))
	case "source-document":
		writer.SourceDocument(ctx, citations.SourceDocument{
			ID:        partString(part, "sourceId"),
			Title:     partString(part, "title"),
			MediaType: partString(part, "mediaType"),
			Filename:  partString(part, "filename"),
		})
	case "source-url":
		writer.SourceURL(ctx, citations.SourceCitation{
			URL:   partString(part, "url"),
			Title: partString(part, "title"),
		})
	case "error":
		writer.Error(ctx, partString(part, "errorText"))
	case "finish":
		if !opts.HandleTerminalEvents {
			return false
		}
		finishReason := partString(part, "finishReason")
		if finishReason == "" {
			finishReason = strings.TrimSpace(opts.DefaultFinishReason)
		}
		if finishReason == "" {
			finishReason = "stop"
		}
		turn.End(finishReason)
	case "abort":
		if !opts.HandleTerminalEvents {
			return false
		}
		if opts.ResetMetadataOnAbort {
			writer.MessageMetadata(ctx, nil)
		}
		turn.Abort(partString(part, "reason"))
	default:
		if strings.HasPrefix(partType, "data-") {
			if opts.ResetMetadataOnDataParts {
				writer.MessageMetadata(ctx, nil)
			}
			writer.RawPart(ctx, part)
			return true
		}
		return false
	}
	return true
}

func partString(part map[string]any, key string) string {
	return strings.TrimSpace(stringValue(part[key]))
}

func partBool(part map[string]any, key string) bool {
	value, _ := part[key].(bool)
	return value
}
