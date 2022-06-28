package redis

import (
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/puoklam/chat-app-backend/env"
)

var Conn redis.Conn

func init() {
	c, err := redis.Dial("tcp", env.REDIS_CONN)
	if err != nil {
		log.Fatalln(err)
	}
	Conn = c
}
