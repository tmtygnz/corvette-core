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
	//govalid:required
	//govalid:url
	URL string `json:"url"`

	//govalid:url
	SURL string `json:"surl"`

	//govalid:required
	//govalid:enum=Generic
	Type string `json:"type"`

	//govalid:required
	//govalid:minlength=3
	//govalid:maxlength=100
	CameraName string `json:"cameraName"`
}

type UpdateCameraOpts struct {
	//govalid:url
	URL string `json:"url"`

	//govalid:url
	SURL string `json:"surl"`

	//govalid:enum=Generic
	Type string `json:"type"`

	//govalid:minlength=3
	//govalid:maxlength=100
	CameraName string `json:"cameraName"`

	//govalid:required
	CameraId int `json:"cameraId"`
}

type CameraService interface {
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
