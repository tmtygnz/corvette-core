package domains

import (
	"corvette/internal/database"
	"time"
)

type Camera struct {
	CameraId    int
	CameraName  string
	InstalledAt time.Time
	Status      string
	URL         string
	SURL        string
	Type        string
}

type RepoCreateCameraOpts struct {
	URL        string
	SURL       string
	Type       string
	CameraName string
}

type UpdateCameraOpts struct {
	URL        string
	SURL       string
	Type       string
	CameraName string
}

type CameraRepository interface {
	CreateCamera(opts *RepoCreateCameraOpts) (*Camera, error)
	GetCamera(id int) (*Camera, error)
	ListCameras() ([]*Camera, error)

	UpdateCamera(opts *UpdateCameraOpts) (*Camera, error)
	UpdateCameraStatus(camID int, status string) (*Camera, error)

	DeleteCamera(id int) error

	ListOnlineCameras() ([]*Camera, error)
}

func CameraFromSQLC(raw database.Camera) *Camera {
	return &Camera{
		CameraId:    int(raw.CameraID),
		CameraName:  raw.CameraName,
		InstalledAt: raw.InstalledAt,
		Status:      raw.Status,
		URL:         raw.Url,
		SURL:        raw.SubUrl.String,
		Type:        raw.Type,
	}
}
