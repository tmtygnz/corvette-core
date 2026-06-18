package streamer

import (
	"context"
	"corvette/internal/vendors"
	"log/slog"
)

type Streamer interface {
	StartRecording(eGCtx context.Context) error
	StopRecording() error

	StartAIStreaming(eGCtx context.Context) error
	StopAIStreaming() error
	GetAIFrame() chan []float32
	Vendor() vendors.Vendor
	GetPath() string

	Destroy()
}

func StreamerMapper(vendors []vendors.Vendor) []Streamer {
	var streamers []Streamer
	for _, vendor := range vendors {
		switch vendor.Type() {
		case "Generic":
			slog.Info("Mapping to Streamer", "id", vendor.IDStr(), "name", vendor.CamName())
			newGenericStreamer := CreateRtspStreamer(&CreateRtspStreamerOpts{RtspVendor: vendor, ScalingSize: 640, StreamingFps: 2})
			streamers = append(streamers, newGenericStreamer)
		}
	}

	slog.Info("Streamers mapped.")
	return streamers
}
