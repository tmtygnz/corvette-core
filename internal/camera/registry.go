package camera

import (
	"context"
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
	hlsWatchDog    *playback.VidSegmentWatchdog
}

func CreateCameraRegistry(ctx context.Context, modelInstance *object_detection.ObjectDetectionModelInstance, hlsWatchdog *playback.VidSegmentWatchdog) *CameraRegistry {
	return &CameraRegistry{
		cameraHandlers: map[string]*CameraHandler{},
		parentCtx:      ctx,
		wg:             sync.WaitGroup{},
		modelInstance:  modelInstance,
		hlsWatchDog:    hlsWatchdog,
	}
}

func (cr *CameraRegistry) RegisterArrStreamers(streamers []streamer.Streamer) {
	for _, streamer := range streamers {
		slog.Info("Registering streamer, creating camera handler", "for", streamer.Vendor().CamName())
		objectDetectionHandler := object_detection.CreateObjectDetectionHandler(cr.modelInstance)
		newCameraHandler := CreateCameraHandler(streamer, cr.parentCtx, objectDetectionHandler)
		cr.cameraHandlers[streamer.Vendor().CamName()] = newCameraHandler
	}
}

func (cr *CameraRegistry) StartAllRegisteredCameras() {
	for _, handler := range cr.cameraHandlers {
		cr.hlsWatchDog.AddRoute(handler.streamer.GetPath())
		cr.wg.Add(1)
		go func(handler *CameraHandler) {
			defer cr.wg.Done()
			for {
				select {
				case <-cr.parentCtx.Done():
					return
				default:
					slog.Info("Starting ")
					handler.StartAllFunctions()

					select {
					case <-cr.parentCtx.Done():
						return
					default:
						slog.Info("Restarting in 2 seconds.", "for", handler.streamer.Vendor().CamName())
						time.Sleep(2 * time.Second)
					}
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
