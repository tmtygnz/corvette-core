package camera

import (
	"context"
	"corvette/internal/services"
	"log/slog"

	"github.com/fsnotify/fsnotify"
)

type HLSWatchDog struct {
	rs      *services.RecordingService
	watcher *fsnotify.Watcher
}

func CreateHLSWatchDog(rs *services.RecordingService) *HLSWatchDog {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Info("Failed to start file watchdog")
		panic(err)
	}
	return &HLSWatchDog{
		rs:      rs,
		watcher: watcher,
	}
}

func (hwd *HLSWatchDog) Watch(ctx context.Context) {
	go hwd.watchRoutine(ctx)
}

func (hwd *HLSWatchDog) AddRoute(path string) {
	slog.Info("Added to route.", "route", path)
	hwd.watcher.Add(path)
}

func (hwd *HLSWatchDog) RemoveRoute(path string) {
	hwd.watcher.Remove(path)
}

func (hwd *HLSWatchDog) watchRoutine(ctx context.Context) {
	defer hwd.watcher.Close()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled. Exiting HLS watchdog.")
			return

		case event, ok := <-hwd.watcher.Events:
			if !ok {
				slog.Info("Events channel closed. Exiting watchdog.")
				return
			}
			slog.Info("Watchdog event received", "event", event.String())

		case err, ok := <-hwd.watcher.Errors:
			if !ok {
				return
			}
			slog.Error("Watchdog error occurred", "error", err)
		}
	}
}
