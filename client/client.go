package netcat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	nc "netcat/chatlogs"
	"os"
	"strings"
	"time"
)

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

// returns the notification message using notification struct
func newNotification(text string, conn net.Conn) notification {
	addr := conn.RemoteAddr().String()
	return notification{text, addr}
}

// returns the message using message struct
func newMessage(text string, conn net.Conn) message {
	msgtime := time.Now().Format("[2006-01-02 15:04:05]")

	addr := conn.RemoteAddr().String()
	return message{msgtime, addr, text}
}

func loadChatHistory(conn net.Conn) {
	fmt.Fprintln(conn, "Loading chat history...")
	for _, v := range nc.Chathistory {
		fmt.Fprintln(conn, v)
	}
	fmt.Fprintln(conn, "You can begin your chat now...")
}

// function to process the client
// it takes the connection as an argument
// prints the linux logo
// gets the name of the client
// sends the welcome message to all clients except the one who joined
// broadcast messages between clients
// sends the leaving message to all clients except the one who left
func ProcessClient(conn net.Conn) {
	time := time.Now().Format("[2006-01-02 15:04:05]")
	// printing the linux logo
	printLinux(conn)

	// getting the name of the client
	name, err := getName(conn)
	if err != nil {
		// checking what name we get if client quits before entering name
		fmt.Println("client name in case of EOF error", name)
		log.Println(err)
		delete(clients, name)
	}

	// load chat history and display to this client who has entered his name
	loadChatHistory(conn)

	// sending notification to all clients that a new client has joined
	welcome <- newNotification(name+" has joined our chat...", conn)

	//reading client messages using new scanner
	input := bufio.NewScanner(conn)
	for input.Scan() {

		text := input.Text()
		if text == "" {
			continue
		}

		// new message send the new message to the message channel to be received in broadcast
		messages <- newMessage(time+"["+name+"]:"+text, conn)
		fmt.Println("I got here to write the message") //- for debugging

	}
	// sending notification to all clients that a client has left
	leaving <- newNotification(name+" has left our chat...", conn)
	// deleting the client from the map
	delete(clients, name)
	// closing the connection
	conn.Close()

}

// combining the broadcast into one function
func Broadcast(conn net.Conn) {
	for {
		select {
		case msg := <-welcome:
			for _, client := range clients {
				if msg.addr != client.conn.RemoteAddr().String() {
					fmt.Fprintln(client.conn, msg.text)

				}
			}
			nc.AddHistory(msg.text)
			log.Println(msg.text)
		case msg := <-messages:
			for _, client := range clients {
				if msg.senderaddr != client.conn.RemoteAddr().String() {
					fmt.Fprintln(client.conn, msg.text)

				} else if msg.senderaddr == client.conn.RemoteAddr().String() {
					fmt.Fprintln(conn, "\033[1A\033[2K"+msg.text)

				}
			}
			nc.AddHistory(msg.text)
			log.Println(msg.text)
		case msg := <-leaving:
			for _, client := range clients {
				if msg.addr != client.conn.RemoteAddr().String() {
					fmt.Fprintln(client.conn, msg.text)

				}
			}
			nc.AddHistory(msg.text)
			log.Println(msg.text)
		}
	}
}

// function prints the linux logo
func printLinux(conn net.Conn) {
	f, err := os.Open("linux.txt")
	defer func() {
		err := f.Close()
		if err != nil {
			log.Println("unable to close the file named", f, err)
		}
	}()
	if err != nil {
		log.Fatalln(err)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		conn.Write([]byte(scanner.Text()))
		conn.Write([]byte("\n"))
	}
}

// function to get the name of the client
func getName(conn net.Conn) (string, error) {
	for {
		fmt.Fprint(conn, "Enter your name: ")
		name, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			return name, err
		}
		name = strings.Trim(name, "\r\n")
		if name == "" {
			continue
		}
		clients[name] = client{conn, name}
		return name, nil
	}
}

// broadcating welcome message to all clients except the one who joined
// func clientBroadcast(conn net.Conn) {
// 	fmt.Println("conn is ", conn) //- for debugging
// 	for {
// 		msg := <-welcome
// 		fmt.Println("message address is ", msg.addr) //- for debugging
// 		for _, client := range clients {
// 			fmt.Println("client is", client)                    //- for debugging
// 			fmt.Println("client is connected on ", client.conn) //- for debugging
// 			if msg.addr != client.conn.RemoteAddr().String() {
// 				fmt.Println("message address is ", msg.addr)                         //- for debugging
// 				fmt.Println("client address is ", client.conn.RemoteAddr().String()) //- for debugging
// 				fmt.Fprintln(client.conn, msg.text)                                  //- for debugging
// 			}
// 		}
// 	}
// }
