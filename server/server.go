package main

import (
	"github.com/dimaglushkov/go-chat/server/chat"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":13002")
	if err != nil {
		log.Fatal(err)
	}

	butler := chat.NewButler()
	grpcServer := grpc.NewServer()
	chat.RegisterButlerServer(grpcServer, &butler)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
