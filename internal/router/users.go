package router

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
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
	profile, err := users.GetProfile(username)
	if err != nil {
		switch {
		case err.Error() == "user not found":
			c.Status(fiber.StatusNotFound)
			return c.SendString(err.Error())
		default:
			c.Status(fiber.StatusInternalServerError)
			return c.SendString(err.Error())
		}
	}

	fmt.Printf("[USERS]GET: request for %s\n", username)
	c.Status(fiber.StatusOK)
	return c.JSON(profile)
}

func getUserPosts(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire username")
	}
	fmt.Printf("[USERS]GET: request for posts of %s\n", username)
	return c.SendStatus(fiber.StatusOK)
}

func getUserFollowings(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire username")
	}
	list, err := users.GetFollowings(username)
	if err != nil {
		switch {
		case err.Error() == "user not found":
			c.Status(fiber.StatusNotFound)
			return c.SendString(err.Error())
		default:
			c.Status(fiber.StatusInternalServerError)
			return c.SendString(err.Error())
		}
	}

	fmt.Printf("[USERS]GET: request for followings of %s\n", username)
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
	list, err := users.GetFollowers(username)
	if err != nil {
		switch {
		case err.Error() == "user not found":
			c.Status(fiber.StatusNotFound)
			return c.SendString(err.Error())
		default:
			c.Status(fiber.StatusInternalServerError)
			return c.SendString(err.Error())
		}
	}

	fmt.Printf("[USERS]GET: request for followers of %s\n", username)
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
	err := users.Follow(strings.Clone(username), strings.Clone(target))
	if err != nil {
		switch {
		case err.Error() == "try to self-follow":
			c.Status(fiber.StatusBadRequest)
			return c.SendString(err.Error())
		case err.Error() == "acting user not found":
		case err.Error() == "target user not found":
			c.Status(fiber.StatusNotFound)
			return c.SendString(err.Error())
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	fmt.Printf("[USERS]FOLLOW: %s follows %s\n", username, target)
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
	err := users.Unfollow(strings.Clone(username), strings.Clone(target))
	if err != nil {
		switch {
		case err.Error() == "try to self-follow":
			c.Status(fiber.StatusBadRequest)
			return c.SendString(err.Error())
		case err.Error() == "acting user not found":
		case err.Error() == "target user not found":
			c.Status(fiber.StatusNotFound)
			return c.SendString(err.Error())
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	fmt.Printf("[USERS]UNFOLLOW: %s unfollows %s\n", username, target)
	return c.SendStatus(fiber.StatusOK)
}
