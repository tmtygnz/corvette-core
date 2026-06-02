package cameras

import (
	"fmt"
)

type GenericIPCamera struct {
	IP       string
	Port     int
	User     string
	Password string
	Endpoint string
	Name     string
	URL      string
}

func (gipc *GenericIPCamera) GetStreamUrl() string {
	if gipc.URL != "" {
		return gipc.URL
	}

	return fmt.Sprintf("rtsp://%s:%s@%s:%d/%s", gipc.User, gipc.Password, gipc.IP, gipc.Port, gipc.Endpoint)
}

func (gipc *GenericIPCamera) GetType() string {
	return "Generic"
}

func (gipc *GenericIPCamera) GetName() string {
	return gipc.Name
}
