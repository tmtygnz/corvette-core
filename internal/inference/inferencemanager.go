package inference

import (
	"corvette/internal/capture"
	"corvette/internal/configs"
)

type InferenceManager struct {
	model             *ModelInstance
	inferenceHanlders []*InferenceHandler
}

func CreateInferenceManager(configf configs.Config, capturer []capture.Capturer) *InferenceManager {
	modInstance := NewModelInstance(configf.OnnxLibPath, configf.OnnxModelPath)
	inferenceHandlers := createInferenceHandlers(capturer, modInstance)
	return &InferenceManager{
		modInstance,
		inferenceHandlers,
	}
}

func createInferenceHandlers(capturers []capture.Capturer, model *ModelInstance) []*InferenceHandler {
	var inferenceHandlers []*InferenceHandler

	for _, capturer := range capturers {
		newInferenceHandler := CreateNewInferenceHanlder(capturer, model)
		inferenceHandlers = append(inferenceHandlers, newInferenceHandler)
	}

	return inferenceHandlers
}

func (im *InferenceManager) StartAllHandlers() {
	for _, handler := range im.inferenceHanlders {
		go handler.StartHandler()
	}
}
