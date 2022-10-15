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

// green is when a new user joins
// blue is that you receive
// white is message that you send
// red is when a user leaves

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Blue = "\033[34m"
var White = "\033[37m"

var Connections []net.Conn

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

// function loads chat history and displays to the client
func loadChatHistory(conn net.Conn) {
	fmt.Fprintln(conn, "Loading chat history...")
	for _, v := range nc.Chathistory {
		fmt.Fprintln(conn, v)
	}
	fmt.Fprintln(conn, "You can begin your chat now...")
}

// function allow client to update his name
func UpdateName(conn net.Conn) string {
	str := ""
	fmt.Fprintln(conn, "Enter your new name:")
	input := bufio.NewScanner(conn)
	for input.Scan() {
		newname := input.Text()
		// check if the name already exists
		for _, v := range Clients {
			if newname == v.Name {
				fmt.Fprintln(conn, "Name already exists")
				return ""
			}
		}
		// update the name
		for _, v := range Clients {
			if conn.RemoteAddr().String() == v.Conn.RemoteAddr().String() {
				v.Name = newname
				fmt.Fprintln(conn, "Name updated successfully")
				str += newname
			}
		}
	}
	return str
}

// function to process the client
// it takes the connection as an argument
// prints the linux logo
// gets the name of the client
// sends the welcome message to all clients except the one who joined
// broadcast messages between clients
// sends the leaving message to all clients except the one who left
func ProcessClient(conn net.Conn) {
	printLinux(conn)
	name, err := getName(conn)
	if err != nil {
		// checking what name we get if client quits before entering name
		fmt.Println("client name in case of EOF error", name)
		log.Println(err)
		delete(Clients, name)
	}
	// show list of available commands
	fmt.Fprintln(conn, "List of commands:")
	fmt.Fprintln(conn, "/updatename - update your name")
	fmt.Fprintln(conn, "/quit - quit the chat") //-implemented
	fmt.Fprintln(conn, "/exit - exit the chat") //-implemented
	// load chat history and display to this client who has entered his name
	loadChatHistory(conn)
	// sending notification to all clients that a new client has joined
	Welcome <- newNotification(Green+name+" has joined our chat..."+Reset, conn)
	//reading client messages using new scanner
	input := bufio.NewScanner(conn)
	for input.Scan() {
		text := input.Text()
		if text == "" {
			continue
		} else if text == "/updatename" {
			// update the name of the client
			oldname := name
			newname := UpdateName(conn)
			Welcome <- newNotification(Green+oldname+" has updated their name to "+newname+Reset, conn)
			fmt.Println(name)
		} else if text == "/quit" || text == "/exit" {
			// deleting the client from the map
			delete(Clients, name)
			// sending notification to all clients that a client has left
			Leaving <- newNotification(Red+name+" has left our chat..."+Reset, conn)
			// closing the connection
			conn.Close()
			break
		}
		// wg.Add(1)
		time := time.Now().String()[0:19]
		// new message send the new message to the message channel to be received in broadcast
		fmt.Fprintln(conn, "\033[1A\033[K"+"["+time+"]"+"["+name+"]:"+text)
		Messages <- newMessage(Blue+"["+time+"]"+"["+name+"]:"+text+Reset, conn)
		// wg.Done()
	}
	// // taking care of when client leaves without using /quit
	// // deleting the client from the map
	// delete(Clients, name)
	// // sending notification to all clients that a client has left
	// Leaving <- newNotification(Red+name+" has left our chat..."+Reset, conn)
	// // closing the connection
	// conn.Close()
}

// combining the broadcast into one function
func Broadcast() {
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

// type Rooms struct {
// 	Roomname string
// }

// function create chatroom
// func CreateRoom(conn net.Conn, wg *sync.WaitGroup) {
// 	// get the name of the room
// 	fmt.Fprintln(conn, "Enter the name of the room:")
// 	input := bufio.NewScanner(conn)
// 	for input.Scan() {
// 		roomname := input.Text()
// 		// check if the room already exists
// 		for _, v := range ChatRooms {
// 			if v == roomname {
// 				fmt.Fprintln(conn, "Room already exists")
// 				return
// 			}
// 		}
// 		// add the room to the list of rooms
// 		ChatRooms = append(ChatRooms, roomname)
// 		fmt.Println(ChatRooms)
// 		fmt.Fprintln(conn, "Room created successfully")
// 		return
// 	}
// }

// function list all available chatrooms
// func ListRooms(conn net.Conn) {
// 	if len(ChatRooms) == 0 {
// 		fmt.Println(len(ChatRooms))
// 		fmt.Println(ChatRooms)
// 		fmt.Println("before rooms is created")

// 		fmt.Fprintln(conn, "No rooms created yet, Please create a room")
// 		CreateRoom(conn, nil)
// 		fmt.Println(ChatRooms)
// 		fmt.Println(len(ChatRooms))
// 		fmt.Println("after rooms is created")
// 		return
// 	} else if len(ChatRooms) > 0 {
// 		fmt.Println("i got here")
// 		for _, v := range ChatRooms {
// 			fmt.Fprintln(conn, "List of rooms:")
// 			fmt.Fprintln(conn, v)
// 			return
// 		}
// 	}
// }

// fmt.Fprintln(conn, "/createroom - create a new room")
// fmt.Fprintln(conn, "/listrooms - list all available rooms")
// fmt.Fprintln(conn, "/joinroom - join a room")
// fmt.Fprintln(conn, "/leaveroom - leave a room")
