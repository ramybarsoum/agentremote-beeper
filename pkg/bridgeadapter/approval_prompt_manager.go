package bridgeadapter

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// ApprovalPromptManagerConfig holds the bridge-specific callbacks needed by ApprovalPromptManager.
type ApprovalPromptManagerConfig struct {
	// Login returns the current UserLogin. Required.
	Login func() *bridgev2.UserLogin

	// Sender returns the EventSender to use for a given portal (e.g. the agent ghost).
	Sender func(portal *bridgev2.Portal) bridgev2.EventSender

	// SendMsg sends a ConvertedMessage to a portal. Returns the Matrix event ID
	// and the network message ID. If nil, bridgeadapter.SendViaPortal is used
	// with the IDPrefix and LogKey from SendPromptParams.
	SendMsg func(ctx context.Context, portal *bridgev2.Portal, converted *bridgev2.ConvertedMessage) (id.EventID, networkid.MessageID, error)

	// Resolve is called when a matching reaction is received and the approval
	// should be resolved. The bridge converts match.Decision to its internal type.
	Resolve func(ctx context.Context, roomID id.RoomID, match ApprovalPromptReactionMatch) error

	// OnError is called when resolution fails or the prompt is expired.
	// Bridges typically send a system notice or rejection event here.
	OnError func(ctx context.Context, portal *bridgev2.Portal, approvalID string, err error)

	// DBMetadata produces the bridge-specific metadata for the approval prompt
	// message part. The return value is stored in ConvertedMessagePart.DBMetadata
	// (which is typed as any upstream). If nil, a default *BaseMessageMetadata is used.
	DBMetadata func(prompt ApprovalPromptMessage) any

	// IDPrefix is used when SendMsg is nil to generate message IDs (e.g. "codex", "openclaw").
	IDPrefix string

	// LogKey is used when SendMsg is nil for zerolog field names (e.g. "codex_msg_id").
	LogKey string

	// SendTimeout is the timeout applied to the send context. Defaults to 10s.
	SendTimeout time.Duration

	// BackgroundContext optionally returns a background context detached from
	// the request lifecycle. If nil, the passed context is used directly.
	BackgroundContext func(ctx context.Context) context.Context
}

// ApprovalPromptManager owns the full lifecycle of approval prompts:
// building & sending prompt messages, registering them, prefilling reactions,
// matching incoming reactions, resolving approvals, and cleaning up.
type ApprovalPromptManager struct {
	store         *ApprovalPromptStore
	login         func() *bridgev2.UserLogin
	sender        func(portal *bridgev2.Portal) bridgev2.EventSender
	sendMsg       func(ctx context.Context, portal *bridgev2.Portal, converted *bridgev2.ConvertedMessage) (id.EventID, networkid.MessageID, error)
	resolve       func(ctx context.Context, roomID id.RoomID, match ApprovalPromptReactionMatch) error
	onError       func(ctx context.Context, portal *bridgev2.Portal, approvalID string, err error)
	dbMetadata    func(prompt ApprovalPromptMessage) any
	idPrefix      string
	logKey        string
	sendTimeout   time.Duration
	backgroundCtx func(ctx context.Context) context.Context
}

// NewApprovalPromptManager creates an ApprovalPromptManager from the given config.
func NewApprovalPromptManager(cfg ApprovalPromptManagerConfig) *ApprovalPromptManager {
	timeout := cfg.SendTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &ApprovalPromptManager{
		store:         NewApprovalPromptStore(),
		login:         cfg.Login,
		sender:        cfg.Sender,
		sendMsg:       cfg.SendMsg,
		resolve:       cfg.Resolve,
		onError:       cfg.OnError,
		dbMetadata:    cfg.DBMetadata,
		idPrefix:      cfg.IDPrefix,
		logKey:        cfg.LogKey,
		sendTimeout:   timeout,
		backgroundCtx: cfg.BackgroundContext,
	}
}

// Drop removes a registered approval prompt by ID.
func (m *ApprovalPromptManager) Drop(approvalID string) {
	if m == nil || m.store == nil {
		return
	}
	m.store.Drop(approvalID)
}

// SendPromptParams holds the parameters for sending an approval prompt.
type SendPromptParams struct {
	ApprovalPromptMessageParams
	RoomID    id.RoomID
	OwnerMXID id.UserID
}

