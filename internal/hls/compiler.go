package hls

import (
	"corvette/internal/domains"
	"errors"
	"log/slog"
	"time"
)

var ErrVideoFileMissing = errors.New("Video file is missing, it might have been deleted.")

type HLSItemType int

const (
	VideoSegment = iota
	Empty
)

type HLSItem struct {
	itemType       HLSItemType
	duration       float64
	videoSegmentId int
}

type HLSCompiler struct {
	rs domains.RecordingService
}

func CreateHLSCompiler(rs domains.RecordingService) *HLSCompiler {
	return &HLSCompiler{
		rs: rs,
	}
}

func (hc *HLSCompiler) Compile(recs []*domains.Recording, timeStartRef time.Time, timeEndRef time.Time) {
	validRecordings := hc.ValidateRecordings(recs)
	startOffset, endOffset := hc.GetStartEndOffset(validRecordings, timeStartRef, timeEndRef)
	orderedItems := hc.OrderBuilder(startOffset, endOffset, validRecordings)
}

func (hc *HLSCompiler) OrderBuilder(startOffset, endOffset time.Duration, validRecordings []*domains.Recording) []HLSItem {
	orderedItems := []HLSItem{}

	if startOffset < 0 {
		orderedItems = append(orderedItems, HLSItem{
			itemType: Empty,
			duration: startOffset.Abs().Seconds(),
		})
	}

	for i, recording := range validRecordings {
		vdur := recording.EndedAt.Sub(recording.StartedAt)
		effectiveVDur := vdur

		var delta time.Duration
		hasNext := i < len(validRecordings)-1
		if hasNext {
			nextSegment := validRecordings[i+1]
			delta = recording.EndedAt.Sub(*&nextSegment.StartedAt)

			if delta > 0 {
				effectiveVDur -= delta
			}
		}

		orderedItems = append(orderedItems, HLSItem{
			itemType:       VideoSegment,
			duration:       effectiveVDur.Seconds(),
			videoSegmentId: i,
		})

		if hasNext && delta < 0 {
			orderedItems = append(orderedItems, HLSItem{
				itemType: Empty,
				duration: delta.Abs().Seconds(),
			})
		}
	}

	if endOffset > 0 {
		orderedItems = append(orderedItems, HLSItem{
			itemType: Empty,
			duration: endOffset.Abs().Seconds(),
		})
	}

	return orderedItems
}

func (hc *HLSCompiler) ValidateRecordings(recs []*domains.Recording) []*domains.Recording {
	valid := recs[:0]

	for _, recording := range recs {
		if recording.EndedAt != nil {
			valid = append(valid, recording)
			continue
		}

		endAt, err := getEndDate(
			recording.FileName,
			recording.StartedAt,
			recording.RecordID,
		)
		if err != nil {
			slog.Info(
				"Failed to probe. Not adding to playlist.",
				"err", err,
				"fileName", recording.FileName,
			)
			continue
		}

		if _, err := hc.rs.SetEndAt(*endAt, recording.RecordID); err != nil {
			slog.Info(
				"Failed to save end date.",
				"err", err,
				"fileName", recording.FileName,
			)
			continue
		}

		recording.EndedAt = endAt
		valid = append(valid, recording)
	}
	return valid
}

func (hc *HLSCompiler) GetStartEndOffset(validRecording []*domains.Recording, currentDate time.Time, endOfDay time.Time) (time.Duration, time.Duration) {
	firstRecording := validRecording[0]
	lastRecording := validRecording[len(validRecording)-1]

	startOffset := currentDate.Sub(firstRecording.StartedAt)
	endOffset := endOfDay.Sub(*lastRecording.EndedAt)

	return startOffset, endOffset
}
