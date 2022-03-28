package main

import (
	"github.com/dimaglushkov/go-chat/app"
	"github.com/dimaglushkov/go-chat/chat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"time"
)

func main() {
	application := app.NewApp(grpcConnector, tcpConnector)

	if err := application.App.Run(); err != nil {
		log.Fatal(err)
	}
}

func grpcConnector(addr, port string) (chat.ButlerClient, *grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(addr+":"+port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second*3))
	if err != nil {
		return nil, nil, err
	}
	return chat.NewButlerClient(conn), nil, nil
}

func tcpConnector(addr, port string) (app.ChatConn, error) {
	conn, err := net.Dial("tcp", addr+":"+port)
	if err != nil {
		return nil, err
	}
	return app.NewChatConn(conn), nil
}

/*
func tcpConnector() {
	roomNameSize := chat.RoomNameSize{Name: "kappa", Size: 2}
	roomPort, err := c.CreateRoom(context.Background(), &roomNameSize)
}
*/
