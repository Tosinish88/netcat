package main

import (
	"fmt"
	netcat "netcat/server"
	"os"
)

func main() {
	Args := os.Args[1:]
	if len(Args) == 0 {
		port := ":8989"
		netcat.StartServer(port)
		return
	} else if len(Args) == 1 {
		port := ":"
		port += Args[0]
		netcat.StartServer(port)
		return
	} else {
		fmt.Println("[USAGE]: ./TCPChat $port")
		fmt.Println()
		return
	}

}
