package services

import (
	"context"
	"corvette/internal/database"
	"corvette/internal/domains"
	"database/sql"
	"log/slog"
	"time"
)

type CameraService struct {
	db  *database.Queries
	ctx context.Context
}

func CreateCameraService(db *database.Queries, ctx context.Context) *CameraService {
	slog.Info("Camera Service created.")
	return &CameraService{
		db:  db,
		ctx: ctx,
	}
}

func (cr *CameraService) CreateCamera(opts *domains.RepoCreateCameraOpts) (*domains.Camera, error) {
	subUrl := sql.NullString{
		String: opts.SURL,
		Valid:  true,
	}
	data, err := cr.db.CreateCamera(cr.ctx,
		database.CreateCameraParams{
			CameraName:  opts.CameraName,
			InstalledAt: time.Now(),
			Status:      "Offline",
			Url:         opts.URL,
			SubUrl:      subUrl,
			Type:        opts.Type,
		})
	if err != nil {
		return nil, err
	}

	return domains.CameraFromSQLC(data), nil
}

func (cr *CameraService) GetCamera(camID int) (*domains.Camera, error) {
	camera, err := cr.db.GetCamera(cr.ctx, int64(camID))
	if err != nil {
		return nil, err
	}

	return domains.CameraFromSQLC(camera), nil
}

func (cr *CameraService) ListCameras() ([]*domains.Camera, error) {
	cameras, err := cr.db.ListCameras(cr.ctx)
	if err != nil {
		return nil, err
	}

	var mappedCameras []*domains.Camera
	for _, rawCamera := range cameras {
		mappedCameras = append(mappedCameras, domains.CameraFromSQLC(rawCamera))
	}

	return mappedCameras, nil
}

func (cr *CameraService) UpdateCamera(opts *domains.UpdateCameraOpts) (*domains.Camera, error) {
	subUrl := sql.NullString{
		String: opts.SURL,
		Valid:  true,
	}
	camera, err := cr.db.UpdateCamera(cr.ctx,
		database.UpdateCameraParams{
			CameraName: opts.CameraName,
			Url:        opts.URL,
			Type:       opts.Type,
			SubUrl:     subUrl,
			CameraID:   int64(opts.CameraId),
		},
	)
	if err != nil {
		return nil, err
	}

	return domains.CameraFromSQLC(camera), nil
}

func (cr *CameraService) UpdateCameraStatus(camID int, status string) (*domains.Camera, error) {
	camera, err := cr.db.UpdateCameraStatus(cr.ctx,
		database.UpdateCameraStatusParams{CameraID: int64(camID), Status: status})
	if err != nil {
		return nil, err
	}

	return domains.CameraFromSQLC(camera), nil
}

func (cr *CameraService) DeleteCamera(camID int) error {
	err := cr.db.DeleteCamera(cr.ctx, int64(camID))
	if err != nil {
		return err
	}
	return nil
}

func (cr *CameraService) ListOnlineCameras() ([]*domains.Camera, error) {
	cameras, err := cr.db.ListOnlineCameras(cr.ctx)
	if err != nil {
		return nil, err
	}

	var mappedCameras []*domains.Camera
	for _, rawCamera := range cameras {
		mappedCameras = append(mappedCameras, domains.CameraFromSQLC(rawCamera))
	}

	return mappedCameras, nil
}
