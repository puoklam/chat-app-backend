# chat-app-backend

## Design

Users can login from any device(s), each devie may have multiple login sessons, so an unique client contains an ip and user.

No log policy for messages

## Synchronization

We dont store messages in database, so it is important to backup your history regularly.

TODO: add backup and resotre feature

## Devices initialization

When user connected to backend server from a sub device, a new websocket connection is established.

Whenever a message received from message queue, backend server will push it to any active connections.

## Devices termination

When user terminate a sub device, all consumer with corresponding channel should be deleted, as well as the nsqd channel

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
- [p2p data sync for mobile](https://vac.dev/p2p-data-sync-for-mobile)
- [Bramble Synchronisation Protocol](https://code.briarproject.org/briar/briar-spec/blob/master/protocols/BSP.md)
- [gorm advanced query](https://gorm.io/docs/advanced_query.html)
- [p2p im](https://our.status.im/status-launches-private-peer-to-peer-messaging-protocol/)
- [p2p im](https://shazzle.com/articles/what-is-peer-to-peer-p2p-messaging/)
- [p2p im](https://github.com/erenulas/p2p-chat)
- [mobile p2p protocol](https://www.quora.com/Is-there-a-simple-P2P-protocol-for-mobile-Apps)
- [p2p group](https://ieeexplore.ieee.org/document/1698602)
- [p2p mq](https://hevodata.com/learn/message-queues/#point)
- [p2p mq](https://www.oreilly.com/library/view/java-message-service/9780596802264/ch04.html)
- [gorm m2m](https://github.com/go-gorm/gorm/issues/3462)
- [gorm association](https://gorm.io/docs/associations.html#Association-Mode)
- [api testing](https://semaphoreci.com/community/tutorials/building-and-testing-a-rest-api-in-go-with-gorilla-mux-and-postgresql)
- [gorm uuid](https://medium.com/@the.hasham.ali/how-to-use-uuid-key-type-with-gorm-cc00d4ec7100)