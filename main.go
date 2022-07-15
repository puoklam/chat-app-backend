package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/puoklam/chat-app-backend/api/auth"
	"github.com/puoklam/chat-app-backend/api/group"
	"github.com/puoklam/chat-app-backend/api/socket"
	"github.com/puoklam/chat-app-backend/api/user"
	_ "github.com/puoklam/chat-app-backend/db"
	"github.com/puoklam/chat-app-backend/mq"
	"github.com/puoklam/chat-app-backend/server"
	"github.com/puoklam/chat-app-backend/ws"
)

func cleanup() {
	mq.StopProducers()
	ws.GetHub().Close()
	// return mq.GetConn().Close()
}

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		cleanup()
		fmt.Println("quit")
		os.Exit(0)
	}()

	go ws.GetHub().Run()
	logger := log.New(os.Stdout, "im-backend", log.LstdFlags|log.Lshortfile)

	r := chi.NewRouter()
	server.SetupMiddlewares(r)

	authHandlers := auth.NewHandlers(logger)
	authHandlers.SetupRoutes(r)

	userHandlers := user.NewHandlers(logger)
	userHandlers.SetupRoutes(r)

	grpHandlers := group.NewHandlers(logger)
	grpHandlers.SetupRoutes(r)

	wsHandlers := socket.NewHandlers(logger)
	wsHandlers.SetupRoutes(r)

	srv := server.New(r)
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatalln(err)
	}
}

// TODO: Input validation middleware

// https://levelup.gitconnected.com/implementing-messaging-queue-nsq-in-golang-using-docker-99b402293b12
// http://txt.fliglio.com/2020/09/nsq-and-golang/
// https://nsq.io/overview/design.html
// https://medium.com/@jawadahmadd/nsq-with-go-77ca1b69c4ec
// https://github.com/manhtai/golang-nsq-chat
// https://dev.to/subhransu/realtime-chat-app-using-kafka-springboot-reactjs-and-websockets-lc
// https://developer.okta.com/blog/2019/11/19/java-kafka
// https://stackoverflow.com/questions/65857152/sending-websocket-messages-to-new-clients
// https://stackoverflow.com/questions/50386211/should-websocket-server-only-handle-get-requests
// https://www.jajaldoang.com/post/nsq-requeue-vs-requeue-without-backoff/
// https://socket.io/docs/v4/adapter/
// https://chowdera.com/2021/05/20210501191844563l.html
// https://github.com/nsqio/nsq/issues/980
// https://github.com/gorilla/websocket/tree/master/examples/chat

// https://stackoverflow.com/questions/31746182/docker-compose-wait-for-container-x-before-starting-y
// https://ubuntu.com/server/docs/security-trust-store

// https://entityframework.net/knowledge-base/61074684/make-api-call-and-database-operations-atomic
// https://blog.wu-boy.com/2019/11/four-tips-with-select-in-golang/
// https://github.com/firebase/firebase-admin-go/blob/master/snippets/messaging.go
// https://medium.com/@jullainc/firebase-push-notifications-to-mobile-devices-using-nodejs-7d514e10dd4
// https://medium.com/google-cloud/firebase-developing-serverless-functions-in-go-963cb011265d
