package camera

import (
	"context"
	"corvette/internal/domains"
	"corvette/internal/object_detection"
	"corvette/internal/streamer"
	"log/slog"

	"golang.org/x/sync/errgroup"
)

type OnlineStatus int

type CameraHandler struct {
	streamer      streamer.Streamer
	context       context.Context
	status        OnlineStatus
	cameraService domains.CameraService
	objDetModel   *object_detection.ObjectDetectionHandler
}

func CreateCameraHandler(streamer streamer.Streamer, ctx context.Context, objDet *object_detection.ObjectDetectionHandler, cameraService domains.CameraService) *CameraHandler {
	return &CameraHandler{
		streamer:      streamer,
		context:       ctx,
		objDetModel:   objDet,
		cameraService: cameraService,
	}
}

func (ch *CameraHandler) StartAllFunctions() error {
	g, ctx := errgroup.WithContext(ch.context)

	ch.cameraService.UpdateCameraStatus(ch.streamer.Vendor().ID(), "online")
	defer ch.cameraService.UpdateCameraStatus(ch.streamer.Vendor().ID(), "offline")

	g.Go(func() error {
		return ch.streamer.StartRecording(ctx)
	})
	g.Go(func() error {
		return ch.streamer.StartAIStreaming(ctx)
	})
	g.Go(func() error {
		return ch.StartAIInference(ctx)
	})

	if err := g.Wait(); err != nil {
		slog.Error("Camera function returned", "err", err.Error())
		ch.cameraService.UpdateCameraStatus(ch.streamer.Vendor().ID(), "offline")
		return err
	}

	return nil
}

func (ch *CameraHandler) StartAIInference(ctx context.Context) error {
	frameChan := ch.streamer.GetAIFrame()
	for {
		select {
		case <-ctx.Done():
			slog.Info("Closing object detection inference loop. Other function crashed.")
			return ctx.Err()
		case frame := <-frameChan:
			err := ch.objDetModel.DoInference(frame)
			if err != nil {
				return err
			}
		}
	}
}

func (ch *CameraHandler) CloseAll() {
	ch.streamer.Destroy()
	ch.objDetModel.ModelInstance().Destroy()
}
