package netcat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	nc "netcat/chatlogs"
	"os"
	"strings"
	"sync"
	"time"
)
// green is when a new user joins
// blue is that you receive
// white is message that you send
// red is when a user leaves

var Reset  = "\033[0m"
var Red    = "\033[31m"
var Green  = "\033[32m"
var Blue   = "\033[34m"
var White = "\033[37m"


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
	time := time.Now().String()[0:19]

	addr := conn.RemoteAddr().String()
	return Message{time, addr, text}
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
func ProcessClient(conn net.Conn, wg *sync.WaitGroup) {

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
	Welcome <- newNotification(Green+name+" has joined our chat..."+ Reset, conn)

	//reading client messages using new scanner
	input := bufio.NewScanner(conn)
	for input.Scan() {

		text := input.Text()
		if text == "" {
			continue
		}
		wg.Add(1)
		time := time.Now().String()[0:19]
		// new message send the new message to the message channel to be received in broadcast
		fmt.Fprintln(conn, "\033[1A\033[K"+"["+time+"]"+"["+name+"]:"+text)
		Messages <- newMessage(Blue+"["+time+"]"+"["+name+"]:"+text + Reset, conn)
		wg.Done()

	}
	// sending notification to all clients that a client has left
	Leaving <- newNotification(Red+name+" has left our chat..."+ Reset, conn)
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
