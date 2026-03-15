package ai

import (
	"context"

	"github.com/beeper/agentremote/pkg/agents"
	"github.com/beeper/agentremote/pkg/agents/tools"
)

func (oc *AIClient) filterEnabledTools(meta *PortalMetadata, allTools []*tools.Tool) []*tools.Tool {
	var enabled []*tools.Tool
	for _, tool := range allTools {
		if oc.isToolEnabled(meta, tool.Name) {
			enabled = append(enabled, tool)
		}
	}
	return enabled
}

func (oc *AIClient) selectedStreamingToolDescriptors(
	ctx context.Context,
	meta *PortalMetadata,
	allowResolvedBossAgent bool,
) []openAIToolDescriptor {
	if meta != nil && !oc.getModelCapabilitiesForMeta(meta).SupportsToolCalling {
		return nil
	}

	var descriptors []openAIToolDescriptor
	builtinTools := oc.selectedBuiltinToolsForTurn(ctx, meta)
	if len(builtinTools) > 0 {
		descriptors = append(descriptors, toolDescriptorsFromDefinitions(builtinTools, &oc.log)...)
	}

	if meta == nil {
		return descriptors
	}

	agentID := resolveAgentID(meta)
	isBossRoom := hasBossAgent(meta) || (allowResolvedBossAgent && agents.IsBossAgent(agentID))
	if isBossRoom {
		descriptors = append(descriptors, toolDescriptorsFromBossTools(oc.filterEnabledTools(meta, tools.BossTools()), &oc.log)...)
		return descriptors
	}

	if agentID == "" {
		return descriptors
	}

	descriptors = append(descriptors, toolDescriptorsFromBossTools(oc.filterEnabledTools(meta, tools.SessionTools()), &oc.log)...)
	return descriptors
}
