package hls

import (
	"corvette/internal/domains"
	"errors"
	"log/slog"
)

var ErrVideoFileMissing = errors.New("Video file is missing, it might have been deleted.")

type HLSCompiler struct {
	rs domains.RecordingService
}

func CreateHLSCompiler(rs domains.RecordingService) *HLSCompiler {
	return &HLSCompiler{
		rs: rs,
	}
}

func (hc *HLSCompiler) Compile(recs []*domains.Recording) {
	for i, recording := range recs {
		slog.Info("", "name", recording.FileName)
		if recording.EndedAt != nil {
			continue
		}

		time, err := getEndDate(recording.FileName, recording.StartedAt, recording.RecordID)
		slog.Info("end date found.", "date", time)

		if err != nil {
			slog.Info("Failed to probe. Not adding to the playlist.", "err", err.Error(), "fileName", recording.FileName)
			recs = append(recs[:i], recs[i+1:]...)
			continue
		}

		if _, err = hc.rs.SetEndAt(*time, recording.RecordID); err != nil {
			slog.Info("Failed to probe. Not adding to the playlist.", "err", err.Error(), "fileName", recording.FileName)
			recs = append(recs[:i], recs[i+1:]...)
		}
	}
}
