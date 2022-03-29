# go-chat
`go-chat` is a TUI-based desktop application created mainly for demonstration purposes.

## The way it works
The app consists of two separate runnable programs -
1. `go-chat-server.go` - as you can guess, this program represents the server-side of the app
2. `go-chat.go` - client app with terminal user interface

### Server app
The server app is represented by two concepts:
1. room - TCP server, to which clients connect to communicate. Each room client is a separate goroutine. Rooms have names and size limits.
2. butler - gRPC server, which accepts gRPC-requests from clients to either find a room (basically return a port number) or to create one.

### Client app
Client app is implemented with [tview](https://github.com/rivo/tview).

I'm horrible at creating UIs and client apps, so it may seem ugly, and also I was too lazy to implement a proper error message display

Nevertheless, I believe the current UI version is somewhat useful and serve its demonstrative purposes 


### Screen samples
![addrPage](https://github.com/dimaglushkov/go-chat/blob/main/assets/addr-page.jpg)

![lobbyPage](https://github.com/dimaglushkov/go-chat/blob/main/assets/lobby-page-create.jpg)

![lobbyPage](https://github.com/dimaglushkov/go-chat/blob/main/assets/lobby-page-join.jpg)

![chatPage](https://github.com/dimaglushkov/go-chat/blob/main/assets/chat-page.jpg)
