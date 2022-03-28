package main

import (
	"github.com/dimaglushkov/go-chat/chat"
	"github.com/dimaglushkov/go-chat/tui"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func main() {
	app := tui.NewApp(grpcConnector)

	if err := app.App.Run(); err != nil {
		log.Fatal(err)
	}
}

func grpcConnector(addr, port string) (chat.ButlerClient, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(addr+":"+port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Duration(time.Second*3)))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return chat.NewButlerClient(conn), nil
}

/*
func tcpConnector() {
	roomNameSize := chat.RoomNameSize{Name: "kappa", Size: 2}
	roomPort, err := c.CreateRoom(context.Background(), &roomNameSize)
}
*/
