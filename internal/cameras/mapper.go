package cameras

import "corvette/internal/configs"

func CreateNewCameraFromConfig(conf configs.Cameras) Camera {
	switch conf.Type {
	case "Generic":
		return &GenericIPCamera{
			IP:       conf.IP,
			Port:     conf.Port,
			User:     conf.User,
			Password: conf.Password,
			Endpoint: conf.Endpoint,
			Name:     conf.Name,
			URL:      conf.Url,
		}
	default:
		return nil
	}
}
