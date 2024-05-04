package router

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func routePosts(router fiber.Router) {
	router.Get("/:postId", getPost)
	router.Put("/", newPost)
	router.Post("/:postId", editPost)
	router.Delete("/:postId", removePost)
	router.Put("/:postId/like", like)
	router.Delete("/:postId/like", unlike)
}

func newPost(c *fiber.Ctx) error {
	c.SendStatus(fiber.StatusOK)
}

func getPost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	c.SendStatus(fiber.StatusOK)
}

func editPost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	c.SendStatus(fiber.StatusOK)
}

func removePost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	c.SendStatus(fiber.StatusOK)
}

func like(c *fiber.Ctx) error {
	postId := c.Params("postId")
	c.SendStatus(fiber.StatusOK)
}

func unlike(c *fiber.Ctx) error {
	postId := c.Params("postId")
	c.SendStatus(fiber.StatusOK)
}