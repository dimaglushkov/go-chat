package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"

	"github.com/dimaglushkov/go-chat/api/butlerpb"
	"github.com/dimaglushkov/go-chat/internal/server"
)

func run(port int64) error {
	listener, err := net.Listen("tcp", ":"+strconv.FormatInt(port, 10))
	if err != nil {
		return fmt.Errorf("error while setting listener: %s", err)
	}

	butler := server.NewButler()
	grpcServer := grpc.NewServer()
	butlerpb.RegisterButlerServer(grpcServer, &butler)

	log.Printf("starting go-server-server listener on port %d\n", port)
	if err = grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("error while serving grpc server: %s", err)
	}

	return nil
}

func main() {
	portFlag := flag.Int64("port", 0, "port number for chat to run on")
	flag.Parse()

	if *portFlag == 0 {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	if err := run(*portFlag); err != nil {
		log.Fatal(err)
	}
}
