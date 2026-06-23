package http_handlers

import (
	"corvette/internal/domains"
	"corvette/internal/hls"
	"corvette/internal/utils"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

type PlaybackHttpHandler struct {
	app         *fiber.App
	rs          domains.RecordingService
	hlsCompiler *hls.HLSCompiler
}

func CreatePlaybackHttpHandler(app *fiber.App, hlsCompiler *hls.HLSCompiler, rs domains.RecordingService) {
	slog.Info("Playback http handler created.")
	handler := PlaybackHttpHandler{
		app:         app,
		hlsCompiler: hlsCompiler,
		rs:          rs,
	}
	app.Get("/playback/hls/today", handler.Today)
}

func (phh *PlaybackHttpHandler) Today(ctx fiber.Ctx) error {
	cameraId := ctx.Query("camID")
	if cameraId == "" {
		return utils.CreateMessage(ctx, fiber.StatusBadRequest, "Cam ID malformed or missing", nil)
	}

	cameraIdInt, err := strconv.Atoi(cameraId)
	if err != nil {
		return utils.CreateMessage(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	location := time.FixedZone("GMT+8", 8*60*60)
	now := time.Now().In(location)

	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	endOfDay := startOfDay.AddDate(0, 0, 1)

	recordings, err := phh.rs.GetRecordingFor(&domains.GetRecordingForOpts{FromCamera: int64(cameraIdInt), StartedAt: startOfDay, EndedAt: endOfDay})

	if err != nil {
		return utils.CreateMessage(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	hlsString := phh.hlsCompiler.Compile(recordings, startOfDay, endOfDay)

	return ctx.SendString(hlsString)
}
