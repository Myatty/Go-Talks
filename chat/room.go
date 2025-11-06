package main

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
}

func (r *room) run {
	for {
		select{

		// client joins the room
		case client := r.join:
			r.clients[client] = true

		//client leave the room
		case client := r.leave:
			delete(r.clients, client)
			close(client.send)
		
		// room receives the msg
		case msg := r.forward:

			// forward the msg to all the clients
			for client := range r.clients {
				select {

				case client.send <- msg:
					// send the msg
				default:
					// fail to send msg
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}
