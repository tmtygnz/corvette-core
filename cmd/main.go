package main

import (
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

	recManager := recorder.CreateRecordingManager(config.Cameras)
	recManager.StartAllRecording()

	select {}
}
