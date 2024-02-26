package svc

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	common "github.com/cod3rboy/grpc-file-upload/common"
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

func (s *uploaderSvc) UploadFile(stream pb.UploaderService_UploadFileServer) error {
	data, err := stream.Recv()
	if err != nil {
		return err
	}
	fileMeta := data.GetMeta()
	fileName := fileMeta.GetFileNameWithExt()
	fileType := fileMeta.GetType().String()
	fileExt := strings.TrimLeft(path.Ext(fileName), ".")

	// file support validation step
	if fileExt == "" {
		return fmt.Errorf("extension not specified in file name")
	}
	if !common.IsFileExtSupported(fileType, fileExt) {
		return fmt.Errorf("file with extension %s is not supported", fileExt)
	}

	uploadFile, err := PrepareFileUpload(s.directory, fileType, fileName)
	if err != nil {
		return err
	}
	defer uploadFile.Close()

	reader := &common.UploadFileReader{
		Stream: stream,
	}

	totalBytes, err := io.Copy(uploadFile, reader)
	if err != nil {
		return fmt.Errorf("failed to save file chunks: %v", err)
	}

	log.Printf("file uploaded %s/%s with total %d bytes.", fileType, fileName, totalBytes)
	fileInfo := &pb.FileInfo{
		Name:        fileName,
		Type:        fileType,
		SizeInBytes: totalBytes,
	}

	return stream.SendAndClose(fileInfo)
}

func (s *uploaderSvc) UploadFileWithProgress(stream pb.UploaderService_UploadFileWithProgressServer) error {
	data, err := stream.Recv()
	if err != nil {
		return err
	}

	fileMeta := data.GetMeta()
	fileName := fileMeta.GetFileNameWithExt()
	fileType := fileMeta.GetType().String()
	fileExt := strings.TrimLeft(path.Ext(fileName), ".")

	// file support validation step
	if fileExt == "" {
		return fmt.Errorf("extension not specified in file name")
	}
	if !common.IsFileExtSupported(fileType, fileExt) {
		return fmt.Errorf("file with extension %s is not supported", fileExt)
	}

	uploadFile, err := PrepareFileUpload(s.directory, fileType, fileName)
	if err != nil {
		return err
	}
	defer uploadFile.Close()

	bytesReportChan := make(chan int)
	done := make(chan struct{})
	go func() {
		defer close(done)
		var received int64
		for {
			bytesReceived, ok := <-bytesReportChan
			if !ok {
				break
			}
			received += int64(bytesReceived)
			fileInfo := &pb.FileInfo{
				Name:        fileName,
				Type:        fileType,
				SizeInBytes: int64(received),
			}
			stream.Send(fileInfo)
		}
	}()
	reader := &common.UploadFileReader{
		Stream: stream,
	}
	bytesReporter := &common.ReadBytesReporter{
		ReportChannel: bytesReportChan,
	}
	totalBytes, err := io.Copy(uploadFile, io.TeeReader(reader, bytesReporter))
	close(bytesReportChan)
	<-done
	if err != nil {
		return fmt.Errorf("failed to save file chunks: %v", err)
	}

	log.Printf("file uploaded %s/%s with total %d bytes.", fileType, fileName, totalBytes)

	return nil
}

func PrepareFileUpload(rootFolder string, fileType string, fileName string) (*os.File, error) {
	wd, _ := os.Getwd()
	normalizedAbsolutePath := path.Join(wd, rootFolder, strings.ToLower(path.Join(fileType, fileName)))
	if err := os.MkdirAll(path.Dir(normalizedAbsolutePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to initialize save location: %v", err)
	}
	file, err := os.Create(normalizedAbsolutePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload file: %v", err)
	}
	return file, nil
}
