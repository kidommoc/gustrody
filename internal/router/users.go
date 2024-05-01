package router

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func routeUsers(router fiber.Router) {
	router.Get("/:username", getUserInfo)
	router.Get("/:username/posts", getUserPosts)
}

func getUserInfo(c *fiber.Ctx) error {
	username := c.Params("username")
	fmt.Printf("GET /users/%s\n", username)
	c.SendStatus(fiber.StatusOK)
	return nil
}

func getUserPosts(c *fiber.Ctx) error {
	username := c.Params("username")
	fmt.Printf("GET /users/%s/posts\n", username)
	c.SendStatus(fiber.StatusOK)
	return nil
}