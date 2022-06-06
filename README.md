# chat-app-backend

# Design

Client has multiple chats, each chat represent a chat room or conversation,
each chat consists of a read offset and multiple consumer, each consumer represent a worker with respect to topic(chat room /conversation id), channel(device id), the read offset should be stored in db (redis or rdb)

# Synchronization

Two device category
- main
- sub

Since mobile devices provide a more persistent local storage, we will be utilizing by classifying mobile device as a main device,
web on the other hand as a sub device.
Each device generate a new socket connection in client's conns map with the key of device id.

## Sub devices initialization

When user connected to backend server from a sub device, a new websocket connection is established, the front end application will request for most recent messages (i.e. 100) from main device, read offset for each conversations, rooms from backend for syncing the latest state. If main device is off, then abort the connection for sub device

whenever a message received from message queue, backend server will push it to any active connections.

Once any device read new messages, the device will emit a signal to update the read offset in db, and notify other devices by listening on db subscription.

## Sub devices termination

When user terminate a sub device, all consumer with corresponding channel(device id) should be deleted, as well as the nsqd channel

# Useful links
- [scaling nsq](https://segment.com/blog/scaling-nsq/)
- [implementing messaging queue nsq in golang using docker](https://levelup.gitconnected.com/implementing-messaging-queue-nsq-in-golang-using-docker-99b402293b12)
- [nsq and golang](http://txt.fliglio.com/2020/09/nsq-and-golang/)
- [nsq design docs](https://nsq.io/overview/design.html)
- [nsq with go](https://medium.com/@jawadahmadd/nsq-with-go-77ca1b69c4ec)
- [golang nsq chat github repo](https://github.com/manhtai/golang-nsq-chat)
- [realtime chat app using kafka springboot reactjs and websocket](https://dev.to/subhransu/realtime-chat-app-using-kafka-springboot-reactjs-and-websockets-lc)
- [java kafka](https://developer.okta.com/blog/2019/11/19/java-kafka)
- [sending websocket messages to new clients(stackoverflow)](https://stackoverflow.com/questions/65857152/sending-websocket-messages-to-new-clients)
- [should websocket server only handle get requests(stackoverflow)](https://stackoverflow.com/questions/50386211/should-websocket-server-only-handle-get-requests)
- [nsq requeue vs requeue without backoff](https://www.jajaldoang.com/post/nsq-requeue-vs-requeue-without-backoff/)
- [socket io adapter](https://socket.io/docs/v4/adapter/)
- [blog](https://chowdera.com/2021/05/20210501191844563l.html)
- [nsq github distributed system issue](https://github.com/nsqio/nsq/issues/980)
- [gorilla websocket chat example(github)](https://github.com/gorilla/websocket/tree/master/examples/chat)
- [docker compose wait for container x before starting y(stackoverflow)](https://stackoverflow.com/questions/31746182/docker-compose-wait-for-container-x-before-starting-y)
- [ubuntu generate and retrieve root ca](https://ubuntu.com/server/docs/security-trust-store)
- [gist](https://gist.github.com/rorycl/d300f3ab942fd79e6cc1f37db0c6260f)
- [jwt blog](https://mkjwk.org/)
- [jwt blog](https://docs.authlib.org/en/latest/specs/rfc8037.html)
- [common anti patterns in go web applications](https://threedots.tech/post/common-anti-patterns-in-go-web-applications/)