package ws

import (
	"sync"
)

var hub *Hub
var once sync.Once

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			c.Lock()
			h.clients[c] = true
			c.Unlock()
		case c := <-h.unregister:
			c.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
				// stop client from receiving msg
				// clear client receive chan?
			}
			c.Unlock()
		}
	}
}

func GetHub() *Hub {
	once.Do(func() {
		hub = &Hub{
			clients:    make(map[*Client]bool),
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
	})
	return hub
}
