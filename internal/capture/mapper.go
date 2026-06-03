package capture

import (
	"corvette/internal/cameras"
	"log"
)

// TODO: Handle if there is a IP
// TODO: Passin config
func CreateCameraCapturer(rawCamera []cameras.Camera) []Capturer {
	var capturers []Capturer
	for _, camera := range rawCamera {
		switch camera.GetType() {
		case "Generic":
			log.Printf("New Generic Caputer for: %s at %s", camera.GetName(), camera.GetStreamUrl())

			newCapturer, err := CreateNewCapturer(CreateNewGenericCapturerOpts{
				URL:             camera.GetStreamUrl(),
				Name:            camera.GetName(),
				AiFPS:           2,
				AiScalingWidth:  640,
				AiScalingHeight: 640,
			})
			if err != nil {
				log.Printf("Can't create capturer for %s due to: %s", camera.GetName(), err.Error())
				continue
			}

			capturers = append(capturers, newCapturer)
		}
	}
	return capturers
}
