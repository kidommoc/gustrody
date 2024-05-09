package router

import (
	"github.com/gofiber/fiber/v2"
)

func Route(app *fiber.App) {
	app.Use("/api", func(c *fiber.Ctx) error {
		c.Accepts("application/json")
		return c.Next()
	})
	api := (*app).Group("/api")
	routeAuth(api.Group("/auth"))
	routeUsers(api.Group("/users"))
	routePosts(api.Group("/posts"))
}
