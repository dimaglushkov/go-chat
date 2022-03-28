package main

import (
	chat2 "github.com/dimaglushkov/go-chat/chat"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":13002")
	if err != nil {
		log.Fatal(err)
	}

	butler := chat2.NewButler()
	grpcServer := grpc.NewServer()
	chat2.RegisterButlerServer(grpcServer, &butler)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
