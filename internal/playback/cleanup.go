package playback

import (
	"corvette/internal/domains"
	"log/slog"
)

type PlaybackCleanup struct {
	rs domains.RecordingService
	cs domains.CameraService
}

func CreatePlaybackCleanup(rs domains.RecordingService, cs domains.CameraService) *PlaybackCleanup {
	return &PlaybackCleanup{
		rs: rs,
		cs: cs,
	}
}

func (pc *PlaybackCleanup) CleanNils() {
	cameras, err := pc.cs.ListCameras()
	if err != nil {
		slog.Error("Failed to list all cameras.", "err", err.Error())
	}

	for _, camera := range cameras {
		pc.CleanupForCamera(camera.CameraId)
	}
}

func (pc *PlaybackCleanup) CleanupForCamera(camId int) {
	nils, err := pc.rs.GetNilStatus(camId)
	if err != nil {
		slog.Error("Failed to retrieve recording with nil end dates.", "err", err.Error())
	}

	for _, recording := range *nils {
		pc.HandleCleanup(recording, camId)
	}
}

func (pc *PlaybackCleanup) HandleCleanup(recording domains.Recording, camId int) {
	slog.Info("Handling null end time.", "recId", recording.RecordID, "status", recording.Status)
	endTime, _, err := getEndDate(recording.FileName, recording.StartedAt, camId)
	if err != nil {
		slog.Error("Failed to get end date. Removing to db.", "cameraId", camId, "fileName", recording.FileName, "err", err.Error())
		pc.rs.DeleteRecording(recording.RecordID)
		return
	}

	if _, err := pc.rs.SetEndAt(endTime, recording.RecordID); err != nil {
		slog.Error("Failed to set end time.", "for", recording.RecordID)
		return
	}

	if recording.Status == domains.StatusDone {
		return
	}

	if _, err := pc.rs.SetStatus(domains.StatusDone, recording.RecordID); err != nil {
		slog.Error("Failed to set recording status to DONE", "for", recording.RecordID)
		return
	}
}
