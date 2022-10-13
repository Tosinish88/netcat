package netcat

import (
	"fmt"
	"log"
	"net"
	ncs "netcat/struct"
	ncl "netcat/logger"
	ncc "netcat/client"
)

// StartServer starts the server
func StartServer(port string) {

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
		go ncc.ProcessClient(conn)
		go ncc.Broadcast(conn)
		ncs.Connections = append(ncs.Connections, conn)

	}
}
