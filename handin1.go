package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

func main() {
	peerAddress := ""
	fmt.Println("Please enter peer IP and port (in the form peerIP:port) to connect")
	fmt.Scanln(&peerAddress)
	d := net.Dialer{Timeout: 60 * time.Second}
	conn, err := d.Dial("tcp", peerAddress)
	if err != nil {
		print("No connection established, starting own network.")
		startOwnNetwork()
	}
	fmt.Printf("Connection to %s established\n", peerAddress)
	appendNetwork(conn)
}

// construct to hold connections slice
type connStruct struct {
	connections []net.Conn
}

func appendNetwork(parent net.Conn) {
	ln, err := net.Listen("tcp", ":0")

	// standard boilerplate for catching errors
	if err != nil {
		log.Fatal(err)
	}

	//get outbound IP address
	ipAddress := GetOutboundIP()

	//get the port the listener is currently listening on
	port := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	ipAndPort := ipAddress + ":" + port + "\n"
	fmt.Printf("Listening for connections on IP:port " + ipAndPort)

	// channel used to communicate between writer and listener
	channel := make(chan string)
	cs := connStruct{make([]net.Conn, 0)}

	//add parent node to connections and create listener:
	cs.connections = append(cs.connections, parent)

	messagesSent := make(map[string]bool)

	go listener(parent, channel)
	go writer(&cs, &messagesSent, channel)
	go readTerminalInput(channel)

	// listen for new connections
	for {
		conn, _ := ln.Accept()
		cs.connections = append(cs.connections, conn)
		sendPreviousMessages(conn, messagesSent)
		go listener(conn, channel)
	}
}

// this starts a peer to peer network on this computer
func startOwnNetwork() {
	ln, err := net.Listen("tcp", ":0")

	// standard boilerplate for catching errors
	if err != nil {
		log.Fatal(err)
	}

	//get outbound IP address
	ipAddress := GetOutboundIP()

	//get the port the listener is currently listening on
	port := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	ipAndPort := ipAddress + ":" + port
	fmt.Printf("Listening for connections on IP:port " + ipAndPort)

	// channel used to communicate between writer and listener
	channel := make(chan string)
	cs := connStruct{make([]net.Conn, 0)}

	messagesSent := make(map[string]bool)

	go writer(&cs, &messagesSent, channel)
	go readTerminalInput(channel)

	for {
		conn, _ := ln.Accept()
		print("Connection established.\n")
		cs.connections = append(cs.connections, conn)
		sendPreviousMessages(conn, messagesSent)

		go listener(conn, channel)
	}
}

func sendPreviousMessages(conn net.Conn, messagesSent map[string]bool) {
	for k := range messagesSent {

		_, err := conn.Write([]byte(k))
		if err != nil {
			log.Fatal(err)
		}
	}
}

// writer that writes any msg sent through channel to every connection made to this computer.
// there is only on writer per computer
func writer(cs *connStruct, messagesSent *map[string]bool, channel <-chan string) {
	for {
		msg := <-channel

		// only if this msg has not been sent before will it be sent
		if (*messagesSent)[msg] == false {
			// prints the msg to the terminal only if it has not been sent before
			print(msg[:len(msg)-11] + "\n")
			for _, client := range (*cs).connections {
				client.Write([]byte(msg))
			}
			(*messagesSent)[msg] = true
		}
	}
}

// listener that listens to a particular connection. Once a msg is received, it sends a msg through channel, that tells writer
// what was received
func listener(conn net.Conn, channel chan<- string) {

	for {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			msg := scanner.Text()

			channel <- msg + "\n"
		}
	}
}

// reads input from stdin, and writes it to the channel that writer can read
func readTerminalInput(channel chan<- string) {
	for {
		// read input from terminal
		reader := bufio.NewReader(os.Stdin)
		msg, err := reader.ReadString('\n')

		if err != nil {
			print("Something went wrong\n")
			break
		}

		// write input to channel
		channel <- (msg[:len(msg)-1] + strconv.FormatInt(time.Now().Unix(), 10) + "\n")
	}
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	hostip, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		log.Fatal(err)
	}
	return hostip
}
