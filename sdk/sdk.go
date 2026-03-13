package sdk

import (
	"maunium.net/go/mautrix/bridgev2/matrix/mxmain"
)

// Bridge is the SDK bridge handle.
type Bridge struct {
	config    *Config
	connector *sdkConnector
	main      *mxmain.BridgeMain
}

// New creates a new SDK bridge instance.
func New(cfg Config) *Bridge {
	conn := newSDKConnector(&cfg)

	port := cfg.Port
	if port == 0 {
		port = 29400
	}
	dbName := cfg.DBName
	if dbName == "" {
		dbName = cfg.Name + ".db"
	}
	desc := cfg.Description
	if desc == "" {
		desc = "A Matrix↔" + cfg.Name + " bridge for Beeper built on agentremote SDK."
	}

	m := &mxmain.BridgeMain{
		Name:        cfg.Name,
		Description: desc,
		URL:         "https://github.com/beeper/agentremote",
		Version:     "0.1.0",
		Connector:   conn,
	}

	return &Bridge{
		config:    &cfg,
		connector: conn,
		main:      m,
	}
}

// Run starts the bridge and blocks until it exits.
func (b *Bridge) Run() {
	b.main.InitVersion("0.1.0", "unknown", "unknown")
	b.main.Run()
}

// Stop stops the bridge.
func (b *Bridge) Stop() {
	// Bridge stop is handled by mxmain's signal handling
}

// Connector returns the underlying ConnectorBase.
func (b *Bridge) Connector() *sdkConnector { return b.connector }

// BridgeMain returns the underlying mxmain.BridgeMain.
func (b *Bridge) BridgeMain() *mxmain.BridgeMain { return b.main }
