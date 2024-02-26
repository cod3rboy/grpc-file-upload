package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/cod3rboy/grpc-file-upload/cmd/server/interceptors"
	"google.golang.org/grpc"
)

var (
	port = flag.String("port", "8000", "server port")
	dir  = flag.String("dir", "tmp", "directory which stores uploaded files")
)

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}

	server := grpc.NewServer(grpc.StreamInterceptor(interceptors.StreamLogInterceptor))
	registerServices(server)

	fmt.Printf("server listening on port %s ...\n", *port)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to run grpc server: %v", err)
	}
}
