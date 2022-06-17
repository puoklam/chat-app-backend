package ws

import (
	"fmt"
	"strconv"
	"sync"
)

var hub *Hub
var once sync.Once

type Hub struct {
	clients    map[string]*Client // session(ip + user) -> []client
	register   chan *Client
	unregister chan *Client
}

func (h *Hub) Client(uid uint, ip string) *Client {
	return h.clients[key(uid, ip)]
}

func (h *Hub) Register() chan *Client {
	return h.register
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			c.Lock()
			k := key(c.userId, c.ip)
			if c, ok := h.clients[k]; ok {
				c.Close()
			}
			h.clients[k] = c
			c.Unlock()
		case c := <-h.unregister:
			c.Lock()
			c.Close()
			k := key(c.userId, c.ip)
			if cl, ok := h.clients[k]; ok && cl == c {
				delete(h.clients, k)
			}
			c.Unlock()
		}
	}
}

func (h *Hub) Close() {
	for k, c := range h.clients {
		c.Close()
		delete(h.clients, k)
	}
}

func GetHub() *Hub {
	once.Do(func() {
		hub = &Hub{
			clients:    make(map[string]*Client),
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
	})
	return hub
}

func key(uid uint, ip string) string {
	return fmt.Sprintf("%s:%s", strconv.FormatUint(uint64(uid), 10), ip)
}
