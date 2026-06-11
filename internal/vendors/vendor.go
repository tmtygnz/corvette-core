package vendors

import (
	"corvette/internal/config"
	"corvette/internal/domains"
	"log/slog"
)

type GenericVendor struct {
	url     string
	surl    string
	camType string
	camName string
}

type Vendor interface {
	URL() string
	SURL() string
	Type() string
	CamName() string
}

func VendorMapper(configInfo []config.CameraInfo) []Vendor {
	var vendors []Vendor
	for _, vendorInfo := range configInfo {
		switch vendorInfo.Type {
		case "Generic":
			newGenericCamera := CreateRtspVendor(vendorInfo.URL, vendorInfo.SURL, vendorInfo.Type, vendorInfo.CamName)
			vendors = append(vendors, newGenericCamera)
		}
	}

	slog.Info("Vendors mapped.")

	return vendors
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
			newGenericCamera := CreateRtspVendor(cameraInfo.URL, cameraInfo.SURL, cameraInfo.Type, cameraInfo.CameraName)
			vendors = append(vendors, newGenericCamera)
		}
	}

	slog.Info("Vendors mapped.")

	return vendors
}
