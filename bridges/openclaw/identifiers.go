package openclaw

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"maunium.net/go/mautrix/bridgev2/networkid"

	"github.com/beeper/agentremote/pkg/shared/openclawconv"
)

var (
	openClawValidAgentIDRe   = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
	openClawInvalidAgentIDRe = regexp.MustCompile(`[^a-z0-9_-]+`)
)

func openClawGatewayID(gatewayURL, label string) string {
	key := strings.ToLower(strings.TrimSpace(gatewayURL)) + "|" + strings.ToLower(strings.TrimSpace(label))
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:8])
}

func openClawPortalKey(loginID networkid.UserLoginID, gatewayID, sessionKey string) networkid.PortalKey {
	return networkid.PortalKey{
		ID: networkid.PortalID(
			"openclaw:" +
				string(loginID) + ":" +
				url.PathEscape(strings.TrimSpace(gatewayID)) + ":" +
				url.PathEscape(strings.TrimSpace(sessionKey)),
		),
		Receiver: loginID,
	}
}

func openClawGhostUserID(agentID string) networkid.UserID {
	trimmed := canonicalOpenClawAgentID(agentID)
	if trimmed == "" {
		trimmed = "gateway"
	}
	return networkid.UserID("openclaw-agent:" + url.PathEscape(trimmed))
}

func parseOpenClawGhostID(ghostID string) (string, bool) {
	suffix, ok := strings.CutPrefix(strings.TrimSpace(ghostID), "openclaw-agent:")
	if !ok {
		return "", false
	}
	value, err := url.PathUnescape(suffix)
	if err != nil {
		return "", false
	}
	value = canonicalOpenClawAgentID(value)
	if value == "" {
		return "", false
	}
	return value, true
}

func openClawAgentIDFromSessionKey(sessionKey string) string {
	return openclawconv.AgentIDFromSessionKey(sessionKey)
}

func openClawDMAgentSessionKey(agentID string) string {
	agentID = canonicalOpenClawAgentID(agentID)
	if agentID == "" {
		agentID = "gateway"
	}
	return fmt.Sprintf("agent:%s:matrix-dm", agentID)
}

func isOpenClawSyntheticDMSessionKey(sessionKey string) bool {
	sessionKey = strings.ToLower(strings.TrimSpace(sessionKey))
	if !strings.HasSuffix(sessionKey, ":matrix-dm") {
		return false
	}
	return openClawAgentIDFromSessionKey(sessionKey) != ""
}

func canonicalOpenClawAgentID(agentID string) string {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return ""
	}
	if openClawValidAgentIDRe.MatchString(agentID) {
		return strings.ToLower(agentID)
	}
	normalized := strings.ToLower(agentID)
	normalized = openClawInvalidAgentIDRe.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	if len(normalized) > 64 {
		normalized = normalized[:64]
	}
	return normalized
}
