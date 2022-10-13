package netcat

import "net"

var Connections []net.Conn

//client type

type client struct {
	conn net.Conn
	name string
}

type notification struct {
	text string
	addr string
}

type message struct {
	time       string
	senderaddr string
	text       string
}

// creating a map of client with name as key and connection as value
var clients = make(map[string]client, 10)

// channels for communicating messages
var welcome = make(chan notification)
var leaving = make(chan notification)
var messages = make(chan message)
var ArrOfconnections = []net.Conn{}
