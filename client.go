package main

import (
	// "bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Minute

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
    CheckOrigin:     func(r *http.Request) bool { return true },
}

type ReceivedMsg struct{
	TypeOfMsg string
	KeyInput int
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan *Game
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	incomingMessage := &ReceivedMsg{
		TypeOfMsg: "",
		KeyInput: 0,
	}
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		err := c.conn.ReadJSON(incomingMessage)
        if err != nil {
            log.Println("Error reading json.", err)
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
        }
        // log.Println(incomingMessage)
        switch incomingMessage.TypeOfMsg {
        case "keyup":
        	switch incomingMessage.KeyInput {
        	case 32:
        		c.hub.done <- true
        	case 38:
        		c.hub.gameState.Player1.Up.Keyup = true
        		c.hub.gameState.Player1.Up.Keydown = false
        	case 40:
        		c.hub.gameState.Player1.Down.Keyup = true
        		c.hub.gameState.Player1.Down.Keydown = false
        	case 87:
        		c.hub.gameState.Player2.Up.Keyup = true
        		c.hub.gameState.Player2.Up.Keydown = false
        	case 83:
        		c.hub.gameState.Player2.Down.Keyup = true
        		c.hub.gameState.Player2.Down.Keydown = false
        	}
        case "keydown":
        	switch incomingMessage.KeyInput{
        	case 38:
        		c.hub.gameState.Player1.Up.Keyup = false
        		c.hub.gameState.Player1.Up.Keydown = true
        	case 40:
        		c.hub.gameState.Player1.Down.Keyup = false
        		c.hub.gameState.Player1.Down.Keydown = true
        	case 87:
        		c.hub.gameState.Player2.Up.Keyup = false
        		c.hub.gameState.Player2.Up.Keydown = true
        	case 83:
        		c.hub.gameState.Player2.Down.Keyup = false
        		c.hub.gameState.Player2.Down.Keydown = true
        	}
        }
        // send to update wala ticker through channel
        // c.hub.broadcast <- c.hub.gameState
	}
}

func (c *Client) writePump() {
	// ticker := time.NewTicker(pingPeriod)
	defer func() {
		// ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(message); err != nil {
	            log.Println(err)
	        }
		// case <-ticker.C:
		// 	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		// 	if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		// 		return
		// 	}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan *Game, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
