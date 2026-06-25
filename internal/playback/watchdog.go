package playback

import (
	"context"
	"corvette/internal/domains"
	"corvette/internal/services"
	"log/slog"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
)

type ETS struct {
	startedAt   time.Time
	recordingId int
}

type VidSegmentWatchdog struct {
	rs             *services.RecordingService
	watcher        *fsnotify.Watcher
	ets            map[string]ETS
	endTimeTracker map[int]time.Time
}

func CreateVidSegmentWatchDog(rs *services.RecordingService) *VidSegmentWatchdog {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Info("Failed to start file watchdog")
		panic(err)
	}
	return &VidSegmentWatchdog{
		rs:             rs,
		watcher:        watcher,
		ets:            make(map[string]ETS),
		endTimeTracker: make(map[int]time.Time),
	}
}

func (hwd *VidSegmentWatchdog) Watch(ctx context.Context) {
	go hwd.watchRoutine(ctx)
}

func (hwd *VidSegmentWatchdog) AddRoute(path string) {
	slog.Info("Added to route.", "route", path)
	hwd.watcher.Add(path)
}

func (hwd *VidSegmentWatchdog) RemoveRoute(path string) {
	hwd.watcher.Remove(path)
}

func (hwd *VidSegmentWatchdog) watchRoutine(ctx context.Context) {
	defer hwd.watcher.Close()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled. Exiting watchdog ended.")
			return

		case event, ok := <-hwd.watcher.Events:
			if !ok {
				slog.Info("Events channel closed. Exiting watchdog.")
				return
			}

			rawCamId := filepath.Base(filepath.Dir(event.Name))

			camId, err := strconv.Atoi(rawCamId)
			if err != nil {
				slog.Error("Failed to convert to int id. CamIDs shall only be an int.", "id", rawCamId)
				continue
			}
			fileName := filepath.Base(event.Name)

			if event.Op == fsnotify.Create {
				hwd.NewFileHandler(camId, fileName)
			}

			if event.Op == fsnotify.Write {
				hwd.HandleProbe(camId, fileName)
			}

		case err, ok := <-hwd.watcher.Errors:
			if !ok {
				return
			}
			slog.Error("Watchdog error occurred", "error", err)
		}
	}
}

func (hwd *VidSegmentWatchdog) HandleProbe(camId int, fileName string) {
	recording := hwd.ets[fileName]
	endTime, duration, err := getEndDate(fileName, recording.startedAt, camId)
	if err != nil {
		slog.Error("Failed to get end date. Removing to db.", "cameraId", camId, "fileName", fileName, "err", err.Error())
		hwd.rs.DeleteRecording(recording.recordingId)
		return
	}

	hwd.endTimeTracker[camId] = hwd.endTimeTracker[camId].Add(duration)
	hwd.rs.SetEndAt(endTime, recording.recordingId)
	delete(hwd.ets, fileName)

	if _, err := hwd.rs.SetStatus(domains.StatusDone, recording.recordingId); err != nil {
		slog.Error("Failed to set recording status to DONE", "for", recording.recordingId)
	}

	slog.Info("New file updated.", "cameraId", camId, "endedAtTime", endTime)
}

func (hwd *VidSegmentWatchdog) NewFileHandler(camId int, fileName string) {
	startedTime := time.Now()
	previousEndTime, ok := hwd.endTimeTracker[camId]

	if !ok {
		hwd.endTimeTracker[camId] = startedTime
	} else if ok {
		startedTime = previousEndTime
	}

	recording, err := hwd.rs.CreateRecording(&domains.CreateRecordingOpts{
		FromCamera: camId,
		FileName:   fileName,
		StartedAt:  startedTime,
	})
	if err != nil {
		slog.Error("Failed to record file.", "cameraId", camId, "fileName", fileName, "err", err.Error())
	}

	hwd.ets[fileName] = ETS{
		startedAt:   recording.StartedAt,
		recordingId: recording.RecordID,
	}

	if _, err := hwd.rs.SetStatus(domains.StatusRecording, recording.RecordID); err != nil {
		slog.Error("Failed to set recording status to RECORDING", "for", recording.RecordID)
	}

	slog.Info("New file recorded.", "cameraId", camId, "createdTime", startedTime)
}
