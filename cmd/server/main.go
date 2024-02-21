package main

import (
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"
)

var (
	port = flag.String("port", "8000", "server port")
	dir  = flag.String("dir", "tmp", "directory which stores uploaded files")
)

func main() {
	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}

	server := grpc.NewServer()
	registerServices(server)

	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to run grpc server: %v", err)
	}
}
