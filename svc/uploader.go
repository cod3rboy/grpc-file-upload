package svc

import (
	pb "github.com/cod3rboy/grpc-file-upload/gen/uploader"
)

type uploaderSvc struct {
	directory string
	pb.UnimplementedUploaderServiceServer
}

func NewUploaderService(uploadDir string) pb.UploaderServiceServer {
	return &uploaderSvc{
		directory: uploadDir,
	}
}
