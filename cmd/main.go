package main

import (
	"corvette/internal/cameras"
	"corvette/internal/capture"
	"corvette/internal/configs"
	"corvette/internal/inference"
	"corvette/internal/recorder"
)

const (
	width  = 1920
	height = 1080
)

func main() {
	config := configs.ReadConfig()

	cameras := cameras.CreateNewCameraFromConfig(config.Cameras)
	capturers := capture.CreateCameraCapturer(cameras)

	recManager := recorder.CreateRecordingManager(capturers)
	recManager.StartAllRecording()

	aiManager := inference.CreateInferenceManager(config, capturers)
	aiManager.StartAllHandlers()

	defer recManager.StopAllRecording()

	select {}
}
