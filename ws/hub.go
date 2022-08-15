package ws

import (
	"sync"
	"sync/atomic"
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
	count      int64
	stop       bool
	OnComplete func()
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
	defer func() {
		go h.OnComplete()
	}()
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
				cl.Close(true)
				cl.Unlock()
				delete(h.clients.c[c.user.ID], c.session.IP)
				atomic.AddInt64(&h.count, -1)
			}
			h.clients.c[c.user.ID][c.session.IP] = c
			atomic.AddInt64(&h.count, 1)
			h.clients.Unlock()
			c.Unlock()
		case c := <-h.unregister:
			if c == nil {
				continue
			}
			c.Lock()
			// c.Close(false)
			h.clients.Lock()
			if ips := h.clients.c[c.user.ID]; ips != nil {
				if cl := ips[c.session.IP]; cl == c {
					delete(h.clients.c[c.user.ID], c.session.IP)
					close(c.send)
					atomic.AddInt64(&h.count, -1)
				}
			}
			h.clients.Unlock()
			c.Unlock()
			if h.stop && atomic.LoadInt64(&h.count) == 0 {
				return
			}
		}
	}
}

func (h *Hub) Close() {
	h.stop = true
	if atomic.LoadInt64(&h.count) == 0 {
		go h.OnComplete()
	}
	for _, ips := range h.clients.c {
		for _, c := range ips {
			c.Close(false)
			// delete(ips, ip)
		}
		// delete(h.clients.c, k)
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
