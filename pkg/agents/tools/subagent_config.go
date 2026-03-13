package tools

import "github.com/beeper/agentremote/pkg/agents/agentconfig"

// SubagentConfig is an alias for the shared type to preserve API compatibility.
type SubagentConfig = agentconfig.SubagentConfig

// cloneSubagentConfig delegates to the shared implementation.
var cloneSubagentConfig = agentconfig.CloneSubagentConfig
