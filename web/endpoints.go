package web
import (
	"net/http"
	"fmt"
	"github.com/brandnetworks/tcpproxy/proxy"
	"log"
	"encoding/json"
)

func InitialiseEndpoints(logLevel int, proxyName string, connectionManager *proxy.Proxy) (*http.ServeMux) {

	mux := http.NewServeMux()

	mux.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	mux.HandleFunc("/connections", func(w http.ResponseWriter, _ *http.Request) {
		if logLevel > 2 {
			log.Println("liveProxyConfigurations", connectionManager.LiveConnections)
		}

		connectionsMap := make(map[string]interface{})

		if proxyName != "" {
			connectionsMap["name"] = proxyName
		}

		if connectionManager.LiveConnections == nil || len(connectionManager.LiveConnections) == 0 {
			connectionsMap["error"] = "No connections!"
		} else {
			connections := make([]string, 0)

			for url, _ := range connectionManager.LiveConnections {
				connections = append(connections, url)
			}

			connectionsMap["connections"] = connections
		}

		out, _ := json.Marshal(connectionsMap)
		fmt.Fprintln(w, string(out))
	})

	return mux
}