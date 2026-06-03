package recorder

import (
	"corvette/internal/capture"
)

type RecordingManager struct {
	capturers []capture.Capturer
}

func CreateRecordingManager(rawCameras []capture.Capturer) *RecordingManager {
	return &RecordingManager{
		rawCameras,
	}
}

func (rm *RecordingManager) StartAllRecording() {
	for _, capturer := range rm.capturers {
		go capturer.StartRecorder()
		go capturer.StartAIStreamer()
	}
}

func (rm *RecordingManager) StopAllRecording() {
	for _, capturer := range rm.capturers {
		go capturer.StopRecorder()
		go capturer.StopAIStreamer()
	}
}
