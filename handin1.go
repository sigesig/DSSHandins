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

	mainNetworkFunc(conn, err)
}

// construct to hold connections slice
type connStruct struct {
	connections []net.Conn
}

func printIPAndPortInformation(ln net.Listener) {
	ipAddress := GetOutboundIP()
	port := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	ipAndPort := ipAddress + ":" + port + "\n"
	fmt.Printf("Listening for connections on IP:port " + ipAndPort)
}

func addNewConnection(cs *connStruct, connection net.Conn, channel chan string) {
	(*cs).connections = append((*cs).connections, connection)
	go listener(connection, channel)
}

func mainNetworkFunc(parent net.Conn, err error) {
	ln, listenerErr := net.Listen("tcp", ":0")

	// standard boilerplate for catching errors
	if listenerErr != nil {
		log.Fatal(listenerErr)
	}

	printIPAndPortInformation(ln)

	// channel used to communicate between writer and listener
	channel := make(chan string)
	cs := connStruct{make([]net.Conn, 0)}
	messagesSent := make(map[string]bool)

	// add writer and terminal reader to this node so it can send messages
	go writer(&cs, &messagesSent, channel)
	go readTerminalInput(channel)

	// if parent was found, connect to it
	if err == nil {
		print("Succesfully connected \n")
		addNewConnection(&cs, parent, channel)
	}

	// listen for new connections
	for {
		conn, _ := ln.Accept()
		addNewConnection(&cs, conn, channel)
		sendPreviousMessages(conn, messagesSent)
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
		msgWithTimestamp := msg[:len(msg)-1] + strconv.FormatInt(time.Now().Unix(), 10) + "\n"
		channel <- msgWithTimestamp
	}
}

//gets outboundIP
//Gotten from here https://github.com/blatchley/dist-sys-Golang/blob/master/Networking/basicgob/basicgob.go
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
