package capture

import (
	"io"
	"log"
)

type Capturer interface {
	StartRecorder() error
	StopRecorder()

	StartAIStreamer() (io.ReadCloser, error)
	StopAIStreamer()
}

func setupFolderForCapturer(cameraName string) {
	if FolderExist(cameraName) {
		log.Printf("Folder for %s camera exists.", cameraName)
		return
	} else {
		log.Printf("Folder for %s camera DOES NOT exists. Creating one.", cameraName)
		SetupCameraFolder(cameraName)
	}
}
