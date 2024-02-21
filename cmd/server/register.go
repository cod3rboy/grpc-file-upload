package main

import (
	"github.com/cod3rboy/grpc-file-upload/gen/uploader"
	"github.com/cod3rboy/grpc-file-upload/svc"
	"google.golang.org/grpc"
)

func registerServices(server *grpc.Server) {
	uploader.RegisterUploaderServiceServer(server, svc.NewUploaderService(*dir))
}
