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
	name  string
	state ClientState
}

// Client ...
type Client struct {
	id       int32
	outgoing chan string
	incoming chan string
	player   *Player
}

// CreateClient ...
func CreateClient(conn net.Conn) Client {
	client := Client{
		player: &Player{
			name:  "--",
			state: Introduce,
		},
		id:       rand.Int31(),
		outgoing: make(chan string, 4),
		incoming: make(chan string, 4),
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
func (client *Client) Reader(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("End of Reader goroutine for ", client.id)
			break
		}
		message, err := parseClientMessage(line)
		if err != nil {
			client.outgoing <- "Sorry, I don't understand that!\n"
		} else {
			switch client.player.state {
			case Introduce:
				switch message.command {
				case "name":
					client.player.name = message.body
					client.outgoing <- fmt.Sprintf("Hello %s\n", client.player.name)
					client.outgoing <- "Welcome to the lobby!\n"
					client.player.state = Lobby
				}
				break
			case Lobby:
				switch message.command {
				case "play":
					client.player.state = SearchForGame
					client.outgoing <- "Looking a challenger for you..."
					lookingForGameChannel <- client.id
				}
			case Play:
				fmt.Println("Player the game!")
			default:
				fmt.Println("Default: ", message)
			}
		}
	}
}

func (client Client) willDisconnect() {
	close(client.incoming)
	close(client.outgoing)
}

var lookingForGameChannel = make(chan int32, 1)
var connections = make([]Client, 0)

func searchChallengerFor(id int32) (int32, bool) {
	for _, client := range connections {
		fmt.Println("dbg ", client)
		if client.player.state == SearchForGame && client.id != id {
			return client.id, true
		}
	}

	return 0, false
}

func getClientForId(id int32) (Client, bool) {
	for _, client := range connections {
		if client.id == id {
			return client, true
		}
	}
	return Client{}, false
}

func matchClientsSearchingForGame() {
	for {
		select {
		case clientID := <-lookingForGameChannel:
			opponentClientID, ok := searchChallengerFor(clientID)
			if ok {
				fmt.Println("Matched with ", opponentClientID)
				clientA, ok := getClientForId(clientID)
				if ok == false {
					panic("Client id didn't match with client")
				}
				clientB, ok := getClientForId(opponentClientID)
				if ok == false {
					panic("Opponent id didn't match with client")
				}

				clientA.player.state = Play
				clientB.player.state = Play

				clientA.outgoing <- fmt.Sprintf("The match is starting with %s\n", clientB.player.name)
				clientB.outgoing <- fmt.Sprintf("The match is starting with %s\n", clientA.player.name)

				fmt.Printf("Start match between %v and %v", clientA, clientB)
			} else {
				fmt.Println("No opponent available")
			}
		}
	}
}

func main() {
	fmt.Println("Starting up")

	listener, _ := net.Listen("tcp", ":8081")
	defer listener.Close()

	go matchClientsSearchingForGame()
	defer func() { close(lookingForGameChannel) }()

	for {
		conn, _ := listener.Accept()
		newClient := CreateClient(conn)
		fmt.Println("New client ", newClient)
		connections = append(connections, newClient)
		newClient.outgoing <- "Welcome to Ticky Tacky Tocky World\n"
		newClient.outgoing <- "What's your name?\n"
	}
}
