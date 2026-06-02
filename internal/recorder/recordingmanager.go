package recorder

import (
	"corvette/internal/cameras"
	"corvette/internal/configs"
	recordingprotocol "corvette/internal/recorder/protocol"
	"log"
)

type RecordingManager struct {
	recorders []Recorder
}

func CreateRecordingManager(rawCameras []configs.Cameras) *RecordingManager {
	cameraInstances := getCameraType(rawCameras)
	var recorders []Recorder

	for _, camInstance := range cameraInstances {
		newRecorder := mapReecorder(camInstance)
		recorders = append(recorders, newRecorder)
	}

	return &RecordingManager{
		recorders,
	}
}

func getCameraType(rawCameras []configs.Cameras) []cameras.Camera {
	var cameraInstances []cameras.Camera
	for _, camera := range rawCameras {
		setupFolderForCamera(camera.Name)
		newCamInstance := cameras.CreateNewCameraFromConfig(camera)
		cameraInstances = append(cameraInstances, newCamInstance)
	}
	return cameraInstances
}

func setupFolderForCamera(cameraName string) {
	if FolderExist(cameraName) {
		log.Printf("Folder for %s camera exists.", cameraName)
		return
	} else {
		log.Printf("Folder for %s camera DOES NOT exists. Creating one.", cameraName)
		SetupCameraFolder(cameraName)
	}
}

func mapReecorder(camera cameras.Camera) Recorder {
	switch camera.GetType() {
	case "Generic":
		return recordingprotocol.CreateNewRtspRecorder(camera.GetStreamUrl(), camera.GetName())
	}

	return nil
}

func (rm *RecordingManager) StartAllRecording() {
	for _, recorder := range rm.recorders {
		go recorder.StartStream()
	}
}
