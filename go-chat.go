package main

import (
	"github.com/dimaglushkov/go-chat/app"
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

func grpcConnector(addr, port string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(addr+":"+port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second*3))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func tcpConnector(addr, port string) (*net.TCPConn, error) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr+":"+port)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)

	if err != nil {
		return nil, err
	}
	return conn, nil
}
