package router

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/posts"
	"github.com/kidommoc/gustrody/internal/users"
)

func routePosts(router fiber.Router) {
	router.Get("/:postID", func(c *fiber.Ctx) error {
		c.Locals("forced", false)
		return c.Next()
	}, mAuth, getPost)
	router.Get("/:postID/likes", func(c *fiber.Ctx) error {
		c.Locals("forced", false)
		return c.Next()
	}, mAuth, getPostLikes)
	router.Get("/:postID/shares", func(c *fiber.Ctx) error {
		c.Locals("forced", false)
		return c.Next()
	}, mAuth, getPostShares)
	router.Put("/", mAuth, newPost)
	router.Post("/:postID", mAuth, editPost)
	router.Delete("/:postID", mAuth, removePost)

	router.Put("/:postID/like", mAuth, likePost)
	router.Delete("/:postID/like", mAuth, unlikePost)
	router.Put("/:postID/share", mAuth, sharePost)
	router.Delete("/:postID/share", mAuth, unsharePost)
	router.Put("/:postID/reply", mAuth, replyPost)
}

func getPost(c *fiber.Ctx) error {
	username, ok := c.Locals("username").(string)
	if !ok {
		username = ""
	}
	postID := c.Params("postID")
	if postID == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Acquire post id.")
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	post, err := postService.Get(username, postID)
	if err != nil {
		switch err.Code() {
		case posts.ErrNotPermitted:
			return c.SendStatus(fiber.StatusForbidden)
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

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]GET: request for %s", postID)
	logger.Info(msg)
	c.Status(fiber.StatusOK)
	return c.JSON(post)
}

func getPostLikes(c *fiber.Ctx) error {
	username, ok := c.Locals("username").(string)
	if !ok {
		username = ""
	}
	postID := c.Params("postID")
	if postID == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Acquire post id.")
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	list, err := postService.GetLikes(username, postID)
	if err != nil {
		switch err.Code() {
		case posts.ErrNotPermitted:
			return c.SendStatus(fiber.StatusForbidden)
		case posts.ErrPostNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]GET: request for likes of %s", postID)
	logger.Info(msg)
	c.Status(fiber.StatusOK)
	return c.JSON(list)
}

func getPostShares(c *fiber.Ctx) error {
	username, ok := c.Locals("username").(string)
	if !ok {
		username = ""
	}
	postID := c.Params("postID")
	if postID == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Acquire post id.")
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	list, err := postService.GetShares(username, postID)
	if err != nil {
		switch err.Code() {
		case posts.ErrNotPermitted:
			return c.SendStatus(fiber.StatusForbidden)
		case posts.ErrPostNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]GET: request for shares of %s", postID)
	logger.Info(msg)
	c.Status(fiber.StatusOK)
	return c.JSON(list)
}

type contentBody struct {
	Content     string            `json:"content"`
	Vsb         string            `json:"vsb"`
	Attachments []posts.AttachImg `json:"attachments,omitempty"`
}

func newPost(c *fiber.Ctx) error {
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	body := new(contentBody)
	c.BodyParser(body) // should not reach error..?
	if body.Attachments == nil {
		body.Attachments = []posts.AttachImg{}
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.New(
		username, body.Vsb, body.Content, body.Attachments,
	); err != nil {
		switch err.Code() {
		case posts.ErrUserNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Actor(User) not found.")
		case posts.ErrContent:
			switch err.Error() {
			case "empty":
				c.Status(fiber.StatusBadRequest)
				return c.SendString("Acquire content to post.")
			case "long":
				c.Status(fiber.StatusBadRequest)
				return c.SendString("Content is too long to post.")
			}
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]NEW: %s posts a new post", username)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}

func replyPost(c *fiber.Ctx) error {
	postID := c.Params("postID")
	if postID == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Acquire post id")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	content := new(contentBody)
	if err := c.BodyParser(content); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Wrong request body.")
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)

	if err := postService.Reply(
		username, postID, content.Vsb, content.Content, content.Attachments,
	); err != nil {
		switch err.Code() {
		case posts.ErrUserNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Actor(User) not found.")
		case posts.ErrPostNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post not found.")
		case posts.ErrContent:
			switch err.Error() {
			case "empty":
				c.Status(fiber.StatusBadRequest)
				return c.SendString("Acquire content to post.")
			case "long":
				c.Status(fiber.StatusBadRequest)
				return c.SendString("Content is too long to post.")
			}
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]REPLY: %s replies %s", username, postID)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}

func editPost(c *fiber.Ctx) error {
	postID := c.Params("postID")
	if postID == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Acquire post id.")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	content := new(contentBody)
	if err := c.BodyParser(content); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Wrong request body.")
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.Edit(
		username, postID, content.Content, content.Attachments,
	); err != nil {
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
				return c.SendString("Acquire content to post.")
			case "long":
				c.Status(fiber.StatusBadRequest)
				return c.SendString("Content is too long to post.")
			}
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]EDIT: %s edits %s", username, postID)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}

func removePost(c *fiber.Ctx) error {
	postID := c.Params("postID")
	if postID == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Acquire post id")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.Remove(username, postID); err != nil {
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

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]REMOVE: %s removes %s", username, postID)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}

func likePost(c *fiber.Ctx) error {
	postID := c.Params("postID")
	if postID == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Acquire post id.")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.Like(username, postID); err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]LIKE: %s likes %s", username, postID)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}

func unlikePost(c *fiber.Ctx) error {
	postID := c.Params("postID")
	if postID == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Acquire post id.")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.Unlike(username, postID); err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post not found.")
		case posts.ErrLikeNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Like not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]UNLIKE: %s unlikes %s", username, postID)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}

func sharePost(c *fiber.Ctx) error {
	postID := c.Params("postID")
	if postID == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Acquire post id.")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	vsb := c.Query("visibility")

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.Share(username, postID, vsb); err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]SHARE: %s shares %s", username, postID)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}

func unsharePost(c *fiber.Ctx) error {
	postID := c.Params("postID")
	if postID == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Acquire post id.")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.Unshare(username, postID); err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post not found.")
		case posts.ErrShareNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Share not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]UNSHARE: %s unshares %s", username, postID)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}
