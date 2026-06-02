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

	inference.NewModelInstance(config.Onnxlibpath, config.Onnxmodelpath)

	cameras := cameras.CreateNewCameraFromConfig(config.Cameras)
	capturers := capture.CreateCameraCapturer(cameras)

	recManager := recorder.CreateRecordingManager(capturers)
	recManager.StartAllRecording()
	defer recManager.StopAllRecording()

	select {}
}
