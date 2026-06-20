package hls

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

// Refer to u2takey/ffmpeg-go showProgress.go example.
type probeFormat struct {
	Duration string `json:"duration"`
}

type probeData struct {
	Format probeFormat `json:"format"`
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

func getEndDate(fileName string, startedAt time.Time, id int) (*time.Time, error) {
	filePath := fmt.Sprintf("./recordings/%d/%s", id, fileName)
	exist, err := recordingExist(filePath)
	if err != nil {
		return nil, err
	}

	if !exist {
		slog.Error("Failed to find file.", "filePath", filePath)
		return nil, ErrVideoFileMissing
	}

	secs, err := probeSecs(filePath)
	if err != nil {
		return nil, err
	}

	etime := startedAt.Add(secs)

	return &etime, nil
}

func recordingExist(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		slog.Error("Recording file is missing.", "filePath", filePath)
		return false, nil
	}

	return false, err
}
