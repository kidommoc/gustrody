package main

import (
	"github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/router"

	"fmt"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	db.Init()

	router.Route(app)

	fmt.Println(app.Listen(":8000"))
}