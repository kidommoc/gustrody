package main

import (
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/database"
	"github.com/kidommoc/gustrody/internal/router"

	"fmt"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.Get()
	database.Init()

	app := fiber.New()
	router.Route(app)

	addr := fmt.Sprintf(":%d", cfg.Port)
	fmt.Println(app.Listen(addr))
}
