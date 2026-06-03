package cameras

import "corvette/internal/configs"

func CreateNewCameraFromConfig(configCameras []configs.Camera) []Camera {
	var cameras []Camera
	for _, confCam := range configCameras {
		switch confCam.Type {
		case "Generic":
			genericCamera := &GenericIPCamera{
				IP:       confCam.IP,
				Port:     confCam.Port,
				User:     confCam.User,
				Password: confCam.Password,
				Endpoint: confCam.Endpoint,
				Name:     confCam.Name,
				URL:      confCam.URL,
			}
			cameras = append(cameras, genericCamera)
		}
	}
	return cameras
}
