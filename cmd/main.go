package main

import (
	"context"
	"corvette/internal/camera"
	"corvette/internal/config"
	"corvette/internal/streamer"
	"corvette/internal/vendors"
	"log/slog"
)

func main() {
	coreCtx := context.Background()

	slog.Info("Corvette started.")
	config := config.ReadConfig()

	vendor := &vendors.RtspVendor{
		GenericVendor: vendors.GenericVendor{
			URL:     config.Cameras[0].URL,
			Type:    config.Cameras[0].Type,
			CamName: config.Cameras[0].CamName,
		},
	}

	streamingOpts := &streamer.CreateRtspStreamerOpts{
		RtspVendor:   vendor,
		StreamingFps: 2,
		ScalingSize:  640,
	}
	streamer := streamer.CreateRtspStreamer(streamingOpts)
	cameraHandler := camera.CreateCameraHandler(streamer, coreCtx)
	cameraHandler.StartAllFunctions()

}
