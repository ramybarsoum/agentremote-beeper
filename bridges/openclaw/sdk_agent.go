package openclaw

import (
	"strings"

	bridgesdk "github.com/beeper/agentremote/sdk"
)

func (oc *OpenClawClient) sdkAgentForProfile(profile openClawAgentProfile) *bridgesdk.Agent {
	displayName := oc.displayNameFromAgentProfile(profile)
	agentID := strings.TrimSpace(profile.AgentID)
	return &bridgesdk.Agent{
		ID:          string(openClawGhostUserID(agentID)),
		Name:        displayName,
		Description: "OpenClaw agent",
		AvatarURL:   profile.AvatarURL,
		Identifiers: oc.configuredAgentIdentifiers(agentID),
		ModelKey:    agentID,
		Capabilities: bridgesdk.AgentCapabilities{
			SupportsStreaming:   true,
			SupportsReasoning:   true,
			SupportsToolCalling: true,
			SupportsTextInput:   true,
			SupportsFilesOutput: true,
			MaxTextLength:       100000,
		},
	}
}
