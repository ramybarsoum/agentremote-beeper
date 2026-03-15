package codex

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/event"

	"github.com/beeper/agentremote"
	bridgesdk "github.com/beeper/agentremote/sdk"
)

func isWelcomeCodexPortal(meta *PortalMetadata) bool {
	return meta != nil && meta.IsCodexRoom && meta.AwaitingCwdSetup
}

func codexTopicForPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	return fmt.Sprintf("Working directory: %s", path)
}

func codexTitleForPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "Codex"
	}
	base := strings.TrimSpace(filepath.Base(path))
	switch base {
	case "", ".", string(filepath.Separator):
		return path
	default:
		return base
	}
}

func (cc *CodexClient) codexTopicForPortal(_ *bridgev2.Portal, meta *PortalMetadata) string {
	if meta == nil || isWelcomeCodexPortal(meta) {
		return ""
	}
	return codexTopicForPath(meta.CodexCwd)
}

func (cc *CodexClient) setRoomName(ctx context.Context, portal *bridgev2.Portal, name string) error {
	if cc == nil || cc.UserLogin == nil || cc.UserLogin.Bridge == nil || portal == nil {
		return fmt.Errorf("portal unavailable")
	}
	if portal.MXID == "" {
		return fmt.Errorf("portal has no Matrix room ID")
	}
	_, err := cc.UserLogin.Bridge.Bot.SendState(ctx, portal.MXID, event.StateRoomName, "", &event.Content{
		Parsed: &event.RoomNameEventContent{Name: name},
	}, time.Time{})
	if err != nil {
		return fmt.Errorf("failed to set room name: %w", err)
	}
	portal.Name = name
	portal.NameSet = true
	return portal.Save(ctx)
}

func (cc *CodexClient) setRoomTopic(ctx context.Context, portal *bridgev2.Portal, topic string) error {
	if cc == nil || cc.UserLogin == nil || cc.UserLogin.Bridge == nil || portal == nil {
		return fmt.Errorf("portal unavailable")
	}
	if portal.MXID == "" {
		return fmt.Errorf("portal has no Matrix room ID")
	}
	_, err := cc.UserLogin.Bridge.Bot.SendState(ctx, portal.MXID, event.StateTopic, "", &event.Content{
		Parsed: &event.TopicEventContent{Topic: topic},
	}, time.Time{})
	if err != nil {
		return fmt.Errorf("failed to set room topic: %w", err)
	}
	portal.Topic = topic
	return portal.Save(ctx)
}

func (cc *CodexClient) syncCodexRoomTopic(ctx context.Context, portal *bridgev2.Portal, meta *PortalMetadata) {
	if cc == nil || portal == nil || meta == nil {
		return
	}
	want := cc.codexTopicForPortal(portal, meta)
	if strings.TrimSpace(portal.Topic) == strings.TrimSpace(want) {
		return
	}
	if err := cc.setRoomTopic(ctx, portal, want); err != nil {
		cc.log.Warn().Err(err).Stringer("room", portal.MXID).Msg("Failed to sync Codex room topic")
	}
}

func (cc *CodexClient) welcomeCodexPortals(ctx context.Context) ([]*bridgev2.Portal, error) {
	if cc == nil || cc.UserLogin == nil || cc.UserLogin.Bridge == nil || cc.UserLogin.Bridge.DB == nil {
		return nil, nil
	}
	userPortals, err := cc.UserLogin.Bridge.DB.UserPortal.GetAllForLogin(ctx, cc.UserLogin.UserLogin)
	if err != nil {
		return nil, err
	}
	out := make([]*bridgev2.Portal, 0, len(userPortals))
	for _, userPortal := range userPortals {
		if userPortal == nil {
			continue
		}
		portal, err := cc.UserLogin.Bridge.GetExistingPortalByKey(ctx, userPortal.Portal)
		if err != nil || portal == nil {
			continue
		}
		if isWelcomeCodexPortal(portalMeta(portal)) {
			out = append(out, portal)
		}
	}
	return out, nil
}

