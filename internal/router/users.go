package router

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func routeUsers(router fiber.Router) {
	router.Get("/:username", getUserInfo)
	router.Get("/:username/posts", getUserPosts)
	router.Get("/:username/followings", getUserFollowings)
	router.Get("/:username/followers", getUserFollowers)
	router.Put("/follow/:username", Auth, follow)
	router.Delete("/follow/:username", Auth, unfollow)
}

func getUserInfo(c *fiber.Ctx) error {
	username := c.Params("username")
	fmt.Printf("[USERS]GET: request for %s\n", username)
	return c.SendStatus(fiber.StatusOK)
}

func getUserPosts(c *fiber.Ctx) error {
	username := c.Params("username")
	fmt.Printf("[USERS]GET: request for posts of %s\n", username)
	return c.SendStatus(fiber.StatusOK)
}

func getUserFollowings(c *fiber.Ctx) error {
	username := c.Params("username")
	fmt.Printf("[USERS]GET: request for followings of %s\n", username)
	return c.SendStatus(fiber.StatusOK)
}

func getUserFollowers(c *fiber.Ctx) error {
	username := c.Params("username")
	fmt.Printf("[USERS]GET: request for followers of %s\n", username)
	return c.SendStatus(fiber.StatusOK)
}

func follow(c *fiber.Ctx) error {
	target := c.Params("username")
	username := c.Locals("username")
	fmt.Printf("[USERS]FOLLOW: %s follows %s\n", username, target)
	return c.SendStatus(fiber.StatusOK)
}

func unfollow(c *fiber.Ctx) error {
	target := c.Params("username")
	username := c.Locals("username")
	fmt.Printf("[USERS]UNFOLLOW: %s unfollows %s\n", username, target)
	return c.SendStatus(fiber.StatusOK)
}
