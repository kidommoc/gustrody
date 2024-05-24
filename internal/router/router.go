package router

import (
	"github.com/gofiber/fiber/v2"
)

func Route(app *fiber.App) {
	// ==========================
	// should route web page here
	// ==========================

	// api router
	app.Use("/", func(c *fiber.Ctx) error {
		c.Accepts("application/json")
		return c.Next()
	})
	routeAuth(app.Group("/auth"))
	routeUsers(app.Group("/users"))
	routePosts(app.Group("/posts"))
	app.Use("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotFound)
	})
}
