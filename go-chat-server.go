package main

import (
	"flag"
	"github.com/dimaglushkov/go-chat/server"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
)

func main() {
	var err error
	portFlag := flag.Int64("port", 0, "port number for app to run on")
	flag.Parse()

	listener, err := net.Listen("tcp", strconv.FormatInt(*portFlag, 10))
	if err != nil {
		log.Fatal(err)
	}

	butler := server.NewButler()
	grpcServer := grpc.NewServer()
	server.RegisterButlerServer(grpcServer, &butler)

	log.Printf("starting go-server-server listener on port %d\n", *portFlag)
	if err = grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
