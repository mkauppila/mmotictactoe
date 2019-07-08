package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
)

type Tile uint32

const (
	Empty Tile = iota
	Cross
	Nought
)

// Game ..
type Game struct {
	board   [][]Tile
	playerA Client
	playerB Client
}

// CreateGame ...
func CreateGame(playerA, playerB Client) Game {
	return Game{}
}

// Client ...
type Client struct {
	id         int32
	conn       net.Conn
	writer     *bufio.Writer
	reader     *bufio.Reader
	pooling    bool
	outgoing   chan string
	incoming   chan string
	disconnect chan bool
}

// CreateClient ...
func CreateClient(conn net.Conn) Client {
	client := Client{
		id:         rand.Int31(),
		conn:       conn,
		pooling:    true,
		writer:     bufio.NewWriter(conn),
		reader:     bufio.NewReader(conn),
		outgoing:   make(chan string, 4),
		incoming:   make(chan string, 4),
		disconnect: make(chan bool, 1),
	}
	go client.Writer()
	go client.Reader()
	return client
}

// Writer ...
func (client Client) Writer() {
	for {
		select {
		case msg := <-client.outgoing:
			client.writer.Write([]byte(msg))
			client.writer.Flush()
		case <-client.disconnect:
			fmt.Println("End of Writer goroutine for ", client.id)
			break
		}
	}
}

// Reader ...
func (client Client) Reader() {
	for {
		line, _ := client.reader.ReadString('\n')
		fmt.Println(client.id, " says ", line)
	}
}

func (client Client) willDisconnect() {
	client.disconnect <- true
}

func hasEnoughPoolingClients(conns []Client) bool {
	counter := 0
	for _, client := range conns {
		if client.pooling {
			counter++

		}
	}
	return counter >= 2
}

func findTwoPoolingClients(conns *[]Client) (first int32, second int32) {
	for _, client := range *conns {
		if first != 0 && second != 0 {
			break
		}

		if client.pooling {
			if first == 0 {
				first = client.id
			} else if second == 0 {
				second = client.id
			}
		}
	}
	return
}

func main() {
	fmt.Println("Starting up")

	connections := make([]Client, 0)

	listener, _ := net.Listen("tcp", ":8081")
	defer listener.Close()

	for {
		conn, _ := listener.Accept()
		newClient := CreateClient(conn)
		fmt.Println("New client ", newClient)
		connections = append(connections, newClient)

		if hasEnoughPoolingClients(connections) {
			f, s := findTwoPoolingClients(&connections)
			fmt.Printf("Game between %v and %v\n", f, s)
			for _, client := range connections {
				client.outgoing <- "Game starting\n"
			}
		} else {
			fmt.Println("Not enough clients yet")
			for _, client := range connections {
				client.outgoing <- "Waiting for others\n"
			}
		}
	}
}
