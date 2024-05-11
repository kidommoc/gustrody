package router

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func routePosts(router fiber.Router) {
	router.Get("/:postId", getPost)
	router.Put("/", mAuth, newPost)
	router.Post("/:postId", mAuth, editPost)
	router.Delete("/:postId", mAuth, removePost)
	router.Put("/:postId/like", mAuth, like)
	router.Delete("/:postId/like", mAuth, unlike)
}

func newPost(c *fiber.Ctx) error {
	username := c.Locals("username")
	fmt.Printf("[POSTS]NEW: %s posts a new post\n", username)
	return c.SendStatus(fiber.StatusOK)
}

func getPost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	fmt.Printf("[POSTS]GET: request for %s\n", postId)
	return c.SendStatus(fiber.StatusOK)
}

func editPost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	username := c.Locals("username")
	fmt.Printf("[POSTS]EDIT: %s edits %s\n", username, postId)
	return c.SendStatus(fiber.StatusOK)
}

func removePost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	username := c.Locals("username")
	fmt.Printf("[POSTS]REMOVE: %s removes %s\n", username, postId)
	return c.SendStatus(fiber.StatusOK)
}

func like(c *fiber.Ctx) error {
	postId := c.Params("postId")
	username := c.Locals("username")
	fmt.Printf("[POSTS]LIKE: %s likes %s\n", username, postId)
	return c.SendStatus(fiber.StatusOK)
}

func unlike(c *fiber.Ctx) error {
	postId := c.Params("postId")
	username := c.Locals("username")
	fmt.Printf("[POSTS]UNLIKE: %s unlikes %s\n", username, postId)
	return c.SendStatus(fiber.StatusOK)
}
