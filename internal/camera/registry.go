package camera

import (
	"context"
	"corvette/internal/domains"
	"corvette/internal/object_detection"
	"corvette/internal/playback"
	"corvette/internal/streamer"
	"log/slog"
	"sync"
	"time"
)

type CameraRegistry struct {
	cameraHandlers map[string]*CameraHandler
	parentCtx      context.Context
	wg             sync.WaitGroup
	modelInstance  *object_detection.ObjectDetectionModelInstance
	vidSegWatchdog *playback.VidSegmentWatchdog
	cameraService  domains.CameraService
}

func CreateCameraRegistry(ctx context.Context, modelInstance *object_detection.ObjectDetectionModelInstance, vidSegWatchdog *playback.VidSegmentWatchdog, cameraService domains.CameraService) *CameraRegistry {
	return &CameraRegistry{
		cameraHandlers: map[string]*CameraHandler{},
		parentCtx:      ctx,
		wg:             sync.WaitGroup{},
		modelInstance:  modelInstance,
		vidSegWatchdog: vidSegWatchdog,
		cameraService:  cameraService,
	}
}

func (cr *CameraRegistry) RegisterArrStreamers(streamers []streamer.Streamer) {
	for _, streamer := range streamers {
		slog.Info("Registering streamer, creating camera handler", "for", streamer.Vendor().CamName())
		objectDetectionHandler := object_detection.CreateObjectDetectionHandler(cr.modelInstance)
		newCameraHandler := CreateCameraHandler(streamer, cr.parentCtx, objectDetectionHandler, cr.cameraService)
		cr.cameraHandlers[streamer.Vendor().CamName()] = newCameraHandler
	}
}

func (cr *CameraRegistry) StartAllRegisteredCameras() {
	// Starts each registered camera by creating one goroutine for each.
	// The goroutine will then select, if parentCtx is done it will return, if not,
	// start all camera functions which waits. If that fails it will look if the parent Ctx is done, if not restart in 2 seconds.
	for _, handler := range cr.cameraHandlers {
		cr.vidSegWatchdog.AddRoute(handler.streamer.GetPath())
		cr.wg.Add(1)
		go func(handler *CameraHandler) {
			defer cr.wg.Done()
			for cr.parentCtx.Err() == nil {
				slog.Info("Starting reigstered camera:", handler.streamer.Vendor().CamName())
				handler.StartAllFunctions() // Waits

				if cr.parentCtx.Err() != nil {

					return
				}

				select {
				case <-cr.parentCtx.Done():
					return
				default:
					slog.Info("Restarting in 2 seconds.", "for", handler.streamer.Vendor().CamName())
					time.Sleep(2 * time.Second)
				}
			}
		}(handler)
	}

}

func (cr *CameraRegistry) WaitToClose() {
	cr.wg.Wait()
	cr.modelInstance.Destroy()
	slog.Info("All camera handlers closed.")
}
