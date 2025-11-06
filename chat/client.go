package main

import (
	"github.com/gorilla/websocket"
)

// Client is a single chatting user
type client struct {
	socket *websocket.Conn

	// send is a buffered channel on which msgs are sent
	send chan []byte

	// room is where the user is chatting in
	room *room
}

// read continuously receives msgs from the client's websocket conn
// and forwards them to the room's forward channel
func (c *client) read() {
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

// write continuously sends msgs from the client's send channel to the
// websocket conn. it stops and closes the socket when sending fails or
// the send channel is closed.
func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
