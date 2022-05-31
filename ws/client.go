package ws

import (
	"encoding/json"
	"errors"
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

type Message struct {
	To      string `json:"to"`
	Topic   string `json:"topic"`
	Content string `json:"content"`
}

type Client struct {
	sync.Mutex
	logger    *log.Logger
	hub       *Hub
	conn      *websocket.Conn
	producer  *nsq.Producer
	consumers map[string]*nsq.Consumer
	send      chan []byte
	receive   chan []byte
}

// user send msg from frontend to backend
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
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
		var data *Message
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
func (c *Client) writePump() {
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
				return
			}
			w.Write(msg)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
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

func (c *Client) AddConsumer(topic string, consumer *nsq.Consumer) error {
	if _, ok := c.consumers[topic]; ok {
		return errors.New("topic exists")
	}
	c.consumers[topic] = consumer
	return nil
}

func (c *Client) StopConsumer(topic string) error {
	if _, ok := c.consumers[topic]; !ok {
		return errors.New("topic doesn't exist")
	}
	c.consumers[topic].Stop()
	delete(c.consumers, topic)
	return nil
}

func (c *Client) ClearConsumers() {
	for topic, consumer := range c.consumers {
		consumer.Stop()
		delete(c.consumers, topic)
	}
}
