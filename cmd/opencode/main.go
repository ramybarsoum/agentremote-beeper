package main

import (
	"maunium.net/go/mautrix/bridgev2/matrix/mxmain"

	"github.com/beeper/agentremote/bridges/opencode"
)

var (
	Tag       = "unknown"
	Commit    = "unknown"
	BuildTime = "unknown"
)

var m = mxmain.BridgeMain{
	Name:        "opencode",
	Description: "A Matrix↔OpenCode bridge built on mautrix-go bridgev2.",
	URL:         "https://github.com/beeper/agentremote",
	Version:     "0.1.0",
	Connector:   opencode.NewConnector(),
}

func main() {
	m.InitVersion(Tag, Commit, BuildTime)
	m.Run()
}
