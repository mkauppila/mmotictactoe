package main

import (
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
	id      int32
	conn    *net.Conn
	pooling bool
}

// CreateClient ...
func CreateClient(conn *net.Conn) (client Client) {
	client = Client{id: rand.Int31(), conn: conn, pooling: true}
	return
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
		newClient := CreateClient(&conn)
		fmt.Println("New client ", newClient)
		connections = append(connections, newClient)

		if hasEnoughPoolingClients(connections) {
			f, s := findTwoPoolingClients(&connections)
			fmt.Printf("Game between %v and %v\n", f, s)
		} else {
			fmt.Println("Not enough clients yet")
		}
	}
}
