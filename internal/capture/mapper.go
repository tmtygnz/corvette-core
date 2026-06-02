package capture

import (
	"corvette/internal/cameras"
	"log"
)

// TODO: Handle if there is a IP
func CreateCameraCapturer(rawCamera []cameras.Camera) []Capturer {
	var capturers []Capturer
	for _, camera := range rawCamera {
		switch camera.GetType() {
		case "Generic":
			log.Printf("New Generic Caputer for: %s at %s", camera.GetName(), camera.GetStreamUrl())

			newCapturer := CreateNewCapturer(CreateNewGenericCapturerOpts{
				URL:   camera.GetStreamUrl(),
				Name:  camera.GetName(),
				AiFPS: 2,
			})
			capturers = append(capturers, newCapturer)
		}
	}
	return capturers
}
