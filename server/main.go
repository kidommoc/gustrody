package main

import (
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/router"

	"fmt"

	"github.com/gofiber/fiber/v2"
)

func main() {
	config.Get()
	db.Init()
	models.Init()

	app := fiber.New()
	router.Route(app)

	// addr := fmt.Sprintf(":%d", cfg.Port)
	fmt.Println(app.Listen(":8000"))
}
