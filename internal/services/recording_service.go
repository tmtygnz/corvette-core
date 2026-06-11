package services

import (
	"context"
	"corvette/internal/database"
	"corvette/internal/domains"
	"log/slog"
)

type RecordingService struct {
	db  *database.Queries
	ctx context.Context
}

func CreateRecordingService(db *database.Queries, ctx context.Context) *RecordingService {
	slog.Info("Recording Service created.")
	return &RecordingService{
		db:  db,
		ctx: ctx,
	}
}

func (rs *RecordingService) CreateRecording(opts *domains.CreateRecordingOpts) (*domains.Recording, error) {
	recording, err := rs.db.CreateRecording(rs.ctx, database.CreateRecordingParams{
		FromCamera: int64(opts.FromCamera),
		FileName:   opts.FileName,
		StartedAt:  opts.StartedAt,
		Duration:   0,
	})

	if err != nil {
		return nil, err
	}

	return domains.RecordingFromSQLC(recording), nil
}

func (rs *RecordingService) SetDuration(duration int, id int) (*domains.Recording, error) {
	recording, err := rs.db.SetDuration(rs.ctx, database.SetDurationParams{
		Duration: int64(duration),
		RecordID: int64(id),
	})

	if err != nil {
		return nil, err
	}

	return domains.RecordingFromSQLC(recording), nil
}

func (rs *RecordingService) ListRecordings() ([]*domains.Recording, error) {
	rawRecordings, err := rs.db.ListRecordings(rs.ctx)

	if err != nil {
		return nil, err
	}

	var recordings []*domains.Recording
	for _, rawRecording := range rawRecordings {
		recordings = append(recordings, domains.RecordingFromSQLC(rawRecording))
	}

	return recordings, nil
}

func (rs *RecordingService) DeleteRecording(id int) error {
	return rs.db.DeleteCamera(rs.ctx, int64(id))
}
