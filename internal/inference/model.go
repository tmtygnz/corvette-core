package inference

import (
	"log"

	"github.com/yalue/onnxruntime_go"
)

type ModelInstance struct {
	Session      *onnxruntime_go.AdvancedSession
	InputTensor  *onnxruntime_go.Tensor[float32]
	OutputTensor *onnxruntime_go.Tensor[float32]
	Categories   []string
}

func NewModelInstance(dllPath string, onnxModelPath string) *ModelInstance {
	log.Printf("Using onnxlib at: %s", dllPath)

	onnxruntime_go.SetSharedLibraryPath(dllPath)

	err := onnxruntime_go.InitializeEnvironment()
	if err != nil {
		log.Panic("Failed to initialize onnx env.", err)
	}

	inputShape := onnxruntime_go.NewShape(1, 3, 640, 640)
	inputTensor, _ := onnxruntime_go.NewEmptyTensor[float32](inputShape)

	outputShape := onnxruntime_go.NewShape(1, 300, 6)
	outputTensor, _ := onnxruntime_go.NewEmptyTensor[float32](outputShape)

	log.Printf("Using %s for inference.", onnxModelPath)

	session, err := onnxruntime_go.NewAdvancedSession(
		onnxModelPath,
		[]string{"images"},  // Input layer name
		[]string{"output0"}, // Output layer name
		[]onnxruntime_go.Value{inputTensor},
		[]onnxruntime_go.Value{outputTensor},
		nil,
	)
	if err != nil {
		log.Panic("Failed to load onnx model", err)
	}

	categories := ReadCategories("./models/yolo26n.txt")

	return &ModelInstance{
		Session:      session,
		InputTensor:  inputTensor,
		OutputTensor: outputTensor,
		Categories:   categories,
	}
}
