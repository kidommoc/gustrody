package router

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func routePosts(router fiber.Router) {
	router.Put("/", newPost)
	router.Get("/:postId", getPost)
	router.Post("/:postId", editPost)
	router.Delete("/:postId", removePost)
}

func newPost(c *fiber.Ctx) error {
	fmt.Println("PUT /api/posts")
	c.SendStatus(fiber.StatusOK)
	return nil
}

func getPost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	fmt.Printf("GET /api/posts/%s\n", postId)
	return c.SendStatus(fiber.StatusOK)
}

func editPost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	fmt.Printf("POST /api/posts/%s\n", postId)
	c.SendStatus(fiber.StatusOK)
	return nil
}

func removePost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	fmt.Printf("DELETE /api/posts/%s\n", postId)
	c.SendStatus(fiber.StatusOK)
	return nil
}