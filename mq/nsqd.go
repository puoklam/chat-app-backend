package mq

import (
	"log"
	"os"
	"sync"

	"github.com/nsqio/go-nsq"
)

var cOnce sync.Once
var conn *nsq.Conn

type ConnDelegate struct{}

func (d *ConnDelegate) OnResponse(*nsq.Conn, []byte)              {}
func (d *ConnDelegate) OnError(*nsq.Conn, []byte)                 {}
func (d *ConnDelegate) OnMessage(*nsq.Conn, *nsq.Message)         {}
func (d *ConnDelegate) OnMessageFinished(*nsq.Conn, *nsq.Message) {}
func (d *ConnDelegate) OnMessageRequeued(*nsq.Conn, *nsq.Message) {}
func (d *ConnDelegate) OnBackoff(*nsq.Conn)                       {}
func (d *ConnDelegate) OnContinue(*nsq.Conn)                      {}
func (d *ConnDelegate) OnResume(*nsq.Conn)                        {}
func (d *ConnDelegate) OnIOError(*nsq.Conn, error)                {}
func (d *ConnDelegate) OnHeartbeat(*nsq.Conn)                     {}
func (d *ConnDelegate) OnClose(*nsq.Conn)                         {}

func GetConn() *nsq.Conn {
	cOnce.Do(func() {
		cfg := nsq.NewConfig()
		delegate := &ConnDelegate{}
		conn = nsq.NewConn(os.Getenv("NSQD_ADDR"), cfg, delegate)
		if _, err := conn.Connect(); err != nil {
			log.Println(err)
		}
	})
	return conn
}
