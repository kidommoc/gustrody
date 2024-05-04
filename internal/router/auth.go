package router

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func routeAuth(router fiber.Router) {
	router.Post("/login", login)
	router.Get("/token", refreshToken)
}

func login(c *fiber.Ctx) error {
	c.SendStatus(fiber.StatusOK)
}

func refreshToken(c *fiber.Ctx) error {
	c.SendStatus(fiber.StatusOK)
}