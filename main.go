package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strings"
)

type Tile uint32

const (
	Empty Tile = iota
	Cross
	Nought
)

type ClientState uint32

const (
	Introduce ClientState = iota
	Lobby
	SearchForGame
	Play
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

type Player struct {
	name string
}

// Client ...
type Client struct {
	state      ClientState
	id         int32
	outgoing   chan string
	incoming   chan string
	disconnect chan bool
	player     Player
}

// CreateClient ...
func CreateClient(conn net.Conn) Client {
	client := Client{
		player:     Player{},
		id:         rand.Int31(),
		state:      Introduce,
		outgoing:   make(chan string, 4),
		incoming:   make(chan string, 4),
		disconnect: make(chan bool, 1),
	}
	go client.Writer(conn)
	go client.Reader(conn)
	return client
}

// Writer ...
func (client Client) Writer(conn net.Conn) {
	writer := bufio.NewWriter(conn)
	for {
		select {
		case msg := <-client.outgoing:
			writer.Write([]byte(msg))
			writer.Flush()
		case <-client.disconnect:
			fmt.Println("End of Writer goroutine for ", client.id)
			break
		}
	}
}

type ClientMessage struct {
	command, body string
}

func parseClientMessage(message string) (ClientMessage, error) {
	splits := strings.SplitN(message, " ", 2)
	if len(splits) != 2 {
		return ClientMessage{}, fmt.Errorf("Invalid client message: %v", message)
	}
	return ClientMessage{command: splits[0], body: splits[1]}, nil
}

// Reader ...
func (client Client) Reader(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("End of REader goroutine for ", client.id)
			break
		}
		message, err := parseClientMessage(line)
		if err != nil {
			client.outgoing <- "Sorry, I don't understand that!\n"
		} else {
			switch client.state {
			case Introduce:
				client.player.name = message.body
				client.outgoing <- fmt.Sprintf("Hello %s\n", client.player.name)
				client.outgoing <- "Welcome to the lobby!\n"
				client.state = Lobby
				break
			default:
				fmt.Println("Default: ", message)
			}
		}

		// fmt.Println(client.id, " says ", line)
		// fmt.Println(message)
	}
}

func (client Client) willDisconnect() {
	client.disconnect <- true
}

func hasEnoughPoolingClients(conns []Client) bool {
	counter := 0
	for _, client := range conns {
		if client.state == SearchForGame {
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

		if client.state == SearchForGame {
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
		newClient.outgoing <- "Welcome to Ticky Tacky Tocky World\n"
		newClient.outgoing <- "What's your name?\n"

		// if hasEnoughPoolingClients(connections) {
		// 	f, s := findTwoPoolingClients(&connections)
		// 	fmt.Printf("Game between %v and %v\n", f, s)
		// 	for _, client := range connections {
		// 		client.outgoing <- "Game starting\n"
		// 	}
		// } else {
		// 	fmt.Println("Not enough clients yet")
		// 	for _, client := range connections {
		// 		client.outgoing <- "Waiting for others\n"
		// 	}
		// }
	}
}
