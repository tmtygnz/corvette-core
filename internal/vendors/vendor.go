package vendors

import (
	"corvette/internal/domains"
	"log/slog"
)

type GenericVendor struct {
	id      int
	url     string
	surl    string
	camType string
	camName string
}

type Vendor interface {
	ID() int
	IDStr() string
	URL() string
	SURL() string
	Type() string
	CamName() string
}

func VendorMapperFromDb(cameraService domains.CameraService) []Vendor {
	cameras, err := cameraService.ListCameras()
	if err != nil {
		slog.Error("Failed to load cameras", "err", err.Error())
	}

	var vendors []Vendor
	for _, cameraInfo := range cameras {
		switch cameraInfo.Type {
		case "Generic":
			slog.Info("Created RTSP vendor.", "for", cameraInfo.CameraName)
			newGenericCamera := CreateRtspVendor(cameraInfo.CameraId, cameraInfo.URL, cameraInfo.SURL, cameraInfo.Type, cameraInfo.CameraName)
			vendors = append(vendors, newGenericCamera)
		}
	}

	slog.Info("Vendors mapped.")

	return vendors
}
