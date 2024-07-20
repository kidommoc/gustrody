package main

import (
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/router"
	"github.com/kidommoc/gustrody/internal/services"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.Get()
	models.Init()
	services.Init()

	app := fiber.New()
	router.Route(app)

	// addr := fmt.Sprintf(":%d", cfg.Port)
	if cfg.Scheme == "http" {
		panic(app.Listen(":8000"))
	} else if cfg.Scheme == "https" {
		// app.ListenTLS(":8000", certFile, keyFile)
	} else {
		panic("Failed to start server. Unknown scheme.")
	}
}
