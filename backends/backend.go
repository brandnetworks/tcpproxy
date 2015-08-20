package backends
import (

	"fmt"
	"strings"
)

type ConnectionConfig struct {
	Name          string
	LocalAddress  string   "local_address"
	RemoteAddress string   "remote_address"
	Url           string
}

type ReadWrite interface {
	CreateProxyConfiguration(proxy_configuration string) error
	DeleteProxyConfiguration(proxy_configuration string) error
	GetProxyConfigurations() ([]ConnectionConfig, error)
}

type ReadOnly interface {
	GetProxyConfigurations() ([]ConnectionConfig, error)
	IsPollable() bool
}

func ParseConnectionsParameter(connectionsArg string) ([]ConnectionConfig, error) {
	if len(connectionsArg) == 0 {
		return nil, fmt.Errorf("Connection must not be empty")
	}

	connections := strings.Split(connectionsArg, ",");
	connectionsConfig := make([]ConnectionConfig, len(connections))

	for i := range connections {
		config, err := ParseConnection(connections[i])

		if err != nil {
			fmt.Errorf("Error parsing connection %s", connections[i])
		} else {
			connectionsConfig[i] = *config
		}
	}

	return connectionsConfig, nil
}

func ParseConnection(connection string) (*ConnectionConfig, error) {
	connectionParts := strings.Split(connection, ":")
	if len(connectionParts) != 3 {
		return nil, fmt.Errorf("A connection must have three parts: srcPort:destHost:destPort '", connection, "'")
	}

	config := ConnectionConfig{
		Name: connectionParts[1],
		LocalAddress: ":" + connectionParts[0],
		RemoteAddress: connectionParts[1] + ":" + connectionParts[2],
		Url: connection,
	}

	return &config, nil
}