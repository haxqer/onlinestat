package main

type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			if _, ok := h.clients[client.uid]; ok {
				StatData.Offline(client.uid)
				delete(h.clients, client.uid)
				close(client.send)
			}
			h.clients[client.uid] = client
			StatData.Online(client.uid)
		case client := <-h.unregister:
			if _, ok := h.clients[client.uid]; ok {
				StatData.Offline(client.uid)
				delete(h.clients, client.uid)
				close(client.send)
			}
		case message := <-h.broadcast:
			for _,client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client.uid)
				}
			}
		}
	}
}