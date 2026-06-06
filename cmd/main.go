package main

import (
	"context"
	"corvette/internal/camera"
	"corvette/internal/config"
	"corvette/internal/streamer"
	"corvette/internal/vendors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	coreCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	slog.Info("Corvette started.")
	config := config.ReadConfig()

	vendorsFromConfig := vendors.VendorMapper(config.Cameras)
	streamers := streamer.StreamerMapper(vendorsFromConfig)
	cameraRegistry := camera.CreateCameraRegistry(coreCtx)
	cameraRegistry.RegisterArrStreamers(streamers)
	cameraRegistry.StartAllRegisteredCameras()

	<-coreCtx.Done()
	cameraRegistry.WaitToClose()
	slog.Info("Corvette shutting down.")
}
