package common

import (
	"fmt"
	"io"

	pb "github.com/cod3rboy/grpc-file-upload/gen/uploader"
)

type ClientStreamReceiver interface {
	Recv() (*pb.File, error)
}

type UploadFileReader struct {
	Stream ClientStreamReceiver
}

func (r *UploadFileReader) Read(buf []byte) (n int, err error) {
	// Receive file binary stream from client
	file, err := r.Stream.Recv()
	if err == io.EOF {
		return
	}
	if err != nil {
		err = fmt.Errorf("failed to read file chunk: %v", err)
		return
	}
	chunk := file.GetChunk()
	copy(buf, chunk)
	return len(chunk), nil
}

type ReadBytesReporter struct {
	ReportChannel chan int
}

func (c *ReadBytesReporter) Write(buf []byte) (int, error) {
	n := len(buf)
	c.ReportChannel <- n
	return n, nil
}

type ServerStreamSender interface {
	Send(*pb.File) error
}

type ServerStreamReceiver interface {
	Recv() (*pb.FileInfo, error)
}

type UploadFileWriter struct {
	Stream ServerStreamSender
}

func (w *UploadFileWriter) Write(buf []byte) (n int, err error) {
	chunk := &pb.File{
		Upload: &pb.File_Chunk{
			Chunk: buf,
		},
	}
	if err = w.Stream.Send(chunk); err == nil {
		n = len(buf)
	}
	return
}
