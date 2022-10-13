package netcat

import "net"

var Connections []net.Conn

//client type

type Client struct {
	Conn net.Conn
	Name string
}

type Notification struct {
	Text string
	Addr string
}

type Message struct {
	Time       string
	Senderaddr string
	Text       string
}

// creating a map of client with name as key and connection as value
var Clients = make(map[string]Client, 10)

// channels for communicating messages
var Welcome = make(chan Notification)
var Leaving = make(chan Notification)
var Messages = make(chan Message)
var ArrOfconnections = []net.Conn{}
