package router

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kidommoc/gustrody/internal/posts"
)

func routePosts(router fiber.Router) {
	router.Get("/:postId", getPost)
	router.Put("/", mAuth, newPost)
	router.Post("/:postId", mAuth, editPost)
	router.Delete("/:postId", mAuth, removePost)
	// router.Put("/:postId/like", mAuth, likePost)
	// router.Delete("/:postId/like", mAuth, unlikePost)
	// router.Put("/:postId/share", mAuth, sharePost)
	// router.Delete("/:postId/share", mAuth, unsharePost)
}

func getPost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	if postId == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire post id.")
	}

	post, err := posts.Get(postId)
	if err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post not found.")
		case posts.ErrOwner:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post owner not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	fmt.Printf("[POSTS]GET: request for %s\n", postId)
	c.Status(fiber.StatusOK)
	return c.JSON(post)
}

type contentBody struct {
	Content string `json:"content"`
}

func newPost(c *fiber.Ctx) error {
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	content := new(contentBody)
	if err := c.BodyParser(content); err != nil {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire content to post.")
	}

	if err := posts.New(strings.Clone(username), strings.Clone(content.Content)); err != nil {
		switch err.Code() {
		case posts.ErrUserNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Actor(User) not found.")
		case posts.ErrContent:
			switch err.Error() {
			case "empty":
				c.Status(fiber.StatusBadRequest)
				c.SendString("Acquire content to post.")
			case "long":
				c.Status(fiber.StatusBadRequest)
				c.SendString("Content is too long to post.")
			}
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	fmt.Printf("[POSTS]NEW: %s posts a new post\n", username)
	return c.SendStatus(fiber.StatusOK)
}

func editPost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	if postId == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire post id.")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	content := new(contentBody)
	if err := c.BodyParser(content); err != nil {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire content to post.")
	}

	if err := posts.Edit(username, postId, strings.Clone(content.Content)); err != nil {
		switch err.Code() {
		case posts.ErrOwner:
			c.Status(fiber.StatusForbidden)
			return c.SendString("Not post owner.")
		case posts.ErrPostNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post not found.")
		case posts.ErrContent:
			switch err.Error() {
			case "empty":
				c.Status(fiber.StatusBadRequest)
				c.SendString("Acquire content to post.")
			case "long":
				c.Status(fiber.StatusBadRequest)
				c.SendString("Content is too long to post.")
			}
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	fmt.Printf("[POSTS]EDIT: %s edits %s\n", username, postId)
	return c.SendStatus(fiber.StatusOK)
}

func removePost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	if postId == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire post id")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	if err := posts.Remove(username, postId); err != nil {
		switch err.Code() {
		case posts.ErrOwner:
			c.Status(fiber.StatusForbidden)
			return c.SendString("Not post owner.")
		case posts.ErrPostNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	fmt.Printf("[POSTS]REMOVE: %s removes %s\n", username, postId)
	return c.SendStatus(fiber.StatusOK)
}

func likePost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	if postId == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire post id.")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	fmt.Printf("[POSTS]LIKE: %s likes %s\n", username, postId)
	return c.SendStatus(fiber.StatusOK)
}

func unlikePost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	if postId == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire post id.")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	fmt.Printf("[POSTS]UNLIKE: %s unlikes %s\n", username, postId)
	return c.SendStatus(fiber.StatusOK)
}

func sharePost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	if postId == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire post id.")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	fmt.Printf("[POSTS]SHARE: %s shares %s\n", username, postId)
	return c.SendStatus(fiber.StatusOK)
}

func unsharePost(c *fiber.Ctx) error {
	postId := c.Params("postId")
	if postId == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire post id.")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	fmt.Printf("[POSTS]UNSHARE: %s unshares %s\n", username, postId)
	return c.SendStatus(fiber.StatusOK)
}
