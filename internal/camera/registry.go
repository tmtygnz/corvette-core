package camera

import (
	"context"
	"corvette/internal/streamer"
	"log/slog"
	"sync"
)

type CameraRegistry struct {
	cameraHandlers map[string]*CameraHandler
	parentCtx      context.Context
	wg             sync.WaitGroup
}

func CreateCameraRegistry(ctx context.Context) *CameraRegistry {
	return &CameraRegistry{
		cameraHandlers: map[string]*CameraHandler{},
		parentCtx:      ctx,
		wg:             sync.WaitGroup{},
	}
}

func (cr *CameraRegistry) RegisterArrStreamers(streamers []streamer.Streamer) {
	for _, streamer := range streamers {
		slog.Info("Registering streamer, creating camera handler", "for", streamer.Vendor().CamName())
		newCameraHandler := CreateCameraHandler(streamer, cr.parentCtx)
		cr.cameraHandlers[streamer.Vendor().CamName()] = newCameraHandler
	}
}

func (cr *CameraRegistry) StartAllRegisteredCameras() {
	for _, handler := range cr.cameraHandlers {
		cr.wg.Add(1)
		go func(handler *CameraHandler) {
			defer cr.wg.Done()
			handler.StartAllFunctions()
		}(handler)
	}
}

func (cr *CameraRegistry) WaitToClose() {
	cr.wg.Wait()
	slog.Info("All camera handlers closed.")
}
