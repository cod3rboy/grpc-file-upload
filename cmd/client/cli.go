package main

import (
	"io"

	"github.com/cod3rboy/grpc-file-upload/common"
	pb "github.com/cod3rboy/grpc-file-upload/gen/uploader"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/schollz/progressbar/v3"
)

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

func startProgressIndicator(stream common.ServerStreamReceiver, totalBytes int64) chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		bar := progressbar.Default(100)
		for {
			fileInfo, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				color.Red("error while getting server progress: ", err)
				break
			}
			progressPercent := (fileInfo.SizeInBytes * 100) / totalBytes
			bar.Set64(progressPercent)
		}
	}()
	return done
}
