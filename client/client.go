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

// returns the notification message using notification struct
func newNotification(text string, conn net.Conn) Notification {
	addr := conn.RemoteAddr().String()
	return Notification{text, addr}
}

// returns the message using message struct
func newMessage(text string, conn net.Conn) Message {
	msgtime := time.Now().Format("[2006-01-02 15:04:05]")

	addr := conn.RemoteAddr().String()
	return Message{msgtime, addr, text}
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
		delete(Clients, name)
	}

	// load chat history and display to this client who has entered his name
	loadChatHistory(conn)

	// sending notification to all clients that a new client has joined
	Welcome <- newNotification(name+" has joined our chat...", conn)

	//reading client messages using new scanner
	input := bufio.NewScanner(conn)
	for input.Scan() {

		text := input.Text()
		if text == "" {
			continue
		}

		// new message send the new message to the message channel to be received in broadcast
		Messages <- newMessage(time+"["+name+"]:"+text, conn)
		fmt.Println("I got here to write the message") //- for debugging

	}
	// sending notification to all clients that a client has left
	Leaving <- newNotification(name+" has left our chat...", conn)
	// deleting the client from the map
	delete(Clients, name)
	// closing the connection
	conn.Close()

}

// combining the broadcast into one function
func Broadcast(conn net.Conn) {
	for {
		select {
		case msg := <-Welcome:
			for _, client := range Clients {
				if msg.Addr != client.Conn.RemoteAddr().String() {
					fmt.Fprintln(client.Conn, msg.Text)

				}
			}
			nc.AddHistory(msg.Text)
			log.Println(msg.Text)
		case msg := <-Messages:
			for _, client := range Clients {
				if msg.Senderaddr != client.Conn.RemoteAddr().String() {
					fmt.Fprintln(client.Conn, msg.Text)

				} else if msg.Senderaddr == client.Conn.RemoteAddr().String() {
					fmt.Fprintln(conn, "\033[1A\033[2K"+msg.Text)

				}
			}
			nc.AddHistory(msg.Text)
			log.Println(msg.Text)
		case msg := <-Leaving:
			for _, client := range Clients {
				if msg.Addr != client.Conn.RemoteAddr().String() {
					fmt.Fprintln(client.Conn, msg.Text)

				}
			}
			nc.AddHistory(msg.Text)
			log.Println(msg.Text)
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
		Clients[name] = Client{conn, name}
		return name, nil
	}
}