// SendPrompt builds an approval prompt message, registers it, sends it via
// the configured send function, binds the resulting event ID, and queues
// prefill reactions.
func (m *ApprovalPromptManager) SendPrompt(ctx context.Context, portal *bridgev2.Portal, params SendPromptParams) {
	if m == nil || portal == nil || portal.MXID == "" {
		return
	}
	login := m.loginOrNil()
	if login == nil {
		return
	}

	prompt := BuildApprovalPromptMessage(params.ApprovalPromptMessageParams)

	m.store.Register(ApprovalPromptRegistration{
		ApprovalID: strings.TrimSpace(params.ApprovalID),
		RoomID:     params.RoomID,
		OwnerMXID:  params.OwnerMXID,
		ToolCallID: strings.TrimSpace(params.ToolCallID),
		ToolName:   strings.TrimSpace(params.ToolName),
		TurnID:     strings.TrimSpace(params.TurnID),
		ExpiresAt:  params.ExpiresAt,
		Options:    prompt.Options,
	})

	var dbMeta any
	if m.dbMetadata != nil {
		dbMeta = m.dbMetadata(prompt)
	} else {
		dbMeta = &BaseMessageMetadata{
			Role:               "assistant",
			CanonicalSchema:    "ai-sdk-ui-message-v1",
			CanonicalUIMessage: prompt.UIMessage,
		}
	}

	converted := &bridgev2.ConvertedMessage{
		Parts: []*bridgev2.ConvertedMessagePart{{
			ID:         networkid.PartID("0"),
			Type:       event.EventMessage,
			Content:    &event.MessageEventContent{MsgType: event.MsgNotice, Body: prompt.Body},
			Extra:      prompt.Raw,
			DBMetadata: dbMeta,
		}},
	}

	eventID, msgID, err := m.send(ctx, portal, converted)
	if err != nil {
		return
	}

	m.store.BindPromptEvent(strings.TrimSpace(params.ApprovalID), eventID)

	m.sendPrefillReactions(ctx, portal, login, msgID, prompt.Options)
}

// HandleReaction checks whether a reaction targets a known approval prompt.
// If so, it resolves the approval and redacts prompt reactions. Returns true
// if the reaction was handled (even if resolution failed).
func (m *ApprovalPromptManager) HandleReaction(ctx context.Context, msg *bridgev2.MatrixReaction, targetEventID id.EventID, emoji string) bool {
	if m == nil || m.store == nil || msg == nil || msg.Event == nil || msg.Portal == nil {
		return false
	}
	match := m.store.MatchReaction(targetEventID, msg.Event.Sender, emoji, time.Now())
	if !match.KnownPrompt {
		return false
	}

	keepEventID := id.EventID("")
	if match.ShouldResolve {
		if m.resolve != nil {
			if err := m.resolve(ctx, msg.Portal.MXID, match); err != nil {
				if m.onError != nil {
					m.onError(ctx, msg.Portal, match.ApprovalID, err)
				}
			} else {
				keepEventID = msg.Event.ID
			}
		}
	} else if match.RejectReason == "expired" && m.onError != nil {
		m.onError(ctx, msg.Portal, match.ApprovalID, ErrApprovalExpired)
	}

	sender := bridgev2.EventSender{}
	if m.sender != nil {
		sender = m.sender(msg.Portal)
	}
	redactLogin := m.loginOrNil()
	redactPortal := msg.Portal
	redactTarget := msg.TargetMessage
	triggerID := msg.Event.ID
	go func() {
		redactCtx := context.Background()
		if m.backgroundCtx != nil {
			redactCtx = m.backgroundCtx(redactCtx)
		}
		_ = RedactApprovalPromptReactions(
			redactCtx,
			redactLogin,
			redactPortal,
			sender,
			redactTarget,
			triggerID,
			keepEventID,
		)
	}()
	return true
}

// MatchReaction exposes the underlying store's MatchReaction for bridges that
// need custom handling beyond what HandleReaction provides.
func (m *ApprovalPromptManager) MatchReaction(targetEventID id.EventID, sender id.UserID, key string, now time.Time) ApprovalPromptReactionMatch {
	if m == nil || m.store == nil {
		return ApprovalPromptReactionMatch{}
	}
	return m.store.MatchReaction(targetEventID, sender, key, now)
}

func (m *ApprovalPromptManager) loginOrNil() *bridgev2.UserLogin {
	if m == nil || m.login == nil {
		return nil
	}
	return m.login()
}

