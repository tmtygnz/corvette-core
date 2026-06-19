package hls

import (
	"corvette/internal/domains"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

var ErrVideoFileMissing = errors.New("Video file is missing, it might have been deleted.")

// Refer to u2takey/ffmpeg-go showProgress.go example.
type probeFormat struct {
	Duration string `json:"duration"`
}

type probeData struct {
	Format probeFormat `json:"format"`
}

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
		if recording.EndedAt != nil {
			continue
		}

		time, err := getEndDate(recording.FileName, recording.StartedAt, recording.RecordID)
		slog.Info("end date found.", "date", time)

		if err != nil {
			slog.Info("Failed to probe. Not adding to the playlist.", "err", err.Error(), "fileName", recording.FileName)
			recs = append(recs[:i], recs[i+1:]...)
		}

		hc.rs.SetEndAt(*time, recording.RecordID)
	}
}

func getEndDate(fileName string, startedAt time.Time, id int) (*time.Time, error) {
	filePath := fmt.Sprintf("./recordings/%d/%s", id, fileName)
	exist, err := recordingExist(filePath)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, ErrVideoFileMissing
	}

	secs, err := probeSecs(filePath)
	if err != nil {
		return nil, err
	}

	etime := startedAt.Add(secs)

	return &etime, nil
}

func probeSecs(filePath string) (time.Duration, error) {
	kwargs := ffmpeg_go.KwArgs{
		"show_entries": "format=duration", // Only fetch the duration field
		"v":            "quiet",           // Suppress unnecessary log output
	}
	result, err := ffmpeg_go.Probe(filePath, kwargs)
	if err != nil {
		return 0, err
	}

	pd := new(probeData)
	if err := json.Unmarshal([]byte(result), pd); err != nil {
		return 0, err
	}

	f, err := strconv.ParseFloat(pd.Format.Duration, 64)
	if err != nil {
		return 0, err
	}
	dur := time.Duration(f * float64(time.Second))
	return dur, nil
}

func recordingExist(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if errors.Is(err, os.ErrNotExist) {
		slog.Error("Recording file is missing.", "filePath", filePath)
		return false, nil
	} else if err != nil {
		return true, nil
	}

	return false, err
}
