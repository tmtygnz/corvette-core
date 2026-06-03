package inference

import (
	"bytes"
	"corvette/internal/capture"
	"log"
	"strings"
)

type InferenceHandler struct {
	capturer      capture.Capturer
	modelInstance *ModelInstance
	lastFrame     []byte
}

func CreateNewInferenceHanlder(capturer capture.Capturer, modelInstance *ModelInstance) *InferenceHandler {
	return &InferenceHandler{
		capturer,
		modelInstance,
		nil,
	}
}

func (ih *InferenceHandler) StartHandler() error {
	err := ih.capturer.StartAIStreamer()
	if err != nil {
		return err
	}

	go ih.inferenceLoop()

	return nil
}

func (ih *InferenceHandler) inferenceLoop() {
	for {
		frame, ok := ih.capturer.GetCurrentAIFrame()
		if !ok {
			log.Printf("AI Stream stopped, stopping inference loop")
			break
		}
		if bytes.Equal(frame, ih.lastFrame) {
			continue
		}

		ih.lastFrame = frame
		dataTensor := ih.prepareInput(frame)

		copy(ih.modelInstance.InputTensor.GetData(), dataTensor)

		ih.modelInstance.Session.Run()

		data := ih.modelInstance.OutputTensor.GetData()
		for i := range 300 {
			offset := i * 6
			score := data[offset+4]
			if score > 0.6 {
				classId := int(data[offset+5])
				classStr := strings.TrimSpace(ih.modelInstance.Categories[classId])
				log.Printf("Found class %s with score %f", classStr, score)
			}
		}
	}
}

func (ih *InferenceHandler) prepareInput(frame []byte) []float32 {
	const (
		width  = 640
		height = 640
		size   = width * height
	)

	input := make([]float32, size*3)

	for i := range size {
		// RGB Interleaved (byte) -> RGB Planar (float32)
		input[i] = float32(frame[i*3]) / 255.0          // R
		input[i+size] = float32(frame[i*3+1]) / 255.0   // G
		input[i+size*2] = float32(frame[i*3+2]) / 255.0 // B
	}

	return input
}
