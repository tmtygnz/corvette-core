package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

type HttpServer struct {
	fiberApp *fiber.App
}

func NewHttpHandler() *HttpServer {
	app := fiber.New()
	app.Get("/recordings/*", static.New("./recordings"))
	return &HttpServer{
		fiberApp: app,
	}
}

func (hs *HttpServer) Start(port string) {
	go hs.fiberApp.Listen(port)
}

func (hs *HttpServer) App() *fiber.App {
	return hs.fiberApp
}
