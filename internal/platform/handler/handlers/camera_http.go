package http_handlers

import (
	"context"
	"corvette/internal/domains"
	"corvette/internal/utils"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type CameraHttpHandler struct {
	app           *fiber.App
	cameraService domains.CameraService
}

func CreateCameraHttpHandler(app *fiber.App, cameraService domains.CameraService) {
	slog.Info("Camera http handler created.")
	handler := CameraHttpHandler{
		app:           app,
		cameraService: cameraService,
	}

	app.Post("/cameras/register", handler.RegisterCameraEp)
	app.Patch("/camera", handler.UpdateCameraEp)

	app.Get("/cameras/", handler.GetCamera)
	app.Get("/cameras/online", handler.ListOnlineCameras)

	app.Delete("/cameras", handler.DeleteCamera)
}

func (chh *CameraHttpHandler) RegisterCameraEp(ctx fiber.Ctx) error {
	cameraInfo := new(domains.RepoCreateCameraOpts)
	if err := ctx.Bind().Body(cameraInfo); err != nil {
		return utils.CreateMessage(ctx, fiber.StatusBadRequest, "Bad or malformed request. Check API docs.", nil)
	}

	if err := domains.ValidateRepoCreateCameraOptsContext(ctx.Context(), cameraInfo); err != nil {
		if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
			return utils.CreateMessage(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	}

	created, err := chh.cameraService.CreateCamera(cameraInfo)
	if err != nil {
		return utils.CreateMessage(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return utils.CreateMessage(ctx, fiber.StatusOK, "ok", created)
}

func (chh *CameraHttpHandler) GetCamera(ctx fiber.Ctx) error {
	cameraId := ctx.Query("camID")
	if cameraId == "" {
		return utils.CreateMessage(ctx, fiber.StatusBadRequest, "Cam ID malformed or missing", nil)
	}

	cameraIdInt, err := strconv.Atoi(cameraId)
	if err != nil {
		return utils.CreateMessage(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	camera, err := chh.cameraService.GetCamera(cameraIdInt)
	if err != nil {
		return utils.CreateMessage(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return utils.CreateMessage(ctx, fiber.StatusOK, "", camera)
}

func (chh *CameraHttpHandler) UpdateCameraEp(ctx fiber.Ctx) error {
	updateData := new(domains.UpdateCameraOpts)

	if err := ctx.Bind().Body(updateData); err != nil {
		return utils.CreateMessage(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	if err := domains.ValidateUpdateCameraOptsContext(ctx.Context(), updateData); err != nil {
		if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
			return utils.CreateMessage(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	}

	updatedCamera, err := chh.cameraService.UpdateCamera(updateData)
	if err != nil {
		return utils.CreateMessage(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return utils.CreateMessage(ctx, fiber.StatusOK, "", updatedCamera)
}

func (chh *CameraHttpHandler) DeleteCamera(ctx fiber.Ctx) error {
	cameraId := ctx.Query("camID")
	if cameraId == "" {
		return utils.CreateMessage(ctx, fiber.StatusBadRequest, "Cam ID malformed or missing", nil)
	}

	cameraIdInt, err := strconv.Atoi(cameraId)
	if err != nil {
		return utils.CreateMessage(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := chh.cameraService.DeleteCamera(cameraIdInt); err != nil {
		return utils.CreateMessage(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}
	return utils.CreateMessage(ctx, fiber.StatusOK, "", nil)
}

func (chh *CameraHttpHandler) ListOnlineCameras(ctx fiber.Ctx) error {
	data, err := chh.cameraService.ListOnlineCameras()
	if err != nil {
		return utils.CreateMessage(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}
	return utils.CreateMessage(ctx, fiber.StatusOK, "", data)
}
