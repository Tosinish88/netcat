package netcat

import (
	"fmt"
	"log"
	"net"
	ncc "netcat/client"
	ncl "netcat/logger"
	ncs "netcat/struct"
	"sync"
	"time"
)

// StartServer starts the server
func StartServer(port string) {
	var wg sync.WaitGroup
	// listening on the tcp server
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Unable to start the server: %s", err)
		return
	}
	defer listener.Close()
	fmt.Printf("Listening for connections on %s\n", port)
	ncl.CreateNewLogger()
	// listening for incoming connections as long as the server is running
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Unable to accept the client: %s", err)
			continue
		}
		// fmt.Printf("%s has joined the chat", conn.RemoteAddr().String())
		
		go ncc.ProcessClient(conn, &wg)
		go ncc.Broadcast(conn)
		wg.Wait()
		time.Sleep(300 * time.Millisecond)
		ncs.Connections = append(ncs.Connections, conn)

	}
}
