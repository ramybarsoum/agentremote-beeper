package agentremote

import (
	"context"

	"github.com/rs/zerolog"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/id"

	"github.com/beeper/agentremote/turns"
)

// ResolveStreamTargetEventID resolves a Matrix event ID for a stream target and
// optionally stores the result in bridge-specific state.
func ResolveStreamTargetEventID(
	ctx context.Context,
	bridge *bridgev2.Bridge,
	receiver networkid.UserLoginID,
	target turns.StreamTarget,
	cached id.EventID,
	cache func(id.EventID),
) (id.EventID, error) {
	if cached != "" {
		return cached, nil
	}
	if bridge == nil {
		return "", nil
	}
	eventID, err := turns.ResolveTargetEventIDFromDB(ctx, bridge, receiver, target)
	if err == nil && eventID != "" && cache != nil {
		cache(eventID)
	}
	return eventID, err
}

// UpdateExistingMessageMetadata updates metadata for an existing assistant
// message, resolving it by network message ID first and then by Matrix event ID.
func UpdateExistingMessageMetadata(
	ctx context.Context,
	login *bridgev2.UserLogin,
	portal *bridgev2.Portal,
	networkMessageID networkid.MessageID,
	initialEventID id.EventID,
	metadata any,
	logger *zerolog.Logger,
	loadErrorMsg string,
	updateErrorMsg string,
) {
	if login == nil || login.Bridge == nil || login.Bridge.DB == nil || portal == nil || metadata == nil {
		return
	}
	log := logger
	if log == nil {
		nop := zerolog.Nop()
		log = &nop
	}
	existing, errByID, errByMXID := findExistingMessage(ctx, login, portal, networkMessageID, initialEventID)
	loadErr := errByID
	if loadErr == nil {
		loadErr = errByMXID
	}
	if loadErr != nil {
		log.Warn().
			Err(loadErr).
			Str("network_message_id", string(networkMessageID)).
			Stringer("initial_event_id", initialEventID).
			Msg(loadErrorMsg)
		return
	}
	if existing == nil {
		return
	}
	existing.Metadata = metadata
	if err := login.Bridge.DB.Message.Update(ctx, existing); err != nil {
		log.Warn().
			Err(err).
			Str("network_message_id", string(networkMessageID)).
			Stringer("initial_event_id", initialEventID).
			Msg(updateErrorMsg)
	}
}
