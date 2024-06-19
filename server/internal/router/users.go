package router

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/posts"
	"github.com/kidommoc/gustrody/internal/users"
)

func routeUsers(router fiber.Router) {
	router.Get("/:username", getUserProfile)
	router.Get("/:username/posts", getUserPosts)
	router.Get("/:username/followings", getUserFollowings)
	router.Get("/:username/followers", getUserFollowers)
	router.Put("/follow/:username", mAuth, follow)
	router.Delete("/follow/:username", mAuth, unfollow)
}

func getUserProfile(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire username")
	}

	userService := users.NewService(models.UserInstance())
	profile, err := userService.GetProfile(username)
	if err != nil {
		switch err.Code() {
		case users.ErrNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("User not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]GET: request for %s", username)
	logger.Info(msg)
	c.Status(fiber.StatusOK)
	return c.JSON(profile)
}

func getUserPosts(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire username")
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	list, err := postService.GetByUser(username)
	if err != nil {
		switch err.Code() {
		case posts.ErrUserNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("User not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]GET: request for posts of %s", username)
	logger.Info(msg)
	c.Status(fiber.StatusOK)
	return c.JSON(list)
}

func getUserFollowings(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire username")
	}

	userService := users.NewService(models.UserInstance())
	list, err := userService.GetFollowings(username)
	if err != nil {
		switch err.Code() {
		case users.ErrNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("User not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]GET: request for followings of %s", username)
	logger.Info(msg)
	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"list": list,
	})
}

func getUserFollowers(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire username")
	}

	userService := users.NewService(models.UserInstance())
	list, err := userService.GetFollowers(username)
	if err != nil {
		switch err.Code() {
		case users.ErrNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString("User not found.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]GET: request for followers of %s", username)
	logger.Info(msg)
	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"list": list,
	})
}

func follow(c *fiber.Ctx) error {
	target := c.Params("username")
	if target == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire target username")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	userService := users.NewService(models.UserInstance())
	err := userService.Follow(strings.Clone(username), strings.Clone(target))
	if err != nil {
		switch err.Code() {
		case users.ErrSelfFollow:
			c.Status(fiber.StatusBadRequest)
			return c.SendString("Try to self-follow")
		case users.ErrNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString(fmt.Sprintf("User not found: %s", err.Error()))
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]FOLLOW: %s follows %s", username, target)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}

func unfollow(c *fiber.Ctx) error {
	target := c.Params("username")
	if target == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire target username")
	}
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	userService := users.NewService(models.UserInstance())
	err := userService.Unfollow(strings.Clone(username), strings.Clone(target))
	if err != nil {
		switch err.Code() {
		case users.ErrSelfFollow:
			c.Status(fiber.StatusBadRequest)
			return c.SendString("Try to self-unfollow")
		case users.ErrNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString(fmt.Sprintf("User not found: %s", err.Error()))
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]UNFOLLOW: %s unfollows %s", username, target)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}