func (cc *CodexClient) createWelcomeCodexChat(ctx context.Context) (*bridgev2.Portal, error) {
	if cc == nil || cc.UserLogin == nil || cc.UserLogin.Bridge == nil {
		return nil, fmt.Errorf("login unavailable")
	}
	portalKey, err := codexWelcomePortalKey(cc.UserLogin.ID, generateShortID())
	if err != nil {
		return nil, err
	}
	portal, err := cc.UserLogin.Bridge.GetPortalByKey(ctx, portalKey)
	if err != nil {
		return nil, err
	}
	if portal.Metadata == nil {
		portal.Metadata = &PortalMetadata{}
	}
	meta := portalMeta(portal)
	meta.IsCodexRoom = true
	meta.Title = "New Codex Chat"
	meta.Slug = "codex-welcome"
	meta.CodexThreadID = ""
	meta.CodexCwd = ""
	meta.AwaitingCwdSetup = true
	meta.ManagedImport = false
	portal.RoomType = database.RoomTypeDM
	portal.OtherUserID = codexGhostID
	portal.Name = meta.Title
	portal.NameSet = true
	info := cc.composeCodexChatInfo(portal, meta.Title, false)
	created, err := bridgesdk.EnsurePortalLifecycle(ctx, bridgesdk.PortalLifecycleOptions{
		Login:             cc.UserLogin,
		Portal:            portal,
		ChatInfo:          info,
		SaveBeforeCreate:  true,
		AIRoomKind:        agentremote.AIRoomKindAgent,
		ForceCapabilities: true,
	})
	if err != nil {
		return nil, err
	}
	if created {
		cc.sendSystemNotice(ctx, portal, "AI Chats can make mistakes.")
		cc.sendSystemNotice(ctx, portal, "Send an absolute path or `~/...` to start a Codex session.")
	}
	if err := portal.Save(ctx); err != nil {
		return nil, err
	}
	cc.syncCodexRoomTopic(ctx, portal, meta)
	return portal, nil
}

func (cc *CodexClient) ensureWelcomeCodexChat(ctx context.Context) error {
	cc.defaultChatMu.Lock()
	defer cc.defaultChatMu.Unlock()

	portals, err := cc.welcomeCodexPortals(ctx)
	if err != nil {
		return err
	}
	if len(portals) > 0 {
		return nil
	}
	_, err = cc.createWelcomeCodexChat(ctx)
	return err
}

func (cc *CodexClient) handleWelcomeCodexMessage(ctx context.Context, portal *bridgev2.Portal, meta *PortalMetadata, body string) (*bridgev2.MatrixMessageResponse, error) {
	if cc == nil || cc.UserLogin == nil || portal == nil || meta == nil {
		return &bridgev2.MatrixMessageResponse{Pending: false}, nil
	}
	path, err := resolveCodexWorkingDirectory(body)
	if err != nil {
		cc.sendSystemNotice(ctx, portal, "That path must be absolute. `~/...` is also accepted.")
		return &bridgev2.MatrixMessageResponse{Pending: false}, nil
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		cc.sendSystemNotice(ctx, portal, fmt.Sprintf("That path doesn't exist or isn't a directory: %s", path))
		return &bridgev2.MatrixMessageResponse{Pending: false}, nil
	}

	addManagedCodexPath(loginMetadata(cc.UserLogin), path)
	if err := cc.UserLogin.Save(ctx); err != nil {
		return nil, messageSendStatusError(err, "Failed to save Codex directory.", "")
	}

	meta.CodexCwd = path
	meta.CodexThreadID = ""
	meta.AwaitingCwdSetup = false
	meta.ManagedImport = false
	meta.Title = codexTitleForPath(path)
	meta.Slug = strings.ToLower(strings.ReplaceAll(meta.Title, " ", "-"))
	portal.Name = meta.Title
	portal.NameSet = true
	if err := portal.Save(ctx); err != nil {
		return nil, messageSendStatusError(err, "Failed to save Codex room.", "")
	}
	if err := cc.setRoomName(ctx, portal, meta.Title); err != nil {
		return nil, messageSendStatusError(err, "Failed to rename Codex room.", "")
	}
	if err := cc.ensureRPC(cc.backgroundContext(ctx)); err != nil {
		return nil, messageSendStatusError(err, "Codex isn't available. Sign in again.", "")
	}
	if err := cc.ensureCodexThread(ctx, portal, meta); err != nil {
		return nil, messageSendStatusError(err, "Failed to start Codex thread.", "")
	}
	cc.syncCodexRoomTopic(ctx, portal, meta)
	cc.sendSystemNotice(ctx, portal, fmt.Sprintf("Started a new Codex session in %s", path))
	go func() {
		if _, err := cc.createWelcomeCodexChat(cc.backgroundContext(ctx)); err != nil {
			cc.log.Warn().Err(err).Msg("Failed to create follow-up welcome Codex chat")
		}
	}()
	go func() {
		if _, _, err := cc.syncStoredCodexThreadsForPath(cc.backgroundContext(ctx), path); err != nil {
			cc.log.Warn().Err(err).Str("cwd", path).Msg("Failed to sync stored Codex threads for path")
		}
	}()
	return &bridgev2.MatrixMessageResponse{Pending: false}, nil
}
