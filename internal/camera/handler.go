package camera

import (
	"context"
	"corvette/internal/streamer"
	"log/slog"

	"golang.org/x/sync/errgroup"
)

type CameraHandler struct {
	streamer streamer.Streamer
	context  context.Context
}

func CreateCameraHandler(streamer streamer.Streamer, ctx context.Context) *CameraHandler {
	return &CameraHandler{
		streamer: streamer,
		context:  ctx,
	}
}

func (ch *CameraHandler) StartAllFunctions() {
	g, ctx := errgroup.WithContext(ch.context)
	g.Go(func() error {
		return ch.streamer.StartRecording(ctx)
	})
	g.Go(func() error {
		return ch.streamer.StartAIStreaming(ctx)
	})

	if err := g.Wait(); err != nil {
		slog.Error("Camera function returned", "err", err.Error())
	}
}
