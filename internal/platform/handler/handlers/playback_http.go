package http_handlers

import (
	"corvette/internal/domains"
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

type PlaybackHttpHandler struct {
	app *fiber.App
	rs  domains.RecordingService
}

func CreatePlaybackHttpHandler(app *fiber.App, rs domains.RecordingService) {
	slog.Info("Playback http handler created.")
	// handler := PlaybackHttpHandler{
	// 	app: app,
	// 	rs:  rs,
	// }
}
