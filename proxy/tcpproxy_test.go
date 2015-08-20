package proxy

import (
	"net"
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/brandnetworks/tcpproxy/backends"
)

func TestParseConnectionsParameter(t *testing.T) {
	fmt.Println("Testing TestParseConnectionsParameter")
	connectionsConfig, err := backends.ParseConnectionsParameter("1234:localhost:4567")
	if err != nil {
		t.Error(err)
	}

	name := connectionsConfig[0].Name
	localAddr := connectionsConfig[0].LocalAddress
	remoteAddr := connectionsConfig[0].RemoteAddress

	assert.Equal(t, localAddr, ":1234", "LocalAddress is not the expected one")
	assert.Equal(t, name, "localhost", "Name is not the expected one")
	assert.Equal(t, remoteAddr, "localhost:4567", "RemoteAddress is not the expected one")
}

func echoServer(t *testing.T, quit chan bool) {

	l, err := net.Listen("tcp", ":11111")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	go func() {
		c, err := l.Accept()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintln(c, "OK")
		c.Close()
	}()
	for {
		select {
		case <- quit:
			l.Close()
			return
		}
	}
}

func TestListen(t *testing.T) {
	fmt.Println("Testing TestListen")

	quit := make(chan bool)

	go echoServer(t, quit)
	go Listen(1, ":11110", "localhost:11111", quit)

	conn, err := net.Dial("tcp", "localhost:11110")
	if err != nil {
		t.Fatal(err)
	}

	var cmd []byte
	fmt.Fscan(conn, &cmd)
	t.Log("Message:", string(cmd))

	quit <- true
}

func TestRunProxy(t *testing.T) {
	fmt.Println("Testing TestRunProxy")

	quit := make(chan bool)
	create := make(chan []Connection)
	kill := make(chan []Connection)

	go echoServer(t, quit)

	connectionsConfig, err := backends.ParseConnectionsParameter("11112:localhost:11113")
	if err != nil {
		t.Error(err)
	}

	connections := make([]Connection, len(connectionsConfig))
	for i := range connectionsConfig {
		connections[i] = CreateConnection(connectionsConfig[i])
	}

	go RunTcpProxy(1, create, kill, func() {
		create <-connections

		conn, err := net.Dial("tcp", "localhost:11112")
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		var cmd []byte
		fmt.Fscan(conn, &cmd)
		t.Log("Message:", string(cmd))

		quit <- true
	})

}