func (m *ApprovalPromptManager) send(ctx context.Context, portal *bridgev2.Portal, converted *bridgev2.ConvertedMessage) (id.EventID, networkid.MessageID, error) {
	sendCtx := ctx
	if m.backgroundCtx != nil {
		sendCtx = m.backgroundCtx(ctx)
	}
	sendCtx, cancel := context.WithTimeout(sendCtx, m.sendTimeout)
	defer cancel()

	if m.sendMsg != nil {
		return m.sendMsg(sendCtx, portal, converted)
	}
	login := m.loginOrNil()
	if login == nil {
		return "", "", nil
	}
	sender := bridgev2.EventSender{}
	if m.sender != nil {
		sender = m.sender(portal)
	}
	return SendViaPortal(SendViaPortalParams{
		Login:     login,
		Portal:    portal,
		Sender:    sender,
		IDPrefix:  m.idPrefix,
		LogKey:    m.logKey,
		Converted: converted,
	})
}

func (m *ApprovalPromptManager) sendPrefillReactions(ctx context.Context, portal *bridgev2.Portal, login *bridgev2.UserLogin, msgID networkid.MessageID, options []ApprovalOption) {
	if login == nil || portal == nil || msgID == "" {
		return
	}
	sender := bridgev2.EventSender{}
	if m.sender != nil {
		sender = m.sender(portal)
	}
	now := time.Now()
	seenKeys := map[string]struct{}{}
	for _, option := range options {
		for _, key := range option.prefillKeys() {
			if key == "" {
				continue
			}
			if _, exists := seenKeys[key]; exists {
				continue
			}
			seenKeys[key] = struct{}{}
			login.QueueRemoteEvent(&RemoteReaction{
				Portal:        portal.PortalKey,
				Sender:        sender,
				TargetMessage: msgID,
				Emoji:         key,
				EmojiID:       networkid.EmojiID(key),
				Timestamp:     now,
				LogKey:        m.logKey,
			})
		}
	}
}

// ---------------------------------------------------------------------------
// Shared reaction helpers
// ---------------------------------------------------------------------------

// MatrixSenderID returns the standard networkid.UserID for a Matrix user.
func MatrixSenderID(userID id.UserID) networkid.UserID {
	if userID == "" {
		return ""
	}
	return networkid.UserID("mxid:" + userID.String())
}

// EnsureReactionContent lazily parses the reaction content from a MatrixReaction.
func EnsureReactionContent(msg *bridgev2.MatrixReaction) *event.ReactionEventContent {
	if msg == nil {
		return nil
	}
	if msg.Content != nil {
		return msg.Content
	}
	if msg.Event == nil || len(msg.Event.Content.VeryRaw) == 0 {
		return nil
	}
	var parsed event.ReactionEventContent
	if err := json.Unmarshal(msg.Event.Content.VeryRaw, &parsed); err != nil {
		return nil
	}
	msg.Content = &parsed
	return msg.Content
}

// PreHandleApprovalReaction implements the common PreHandleMatrixReaction logic
// shared by all bridges. The SenderID is derived from the Matrix sender.
func PreHandleApprovalReaction(msg *bridgev2.MatrixReaction) (bridgev2.MatrixReactionPreResponse, error) {
	if msg == nil || msg.Event == nil {
		return bridgev2.MatrixReactionPreResponse{}, bridgev2.ErrReactionsNotSupported
	}
	content := EnsureReactionContent(msg)
	if content == nil {
		return bridgev2.MatrixReactionPreResponse{}, bridgev2.ErrReactionsNotSupported
	}
	return bridgev2.MatrixReactionPreResponse{
		SenderID:     MatrixSenderID(msg.Event.Sender),
		Emoji:        normalizeReactionKey(content.RelatesTo.Key),
		MaxReactions: 1,
	}, nil
}

// ReactionContext holds the extracted emoji and target event ID from a reaction.
type ReactionContext struct {
	Emoji         string
	TargetEventID id.EventID
}

// ExtractReactionContext pulls the emoji and target event ID from a MatrixReaction.
func ExtractReactionContext(msg *bridgev2.MatrixReaction) ReactionContext {
	content := EnsureReactionContent(msg)
	emoji := ""
	if msg != nil && msg.PreHandleResp != nil {
		emoji = msg.PreHandleResp.Emoji
	}
	if emoji == "" && content != nil {
		emoji = normalizeReactionKey(content.RelatesTo.Key)
	}
	targetEventID := id.EventID("")
	if msg != nil && msg.TargetMessage != nil && msg.TargetMessage.MXID != "" {
		targetEventID = msg.TargetMessage.MXID
	} else if content != nil && content.RelatesTo.EventID != "" {
		targetEventID = content.RelatesTo.EventID
	}
	return ReactionContext{Emoji: emoji, TargetEventID: targetEventID}
}
