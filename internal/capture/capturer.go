package capture

import (
	"log"
)

type Capturer interface {
	StartRecorder() error
	StopRecorder()

	StartAIStreamer() error
	StopAIStreamer()

	GetCurrentAIFrame() ([]byte, bool)
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
