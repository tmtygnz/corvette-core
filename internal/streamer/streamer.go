package streamer

import "context"

type Streamer interface {
	StartRecording(eGCtx context.Context) error
	StopRecording() error

	StartAIStreaming(eGCtx context.Context) error
	StopAIStreaming() error
	GetAIFrame() []byte
}
