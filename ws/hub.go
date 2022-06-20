package ws

import (
	"fmt"
	"strconv"
	"sync"
)

var hub *Hub
var once sync.Once

type clients struct {
	sync.Mutex
	c map[string]*Client
}
type Hub struct {
	// clients    map[string]*Client // session(ip + user) -> []client
	clients    *clients
	register   chan *Client
	unregister chan *Client
}

func (h *Hub) Client(uid uint, ip string) *Client {
	return h.clients.c[key(uid, ip)]
}

func (h *Hub) Register() chan *Client {
	return h.register
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			c.Lock()
			k := key(c.user.ID, c.ip)
			h.clients.Lock()
			if cl, ok := h.clients.c[k]; ok {
				cl.Lock()
				cl.Close()
				cl.Unlock()
				delete(h.clients.c, k)
			}
			h.clients.c[k] = c
			h.clients.Unlock()
			c.Unlock()
		case c := <-h.unregister:
			c.Lock()
			c.Close()
			k := key(c.user.ID, c.ip)
			h.clients.Lock()
			if cl, ok := h.clients.c[k]; ok && cl == c {
				delete(h.clients.c, k)
			}
			h.clients.Unlock()
			c.Unlock()
		}
	}
}

func (h *Hub) Close() {
	for k, c := range h.clients.c {
		c.Close()
		delete(h.clients.c, k)
	}
}

func GetHub() *Hub {
	once.Do(func() {
		clients := &clients{
			c: make(map[string]*Client),
		}
		hub = &Hub{
			clients:    clients,
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
	})
	return hub
}

func key(uid uint, ip string) string {
	return fmt.Sprintf("%s:%s", strconv.FormatUint(uint64(uid), 10), ip)
}
