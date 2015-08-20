package proxy

import (
	"time"
	"github.com/brandnetworks/tcpproxy/backends"
	"log"
)

type Proxy struct {
	LiveConnections map[string]Connection
	CreateChannel   chan []Connection
	KillChannel     chan []Connection
	Backend         backends.ReadOnly
}

func RunProxy(backend backends.ReadOnly, logLevel int, callback func(c *Proxy)) error {
	proxy := &Proxy{
		LiveConnections: make(map[string]Connection),
		CreateChannel: make(chan []Connection, 1),
		KillChannel: make(chan []Connection, 1),
		Backend: backend,
	}

	return proxy.Run(logLevel, func() {
		callback(proxy)
	})
}

func diffProxies(logLevel int, newProxyList []backends.ConnectionConfig, live map[string]Connection) ([]Connection, []Connection, map[string]Connection, error) {

	// This is essentially set difference :/

	// This is the map of all connections to create/retain
	connectionsConfigMap := make(map[string]backends.ConnectionConfig)

	toCreate := make([]Connection, 0)
	toRetain := make([]Connection, 0)
	toKill := make([]Connection, 0)

	// Found out which connections to retain or create
	for i := range newProxyList {
		connectionsConfigMap[newProxyList[i].Url] = newProxyList[i]

		if _, ok := live[newProxyList[i].Url]; ok {
			toRetain = append(toRetain, Connection{config: newProxyList[i], channel: live[newProxyList[i].Url].channel})
		} else {
			toCreate = append(toCreate, Connection{config: newProxyList[i], channel: nil})
		}
	}

	// Find out which connections to kill
	for url, connection := range live {
		if _, ok := connectionsConfigMap[url]; !ok {
			toKill = append(toKill, connection)
		}
	}

	newLive := make(map[string]Connection)

	for i := range toRetain {
		newLive[toRetain[i].config.Url] = toRetain[i]
	}

	for i := range toCreate {
		toCreate[i].channel = make(chan bool)

		newLive[toCreate[i].config.Url] = toCreate[i]
	}

	if logLevel > 0 {
		if logLevel > 1 {
			log.Println("Connections", newProxyList)
		}

		log.Println("Retain", toRetain)
		log.Println("Create", toCreate)
		log.Println("Kill  ", toKill)
	}

	return toCreate, toKill, newLive, nil
}

func (c *Proxy) UpdateConnections(logLevel int) error {
	connections, err := c.Backend.GetProxyConfigurations()

	if logLevel > 1 {
		log.Println("Got connections...")
		log.Println("Live", c.LiveConnections)
	}

	if err != nil {
		log.Println("Error fetching proxies from backend %", err)
		return err
	} else {
		var toCreate []Connection
		var toKill   []Connection

		toCreate, toKill, live, err := diffProxies(logLevel, connections, c.LiveConnections)
		c.LiveConnections = live

		if err != nil {
			return err
		}

		if logLevel > 2 {
			log.Println("live", c.LiveConnections)
		}

		c.KillChannel <- toKill
		c.CreateChannel <- toCreate
	}

	return nil
}

func (c *Proxy) Run(logLevel int, callback func()) error {

	err := c.UpdateConnections(logLevel)
	if err != nil {
		return err
	}

	quit := make(chan struct {})

	if c.Backend.IsPollable() {

		go func() {
			timer := time.NewTicker(time.Minute)
			for {
				select {
				case <-quit:
					return
				case <-timer.C:
					err := c.UpdateConnections(logLevel)
					if err != nil {
						log.Println("Error Updating Connections ", err)
					}
				}

			}
		}()
	}

	callback()

	close(quit)

	return nil

}
