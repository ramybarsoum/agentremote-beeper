package opencode

import bridgesdk "github.com/beeper/agentremote/sdk"

func openCodeSDKAgent(instanceID, displayName string) *bridgesdk.Agent {
	if displayName == "" {
		displayName = "OpenCode"
	}
	return &bridgesdk.Agent{
		ID:          string(OpenCodeUserID(instanceID)),
		Name:        displayName,
		Description: "OpenCode instance",
		Identifiers: []string{"opencode:" + instanceID},
		ModelKey:    "opencode:" + instanceID,
		Capabilities: bridgesdk.AgentCapabilities{
			SupportsStreaming:   true,
			SupportsReasoning:   true,
			SupportsToolCalling: true,
			SupportsTextInput:   true,
			SupportsImageInput:  true,
			SupportsAudioInput:  true,
			SupportsVideoInput:  true,
			SupportsFileInput:   true,
			SupportsPDFInput:    true,
			SupportsFilesOutput: true,
			MaxTextLength:       100000,
		},
	}
}
