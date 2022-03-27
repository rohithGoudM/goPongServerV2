package main

import (
	"log"

	"github.com/jtestard/go-pong/pong"
)

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *Game

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	done chan bool

	//json
	gameState *Game
}

func NewHub(g *Game) *Hub {
	return &Hub{
		broadcast:  make(chan *Game),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		done: 		make(chan bool),
		gameState:	g,
		// state: State{[]int{103, 124, 145},[]int{85, 106, 127},115,135,-21,-1,true,0,0},
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			// log.Println(h.state)
			// log.Println(h.state.player1)
			client.send <- h.gameState
			log.Println("no of clients:",len(h.clients))
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Println("no of clients:",len(h.clients))
				if len(h.clients)==0{
					h.gameState.Reset(pong.StartState)
				}
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}