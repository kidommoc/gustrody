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
	fmt.Println("POST /api/auth/login")
	c.SendStatus(fiber.StatusOK)
	return nil
}

func refreshToken(c *fiber.Ctx) error {
	fmt.Println("GET /api/auth/token")
	c.SendStatus(fiber.StatusOK)
	return nil
}