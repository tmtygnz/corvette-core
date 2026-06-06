package camera

import (
	"context"
	"corvette/internal/streamer"
	"log/slog"
)

type CameraRegistry struct {
	cameraHandlers map[string]*CameraHandler
	parentCtx      context.Context
}

func CreateCameraRegistry(ctx context.Context) *CameraRegistry {
	return &CameraRegistry{
		cameraHandlers: map[string]*CameraHandler{},
		parentCtx:      ctx,
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
		go handler.StartAllFunctions()
	}
}
