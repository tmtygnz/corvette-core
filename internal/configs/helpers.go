package configs

import (
	"fmt"
	"strings"
)

func GetCameraRstpURL(cam *Camera) string {
	passStr := ReplaceSpaces(cam.Password)
	conStr := fmt.Sprintf("rtsp://%s:%s@%s:%d/%s", cam.User, passStr, cam.IP, cam.Port, cam.Endpoint)
	return conStr
}

func ReplaceSpaces(url string) string {
	return strings.ReplaceAll(url, " ", "%20")
}
