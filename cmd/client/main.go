package main

import (
	"context"
	"flag"
	"io"
	"os"
	"path"
	"strings"

	"github.com/cod3rboy/grpc-file-upload/common"
	pb "github.com/cod3rboy/grpc-file-upload/gen/uploader"
	"github.com/fatih/color"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	filePath     = flag.String("file", "", "file to upload to server")
	server       = flag.String("server", "localhost:8000", "server address")
	withProgress = flag.Bool("progress", false, "upload file with progress")
)

func main() {
	flag.Parse()

	if strings.TrimSpace(*filePath) == "" {
		color.Red("missing -file flag")
		os.Exit(1)
	}

	if !*withProgress {
		if err := UploadFile(*filePath); err != nil {
			os.Exit(2)
		}
	} else {
		if err := UploadFileWithProgress(*filePath); err != nil {
			os.Exit(2)
		}
	}

	color.Green("file uploaded successfully")
}

func UploadFileWithProgress(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return printReturnError("failed to open file", err)
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return printReturnError("failed to get file stats", err)
	}
	totalFileSize := fileStat.Size()

	conn, err := grpc.Dial(*server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return printReturnError("failed to connect with server", err)
	}

	client := pb.NewUploaderServiceClient(conn)
	stream, err := client.UploadFileWithProgress(context.Background())
	if err != nil {
		return printReturnError("failed to create client stream", err)
	}

	// send metadata
	metaFile := makeMetaFile(filePath)
	if err := stream.Send(metaFile); err != nil {
		return printReturnError("failed to send file meta", err)
	}

	doneCh := startProgressIndicator(stream, totalFileSize)

	// stream file to the server
	writer := &common.UploadFileWriter{
		Stream: stream,
	}
	clientBytesWritten, err := io.Copy(writer, file)
	if err != nil {
		return printReturnError("error while uploading file", err)
	}
	if err = stream.CloseSend(); err != nil {
		return printReturnError("failed to close stream send", err)
	}
	<-doneCh

	color.Green("%d bytes were sent", clientBytesWritten)

	return nil
}

func UploadFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return printReturnError("failed to open file", err)
	}
	defer file.Close()

	conn, err := grpc.Dial(*server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return printReturnError("failed to connect with server", err)
	}

	client := pb.NewUploaderServiceClient(conn)
	stream, err := client.UploadFile(context.Background())
	if err != nil {
		return printReturnError("failed to create client stream", err)
	}

	// send metadata
	metaFile := makeMetaFile(filePath)
	if err := stream.Send(metaFile); err != nil {
		return printReturnError("failed to send file meta", err)
	}

	// send file
	writer := &common.UploadFileWriter{
		Stream: stream,
	}

	clientBytesWritten, err := io.Copy(writer, file)
	if err != nil {
		return printReturnError("error while uploading file", err)
	}
	color.Green("%d bytes were sent", clientBytesWritten)

	// close the stream and get server response
	fileInfo, err := stream.CloseAndRecv()
	if err != nil {
		return printReturnError("server error", err)
	}

	uploadedFileName := fileInfo.GetName()
	uploadedFileType := fileInfo.GetType()
	serverBytesReceived := fileInfo.GetSizeInBytes()
	color.Cyan("Uploaded File Name: %s", uploadedFileName)
	color.Cyan("Uploaded File Type: %s", uploadedFileType)
	color.Cyan("Received %d bytes", serverBytesReceived)

	return nil
}

func makeMetaFile(filePath string) *pb.File {
	return &pb.File{
		Upload: &pb.File_Meta{
			Meta: &pb.FileMeta{
				FileNameWithExt: path.Base(filePath),
				Type:            promptFileType(),
			},
		},
	}
}
