package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
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

type ClientCfg struct {
	Logger    *log.Logger
	Conn      *websocket.Conn
	Producer  *nsq.Producer
	Consumers map[string][]*nsq.Consumer
	UserID    uint
	IP        string
	Send      chan *nsq.Message
}

// Client per connection (each device should have at most 1 connection)
type Client struct {
	sync.Mutex
	logger    *log.Logger
	conn      *websocket.Conn
	producer  *nsq.Producer
	consumers map[string][]*nsq.Consumer
	userId    uint
	ip        string
	send      chan *nsq.Message
}

func (c *Client) Send() chan *nsq.Message {
	return c.send
}

// user send msg from frontend to backend
func (c *Client) ReadPump() {
	defer func() {
		GetHub().unregister <- c
		c.ClearConsumers()
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Printf("error: %v\n", err)
			}
			break
		}
		var data *Data
		if err := json.Unmarshal(msg, &data); err != nil {
			c.logger.Printf("error: %v\n", err)
			continue
		}
		if err := c.producer.Publish(data.Topic, []byte(data.Content)); err != nil {
			c.logger.Println(err)
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
				msg.RequeueWithoutBackoff(0)
				return
			}
			if _, err := w.Write(msg.Body); err != nil {
				msg.RequeueWithoutBackoff(0)
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
		producer:  cfg.Producer,
		consumers: cfg.Consumers,
		userId:    cfg.UserID,
		ip:        cfg.IP,
		send:      cfg.Send,
	}
}
