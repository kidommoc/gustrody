package router

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func Route(app *fiber.App) {
	// ==========================
	// should route web page here
	// ==========================

	routeFiles(app.Group("/"))

	// api router
	routeFederal(app.Group("/"))
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

func mOptionalAuth(c *fiber.Ctx) error {
	c.Locals("forced", false)
	return c.Next()
}

type AcpMap map[string](func(*fiber.Ctx) error)

func switchByAccept(m AcpMap) func(*fiber.Ctx) error {
	return (func(c *fiber.Ctx) error {
		c.Accepts()
		ct := c.Get("Accept")
		if m[ct] != nil {
			return m[ct](c)
		}
		c.Status(fiber.StatusBadRequest)
		msg := fmt.Sprintf("Wrong Accept: `%s`", ct)
		return c.SendString(msg)
	})
}
