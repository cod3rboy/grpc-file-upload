package common

import (
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
		"mov": true,
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

func IsFileExtSupported(fileType string, fileExt string) bool {
	supportedExts := supportedFileTypes[fileType]
	_, fileSupported := supportedExts[fileExt]
	return fileSupported
}
