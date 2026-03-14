package tools

import "github.com/beeper/agentremote/pkg/shared/toolspec"

// AgentsListTool lists agent ids allowed for sessions_spawn.
var AgentsListTool = newBuiltinTool(
	"agents_list",
	"List agent ids you can target with sessions_spawn (based on allowlists).",
	"Agents List",
	toolspec.EmptyObjectSchema(),
	GroupSessions,
	nil,
)
