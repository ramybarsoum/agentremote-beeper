package modules

import (
	integrationcron "github.com/beeper/agentremote/pkg/integrations/cron"
	integrationmemory "github.com/beeper/agentremote/pkg/integrations/memory"
	integrationruntime "github.com/beeper/agentremote/pkg/integrations/runtime"
)

// BuiltinFactories is the compile-time module selection list.
// Removing one import line and one factory line cleanly excludes a module.
var BuiltinFactories = []integrationruntime.ModuleFactory{
	integrationcron.New,
	integrationmemory.New,
}
