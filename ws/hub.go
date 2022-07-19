package ws

import (
	"sync"
)

var hub *Hub
var once sync.Once

type clients struct {
	sync.Mutex
	// user_id -> ip -> []*Client
	c map[uint]map[string]*Client
}
type Hub struct {
	clients    *clients
	register   chan *Client
	unregister chan *Client
}

func (h *Hub) Clients(uid uint) map[string]*Client {
	return h.clients.c[uid]
}
func (h *Hub) Client(uid uint, ip string) *Client {
	if _, ok := h.clients.c[uid]; !ok {
		return nil
	}
	return h.clients.c[uid][ip]
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			c.Lock()
			h.clients.Lock()
			if h.clients.c[c.user.ID] == nil {
				h.clients.c[c.user.ID] = make(map[string]*Client)
			}
			if cl := h.clients.c[c.user.ID][c.session.IP]; cl != nil {
				cl.Lock()
				cl.Close()
				cl.Unlock()
				delete(h.clients.c[c.user.ID], c.session.IP)
			}
			h.clients.c[c.user.ID][c.session.IP] = c
			h.clients.Unlock()
			c.Unlock()
		case c := <-h.unregister:
			if c == nil {
				continue
			}
			c.Lock()
			c.Close()
			h.clients.Lock()
			if ips := h.clients.c[c.user.ID]; ips != nil {
				if cl := ips[c.session.IP]; cl == c {
					delete(h.clients.c[c.user.ID], c.session.IP)
				}
			}
			h.clients.Unlock()
			c.Unlock()
		}
	}
}

func (h *Hub) Close() {
	for k, ips := range h.clients.c {
		for ip, c := range ips {
			c.Close()
			delete(ips, ip)
		}
		delete(h.clients.c, k)
	}
}

func GetHub() *Hub {
	once.Do(func() {
		clients := &clients{
			c: make(map[uint]map[string]*Client),
		}
		hub = &Hub{
			clients:    clients,
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
	})
	return hub
}
