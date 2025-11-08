package main

import (
	"log"
	"net/http"

	"chatapp.myatty.net/trace"
	"github.com/gorilla/websocket"
)

// Use two chann to ensure that we are not trying to access the same data at the same time

type room struct {
	// forward is a channel that holds the incoming msgs
	// which should be forwarded to other Clients
	forward chan []byte

	// join is the clients who want to join the room
	join chan *client

	// leave is the clients who want to leave the room
	leave chan *client

	// clients map holds all current client in this room
	clients map[*client]bool

	// tracer receives the trace information of the activity in the room
	tracer trace.Tracer
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	for {
		select {

		// client joins the room
		case client := <-r.join:
			r.clients[client] = true
			r.tracer.Trace("New Client Joined.")

		//client leave the room
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client left the room.")

		// room receives the msg
		case msg := <-r.forward:

			// forward the msg to all the clients
			for client := range r.clients {
				select {

				case client.send <- msg:
					// send the msg
					r.tracer.Trace(" --- sent to Client")
				default:
					// fail to send msg
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace(" --- failed to send, cleaned up Client")
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	// upgrade HTTP to WebSocket
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP Error: ", err)
		return
	}

	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}

	r.join <- client
	defer func() { r.leave <- client }()

	// Since Gorilla websocket conn are not safe for concurrent writes,
	// each client has its own dedicated writer goroutine
	go client.write()

	// runs in current goroutine(main) and block until client disconnect or an error occurs
	// while blocked, continuously reads msgs from websocket and forwards them to room channel
	client.read()
}
