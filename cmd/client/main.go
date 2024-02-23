package main

import (
	"context"
	"flag"
	"io"
	"os"
	"path"
	"strings"

	pb "github.com/cod3rboy/grpc-file-upload/gen/uploader"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	filePath = flag.String("file", "", "file to upload to server")
	server   = flag.String("server", "localhost:8000", "server address")
)

func main() {
	flag.Parse()

	if strings.TrimSpace(*filePath) == "" {
		color.Red("missing -file flag")
		os.Exit(1)
	}

	if err := UploadFile(*filePath); err != nil {
		os.Exit(2)
	}

	color.Green("file uploaded successfully")
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
	fileMeta := &pb.File{
		Upload: &pb.File_Meta{
			Meta: &pb.FileMeta{
				FileNameWithExt: path.Base(filePath),
				Type:            promptFileType(),
			},
		},
	}
	if err := stream.Send(fileMeta); err != nil {
		return printReturnError("failed to send file meta", err)
	}

	// send file
	writer := &UploadFileWriter{
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

func printReturnError(msg string, err error) error {
	color.Red("%s: %v", msg, err)
	return err
}

func promptFileType() pb.FileType {
	fileTypeEnums := map[string]pb.FileType{
		pb.FileType_IMAGE.String():     pb.FileType_IMAGE,
		pb.FileType_VIDEO.String():     pb.FileType_VIDEO,
		pb.FileType_AUDIO.String():     pb.FileType_AUDIO,
		pb.FileType_DOCUMENT.String():  pb.FileType_DOCUMENT,
		pb.FileType_TEXTPLAIN.String(): pb.FileType_TEXTPLAIN,
	}
	supportedTypes := make([]string, 0, len(fileTypeEnums))
	for fileType := range fileTypeEnums {
		supportedTypes = append(supportedTypes, fileType)
	}

	prompt := promptui.Select{
		Label: "Select one of the file type below",
		Items: supportedTypes,
	}

	_, result, _ := prompt.Run()
	return fileTypeEnums[result]
}

// compile time verification of interface implementation
var _ io.Writer = (*UploadFileWriter)(nil)

type UploadFileWriter struct {
	Stream pb.UploaderService_UploadFileClient
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
