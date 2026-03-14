package tools

import "github.com/beeper/agentremote/pkg/shared/toolspec"

var CronTool = newBuiltinTool(toolspec.CronName, toolspec.CronDescription, "Scheduler", toolspec.CronSchema(), GroupOpenClaw, nil)
