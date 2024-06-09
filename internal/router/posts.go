package router

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kidommoc/gustrody/internal/database"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/posts"
	"github.com/kidommoc/gustrody/internal/users"
)

func routePosts(router fiber.Router) {
	router.Get("/:postID", getPost)
	router.Get("/:postID/likes", getPostLikes)
	router.Get("/:postID/shares", getPostShares)
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
	postID := c.Params("postID")
	if postID == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Acquire post id.")
	}

	userService := users.NewService(database.UserInstance())
	postService := posts.NewService(database.PostInstance(), userService)
	post, err := postService.Get(postID)
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

	logger := logging.Get()
	logger.Info("[POSTS]GET: request for %s\n", postID)
	c.Status(fiber.StatusOK)
	return c.JSON(post)
}

func getPostLikes(c *fiber.Ctx) error {
	postID := c.Params("postID")
	if postID == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Acquire post id.")
	}

	userService := users.NewService(database.UserInstance())
	postService := posts.NewService(database.PostInstance(), userService)
	list, err := postService.GetLikes(postID)
	if err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	logger.Info("[POSTS]GET: request for likes of %s\n", postID)
	c.Status(fiber.StatusOK)
	return c.JSON(list)
}

func getPostShares(c *fiber.Ctx) error {
	postID := c.Params("postID")
	if postID == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Acquire post id.")
	}

	userService := users.NewService(database.UserInstance())
	postService := posts.NewService(database.PostInstance(), userService)
	list, err := postService.GetShares(postID)
	if err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	logger.Info("[POSTS]GET: request for shares of %s\n", postID)
	c.Status(fiber.StatusOK)
	return c.JSON(list)
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
		return c.SendString("Acquire content to post.")
	}

	userService := users.NewService(database.UserInstance())
	postService := posts.NewService(database.PostInstance(), userService)
	if err := postService.New(strings.Clone(username), strings.Clone(content.Content)); err != nil {
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
	logger.Info("[POSTS]NEW: %s posts a new post\n", username)
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
		return c.SendString("Acquire content to post.")
	}

	userService := users.NewService(database.UserInstance())
	postService := posts.NewService(database.PostInstance(), userService)
	if err := postService.Edit(username, postID, strings.Clone(content.Content)); err != nil {
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
	logger.Info("[POSTS]EDIT: %s edits %s\n", username, postID)
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

	userService := users.NewService(database.UserInstance())
	postService := posts.NewService(database.PostInstance(), userService)
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
	logger.Info("[POSTS]REMOVE: %s removes %s\n", username, postID)
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

	userService := users.NewService(database.UserInstance())
	postService := posts.NewService(database.PostInstance(), userService)
	if err := postService.Like(strings.Clone(username), postID); err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	logger.Info("[POSTS]LIKE: %s likes %s\n", username, postID)
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

	userService := users.NewService(database.UserInstance())
	postService := posts.NewService(database.PostInstance(), userService)
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
	logger.Info("[POSTS]UNLIKE: %s unlikes %s\n", username, postID)
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

	userService := users.NewService(database.UserInstance())
	postService := posts.NewService(database.PostInstance(), userService)
	if err := postService.Share(strings.Clone(username), postID); err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("Post not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	logger.Info("[POSTS]SHARE: %s shares %s\n", username, postID)
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

	userService := users.NewService(database.UserInstance())
	postService := posts.NewService(database.PostInstance(), userService)
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
	logger.Info("[POSTS]UNSHARE: %s unshares %s\n", username, postID)
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
		return c.SendString("Acquire content to post.")
	}

	userService := users.NewService(database.UserInstance())
	postService := posts.NewService(database.PostInstance(), userService)
	if err := postService.Reply(strings.Clone(username), postID, strings.Clone(content.Content)); err != nil {
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
	logger.Info("[POSTS]REPLY: %s replies %s\n", username, postID)
	return c.SendStatus(fiber.StatusOK)
}
