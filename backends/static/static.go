package static

import (
	"github.com/brandnetworks/tcpproxy/backends"
)

func CreateStaticBackend(connectionsArgument string) (*StaticBackend, error) {
	connections, err := backends.ParseConnectionsParameter(connectionsArgument)

	if err != nil {
		return nil, err
	} else {
		return &StaticBackend{ connectionsConfig: connections }, nil
	}
}

type StaticBackend struct {
	connectionsConfig []backends.ConnectionConfig
}

func (b *StaticBackend) GetProxyConfigurations() ([]backends.ConnectionConfig, error) {
	return b.connectionsConfig, nil
}

func (b *StaticBackend) IsPollable() bool {
	return false
}
