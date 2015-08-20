package proxy
import (
	"log"
	"sync"
	"net"
	"time"
	"io"
	"fmt"
	"github.com/brandnetworks/tcpproxy/backends"
)

type Connection struct {
	config  backends.ConnectionConfig
	channel chan bool
}

func CreateConnection(configuration backends.ConnectionConfig) Connection {
	return Connection{
		config:  configuration,
		channel: make(chan bool),
	}
}

func Listen(logLevel int, localAddr string, remoteAddr string, kill chan bool) error {
	local, err := net.Listen("tcp", localAddr)

	if err != nil {
		log.Fatal("Error atempting to establish connection", err)
		return err
	}

	defer local.Close()

	for {
		select {

		case die, ok := <-kill:
			if !ok || die {
				break
			}

		default:
			conn, err := local.Accept()

			if err != nil {
				return err
			}

			go forward(logLevel, conn, remoteAddr)
		}
	}

	return nil
}

func forward(logLevel int, local net.Conn, remoteAddr string) error {
	if logLevel > 0 {
		log.Printf("Connecting to on %s", remoteAddr)
	}
	remote, err := net.DialTimeout("tcp", remoteAddr, 1*time.Minute)
	if err != nil {
		local.Close()
		return err
	}

	proxyTCP(logLevel, local.(*net.TCPConn), remote.(*net.TCPConn))
	return nil
}

// proxyTCP proxies data bi-directionally between in and out.
func proxyTCP(logLevel int, in, out *net.TCPConn) {
	var wg sync.WaitGroup
	wg.Add(2)

	if logLevel > 0 {
		log.Printf("Creating proxy between %v <-> %v <-> %v <-> %v",
			in.RemoteAddr(), in.LocalAddr(), out.LocalAddr(), out.RemoteAddr())
	}

	go copyBytes(logLevel, "from backend", in, out, &wg)
	go copyBytes(logLevel, "to backend", out, in, &wg)
	wg.Wait()
	in.Close()
	out.Close()
}

func copyBytes(logLevel int, direction string, dest, src *net.TCPConn, wg *sync.WaitGroup) {
	defer wg.Done()
	if logLevel > 0 {
		log.Printf("Copying %s: %s -> %s", direction, src.RemoteAddr(), dest.RemoteAddr())
	}
	n, err := io.Copy(dest, src)
	if err != nil {
		log.Printf("I/O error: %v", err)
	}
	if logLevel > 0 {
		log.Printf("Copied %d bytes %s: %s -> %s", n, direction, src.RemoteAddr(), dest.RemoteAddr())
	}
	dest.CloseWrite()
	src.CloseRead()
}

func RunTcpProxy(logLevel int, createChannel chan []Connection, killChannel chan []Connection, cb func()) {

	quit := make(chan struct{})

	go func() {
		for {

			select {
			case toKill, ok := <-killChannel:
				if ok {
					// Kill those connections
					for i := range toKill {
						close(toKill[i].channel)

						if logLevel > 0 {
							log.Printf("No longer listening on %s", toKill[i])
						}
					}
				} else {
					fmt.Errorf("Failed to read from kill channel.")
					panic("Couldnt read from toKill in RunTcpProxy")
				}

			case toCreate, ok := <-createChannel:
				if ok {
					// Create those connections
					for i := range toCreate {
						go Listen(logLevel, toCreate[i].config.LocalAddress, toCreate[i].config.RemoteAddress, toCreate[i].channel)

						if logLevel > 0 {
							log.Printf("Listening on %s", toCreate[i].config.LocalAddress)
						}
					}
				} else {
					fmt.Errorf("Failed to read from create channel.")
					panic("Couldnt read from toCreate in RunTcpProxy")
				}

			case <-quit:
				return
			}
		}
	}()

	cb()

	quit <- struct{}{}
}