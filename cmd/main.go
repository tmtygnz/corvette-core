package main

import (
	"context"
	"corvette/internal/camera"
	"corvette/internal/config"
	"corvette/internal/database"
	"corvette/internal/object_detection"
	"corvette/internal/platform/handler"
	http_handlers "corvette/internal/platform/handler/handlers"
	"corvette/internal/platform/provider"
	"corvette/internal/playback"
	"corvette/internal/services"
	"corvette/internal/streamer"
	"corvette/internal/vendors"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

func pprofStuffs() {
	go func() {
		http.ListenAndServe("localhost:6767", nil)
	}()
}

func main() {
	slog.Info("Corvette started.")
	pprofStuffs()
	coreCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	config := config.ReadConfig()

	dbProvider := provider.CreateSQLiteProvider()
	defer dbProvider.Close()

	queries := database.New(dbProvider.Conn)
	cameraService := services.CreateCameraService(queries, coreCtx)
	recordingService := services.CreateRecordingService(queries, coreCtx)

	playbackCleanup := playback.CreatePlaybackCleanup(recordingService, cameraService)
	playbackCleanup.CleanNils()

	objectDetectionModel := object_detection.NewObjectDetectionModelInstance(config.OnnxDllPath, config.ObjDetectionModel)

	vendorsFromConfig := vendors.VendorMapperFromDb(cameraService)
	streamers := streamer.StreamerMapper(vendorsFromConfig)

	vidSegWatchdog := playback.CreateVidSegmentWatchDog(recordingService)

	cameraRegistry := camera.CreateCameraRegistry(coreCtx, objectDetectionModel, vidSegWatchdog)
	cameraRegistry.RegisterArrStreamers(streamers)
	cameraRegistry.StartAllRegisteredCameras()

	vidSegWatchdog.Watch(coreCtx)

	httpHandler := handler.NewHttpHandler()
	http_handlers.CreateCameraHttpHandler(httpHandler.App(), cameraService)
	http_handlers.CreatePlaybackHttpHandler(httpHandler.App(), recordingService)

	httpHandler.Start(":9090")

	<-coreCtx.Done()
	cameraRegistry.WaitToClose()
	slog.Info("Corvette shutting down.")
}
