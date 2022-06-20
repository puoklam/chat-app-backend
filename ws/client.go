package ws

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
	"github.com/puoklam/chat-app-backend/db/model"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Data struct {
	Topic   string `json:"topic"`
	Content string `json:"content"`
}

type Message interface {
	Body() []byte
	OnError(err error)
}

type ClientCfg struct {
	Logger *log.Logger
	Conn   *websocket.Conn
	// Producer  *nsq.Producer
	Consumers map[string][]*nsq.Consumer
	User      *model.User
	IP        string
	Send      chan Message
}

// Client per connection (each device should have at most 1 connection)
type Client struct {
	sync.Mutex
	logger *log.Logger
	conn   *websocket.Conn
	// producer  *nsq.Producer
	consumers map[string][]*nsq.Consumer
	user      *model.User
	ip        string
	send      chan Message
}

func (c *Client) Send() chan Message {
	return c.send
}

// user send msg from frontend to backend
func (c *Client) ReadPump() {
	defer func() {
		GetHub().unregister <- c
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Printf("error: %v\n", err)
			}
			break
		}
	}
}

// user receive msg from backend to frontend
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				msg.OnError(err)
				// msg.RequeueWithoutBackoff(0)
				return
			}
			if _, err := w.Write(msg.Body()); err != nil {
				msg.OnError(err)
				// msg.RequeueWithoutBackoff(0)
				return
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) AddConsumer(topic string, consumer *nsq.Consumer) {
	c.consumers[topic] = append(c.consumers[topic], consumer)
}

func (c *Client) StopConsumers(topic string) {
	for _, csr := range c.consumers[topic] {
		csr.Stop()
	}
	delete(c.consumers, topic)
}

func (c *Client) ClearConsumers() {
	for topic := range c.consumers {
		c.StopConsumers(topic)
	}
}

func (c *Client) Close() {
	c.ClearConsumers()
	c.conn.Close()
}

func NewClient(cfg *ClientCfg) *Client {
	return &Client{
		logger:    cfg.Logger,
		conn:      cfg.Conn,
		consumers: cfg.Consumers,
		user:      cfg.User,
		ip:        cfg.IP,
		send:      cfg.Send,
	}
}
