package services

import (
	"context"
	"corvette/internal/database"
	"corvette/internal/domains"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
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
	})

	if err != nil {
		return nil, err
	}

	return domains.RecordingFromSQLC(recording), nil
}

func (rs *RecordingService) SetEndAt(endTime time.Time, id int) (*domains.Recording, error) {
	nullEndTime := sql.NullTime{
		Time:  endTime,
		Valid: true,
	}
	recording, err := rs.db.SetEndTime(rs.ctx, database.SetEndTimeParams{
		EndedAt:  nullEndTime,
		RecordID: int64(id),
	})

	if err != nil {
		return nil, err
	}

	return domains.RecordingFromSQLC(recording), nil
}

func (rs *RecordingService) GetRecordingFor(opts *domains.GetRecordingForOpts) (*domains.SegmentMetadata, error) {
	qStart := sql.NullTime{
		Time:  opts.QueryStart,
		Valid: true,
	}

	data, err := rs.db.GetRecordingFor(rs.ctx, database.GetRecordingForParams{
		StartedAt:  opts.QueryEnd,
		EndedAt:    qStart,
		FromCamera: opts.FromCamera,
	})

	if err != nil {
		return nil, err
	}

	segments := orderBuilder(data)
	segmentMetadata := domains.SegmentMetadata{
		ForCamera: int(opts.FromCamera),
		Segments:  segments,
	}
	return &segmentMetadata, nil
}

func orderBuilder(recordingSegments []database.Recording) []domains.SegmentData {
	var segmentDatas []domains.SegmentData

	for i, currentSegment := range recordingSegments {
		filePath := fmt.Sprintf("/recordings/%d/%s", currentSegment.FromCamera, currentSegment.FileName)
		segment := domains.SegmentData{
			SegmentStart: currentSegment.StartedAt,
			SegmentEnd:   currentSegment.EndedAt.Time,
			IsGap:        false,
			Source:       filePath,
		}

		segmentDatas = append(segmentDatas, segment)

		if i == len(recordingSegments)-1 {
			continue
		}

		nextSegmentStartTime := recordingSegments[i+1].StartedAt
		if !nextSegmentStartTime.After(segment.SegmentEnd) {
			continue
		}

		segmentDatas = append(segmentDatas, domains.SegmentData{
			SegmentStart: segment.SegmentEnd,
			SegmentEnd:   nextSegmentStartTime,
			IsGap:        true,
		})
	}
	return segmentDatas
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
