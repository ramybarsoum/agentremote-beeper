package ai

import (
	"context"

	"github.com/beeper/agentremote/pkg/agents"
	"github.com/beeper/agentremote/pkg/shared/stringutil"
	bridgesdk "github.com/beeper/agentremote/sdk"
)

func (oc *AIClient) sdkAgentForDefinition(ctx context.Context, agent *agents.AgentDefinition) *bridgesdk.Agent {
	if agent == nil {
		return nil
	}
	displayName := oc.resolveAgentDisplayName(ctx, agent)
	if displayName == "" {
		displayName = agent.Name
	}
	if displayName == "" {
		displayName = agent.ID
	}
	modelID := oc.agentDefaultModel(agent)
	return &bridgesdk.Agent{
		ID:           string(oc.agentUserID(agent.ID)),
		Name:         displayName,
		Description:  agent.Description,
		Identifiers:  stringutil.DedupeStrings(agentContactIdentifiers(agent.ID, modelID, oc.findModelInfo(modelID))),
		ModelKey:     modelID,
		Capabilities: bridgesdk.MultimodalAgentCapabilities(),
	}
}
