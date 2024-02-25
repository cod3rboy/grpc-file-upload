package svc

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	pb "github.com/cod3rboy/grpc-file-upload/gen/uploader"
)

var supportedFileTypes = map[string]map[string]bool{}

func init() {
	// initialize support file types and extensions
	image := map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"gif":  true,
	}
	video := map[string]bool{
		"mp4": true,
		"mkv": true,
	}
	audio := map[string]bool{
		"mp3": true,
		"wav": true,
		"aac": true,
	}
	document := map[string]bool{
		"pdf": true,
		"csv": true,
	}
	plainText := map[string]bool{
		"txt": true,
	}
	supportedFileTypes[pb.FileType_IMAGE.String()] = image
	supportedFileTypes[pb.FileType_VIDEO.String()] = video
	supportedFileTypes[pb.FileType_AUDIO.String()] = audio
	supportedFileTypes[pb.FileType_DOCUMENT.String()] = document
	supportedFileTypes[pb.FileType_TEXTPLAIN.String()] = plainText
}

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
	supportedExts := supportedFileTypes[fileType]
	_, fileSupported := supportedExts[fileExt]
	if !fileSupported {
		return fmt.Errorf("file with extension %s is not supported", fileExt)
	}

	uploadFile, err := PrepareFileUpload(s.directory, fileType, fileName)
	if err != nil {
		return err
	}
	defer uploadFile.Close()

	reader := &UploadFileReader{
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

// compile time verification for interface implementation
var _ io.Reader = (*UploadFileReader)(nil)

type UploadFileReader struct {
	Stream pb.UploaderService_UploadFileServer
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
