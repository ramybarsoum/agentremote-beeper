package sdk

import (
	"encoding/json"
	"time"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"

	"github.com/beeper/agentremote"
)

// ImportedTurn represents a historical turn for backfill.
type ImportedTurn struct {
	ID           string
	Role         string // "user", "assistant", "system"
	Text         string
	HTML         string
	Reasoning    string
	ToolCalls    []ImportedToolCall
	Citations    []ImportedCitation
	Files        []ImportedFile
	Agent        *Agent
	Sender       bridgev2.EventSender
	Timestamp    time.Time
	Metadata     map[string]any
	FinishReason string
}

// ImportedToolCall represents a tool call in a historical turn.
type ImportedToolCall struct {
	ID     string
	Name   string
	Input  string
	Output string
}

// ImportedCitation represents a citation in a historical turn.
type ImportedCitation struct {
	URL   string
	Title string
}

// ImportedFile represents a file attachment in a historical turn.
type ImportedFile struct {
	URL       string
	MediaType string
}

// BackfillParams configures a backfill request.
type BackfillParams struct {
	Forward         bool
	Count           int
	AnchorTimestamp time.Time
}

// ConvertTranscriptTurn converts a historical turn into one backfill message.
func ConvertTranscriptTurn(turn *ImportedTurn, idPrefix string) *bridgev2.BackfillMessage {
	if turn == nil {
		return nil
	}
	msgID := turn.ID
	if msgID == "" {
		msgID = string(agentremote.NewMessageID(idPrefix))
	}

	body := turn.Text
	htmlBody := turn.HTML
	if htmlBody == "" && body != "" {
		rendered := format.RenderMarkdown(body, true, true)
		htmlBody = rendered.FormattedBody
	}

	content := &event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    body,
	}
	if htmlBody != "" {
		content.Format = event.FormatHTML
		content.FormattedBody = htmlBody
	}

	meta := &agentremote.BaseMessageMetadata{
		Role:            turn.Role,
		Body:            body,
		FinishReason:    turn.FinishReason,
		TurnID:          turn.ID,
		ThinkingContent: turn.Reasoning,
	}
	if turn.Agent != nil {
		meta.AgentID = turn.Agent.ID
	}
	if len(turn.ToolCalls) > 0 {
		meta.ToolCalls = make([]agentremote.ToolCallMetadata, len(turn.ToolCalls))
		for i, tc := range turn.ToolCalls {
			meta.ToolCalls[i] = agentremote.ToolCallMetadata{
				CallID:   tc.ID,
				ToolName: tc.Name,
				Status:   "completed",
				Input:    parseJSONOrWrap(tc.Input),
				Output:   parseJSONOrWrap(tc.Output),
			}
		}
	}

	ts := turn.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	return &bridgev2.BackfillMessage{
		ConvertedMessage: &bridgev2.ConvertedMessage{
			Parts: []*bridgev2.ConvertedMessagePart{{
				ID:         networkid.PartID("0"),
				Type:       event.EventMessage,
				Content:    content,
				DBMetadata: meta,
			}},
		},
		Sender:    turn.Sender,
		Timestamp: ts,
		ID:        networkid.MessageID(msgID),
	}
}

// ConvertTranscriptTurns converts a sequence of historical turns into backfill messages.
func ConvertTranscriptTurns(turns []*ImportedTurn, idPrefix string) []*bridgev2.BackfillMessage {
	if len(turns) == 0 {
		return nil
	}
	out := make([]*bridgev2.BackfillMessage, 0, len(turns))
	for _, turn := range turns {
		if msg := ConvertTranscriptTurn(turn, idPrefix); msg != nil {
			out = append(out, msg)
		}
	}
	return out
}

func parseJSONOrWrap(s string) map[string]any {
	if s == "" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err == nil {
		return m
	}
	return map[string]any{"raw": s}
}
