package inference

import (
	"log"

	"github.com/yalue/onnxruntime_go"
)

type ModelInstance struct {
	inferenceSession *onnxruntime_go.AdvancedSession
}

func NewModelInstance(dllPath string, onnxModelPath string) *ModelInstance {
	log.Printf("Using onnxlib at: %s", dllPath)

	onnxruntime_go.SetSharedLibraryPath(dllPath)

	err := onnxruntime_go.InitializeEnvironment()
	if err != nil {
		log.Panic("Failed to initialize onnx env.", err)
	}

	inputShape := onnxruntime_go.NewShape(1, 3, 416, 416)
	inputTensor, _ := onnxruntime_go.NewEmptyTensor[float32](inputShape)

	outputShape := onnxruntime_go.NewShape(1, 3598, 112)
	outputTensor, _ := onnxruntime_go.NewEmptyTensor[float32](outputShape)

	log.Printf("Using %s for inference.", onnxModelPath)
	session, err := onnxruntime_go.NewAdvancedSession(onnxModelPath, []string{"data"}, []string{"output"}, []onnxruntime_go.Value{inputTensor}, []onnxruntime_go.Value{outputTensor}, nil)
	if err != nil {
		log.Panic("Failed to load onnx model", err)
	}

	return &ModelInstance{
		session,
	}
}
