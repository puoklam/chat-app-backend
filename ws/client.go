package ws

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
	"github.com/puoklam/chat-app-backend/db"
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
	OnSuccess()
}

type ClientCfg struct {
	Logger *log.Logger
	Conn   *websocket.Conn
	// Producer  *nsq.Producer
	Consumers map[string][]*nsq.Consumer
	Session   *model.Session
	User      *model.User
	// IP        string
	Send chan Message
}

// Client per connection (each device should have at most 1 connection)
type Client struct {
	sync.Mutex
	logger *log.Logger
	conn   *websocket.Conn
	// producer  *nsq.Producer
	consumers map[string][]*nsq.Consumer // topic -> consumers
	session   *model.Session
	user      *model.User
	// ip        string
	send chan Message
	// database session keep online
	keepAlive bool
}

func (c *Client) Send() chan Message {
	return c.send
}

func (c *Client) Session() *model.Session {
	return c.session
}

func (c *Client) start(ctx context.Context) error {
	hub.register <- c
	// database update must be after hub register due to the possibility of previously active client with same session
	c.session.Status = model.StatusOnline
	if err := db.GetDB(ctx).Save(c.session).Error; err != nil {
		// c.ClearConsumers()
		c.Close(false)
		return err
	}
	go c.WritePump()
	go c.ReadPump()
	return nil
}

func (c *Client) Start() error {
	return c.start(context.Background())
}

func (c *Client) StartWithContext(ctx context.Context) error {
	return c.start(ctx)
}

// user send msg from frontend to backend
func (c *Client) ReadPump() {
	defer func() {
		c.close()
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
		c.close()
		GetHub().unregister <- c
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
			msg.OnSuccess()
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

func (c *Client) AddConsumer(ctx context.Context, topic string, consumer *nsq.Consumer) error {
	var count int
	q := "UPDATE conns SET count = count + 1 WHERE user_id = ? AND topic = ? RETURNING count"
	if err := db.GetDB(ctx).Raw(q, c.user.ID, topic).Scan(&count).Error; err != nil {
		return err
	}
	c.consumers[topic] = append(c.consumers[topic], consumer)
	return nil
}

func (c *Client) StopConsumers(topic string) int {
	// TODO: what if error occur
	var count int
	q := "UPDATE conns SET count = count - 1 WHERE user_id = ? AND topic = ? RETURNING count"
	if err := db.GetDB(nil).Raw(q, c.user.ID, topic).Scan(&count).Error; err != nil {
		return -1
	}
	for _, csr := range c.consumers[topic] {
		if csr != nil {
			csr.Stop()
		}
		// <-csr.StopChan
	}
	delete(c.consumers, topic)
	return count
}

func (c *Client) ClearConsumers() {
	for topic := range c.consumers {
		c.StopConsumers(topic)
	}
}

func (c *Client) close() {
	c.ClearConsumers()
	if !c.keepAlive {
		// TODO: handle error
		c.session.Status = model.StatusOffline
		db.GetDB(nil).Save(c.session)
	}
}

func (c *Client) Close(KeepAlive bool) {
	c.keepAlive = KeepAlive
	c.conn.Close()
}

func NewClient(cfg *ClientCfg) *Client {
	return &Client{
		logger:    cfg.Logger,
		conn:      cfg.Conn,
		consumers: cfg.Consumers,
		user:      cfg.User,
		session:   cfg.Session,
		// ip:        cfg.IP,
		send: cfg.Send,
	}
}
