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
	router.Put("/follow/:username", follow)
	router.Delete("/follow/:username", unfollow)
}

func getUserInfo(c *fiber.Ctx) error {
	username := c.Params("username")
	c.SendStatus(fiber.StatusOK)
}

func getUserPosts(c *fiber.Ctx) error {
	username := c.Params("username")
	c.SendStatus(fiber.StatusOK)
}

func getUserFollowings(c *fiber.Ctx) error {
	username := c.Params("username")
	c.SendStatus(fiber.StatusOK)
}

func getUserFollowers(c *fiber.Ctx) error {
	username := c.Params("username")
	c.SendStatus(fiber.StatusOK)
}

func follow(c *fiber.Ctx) error {
	target := c.Params("username")
	c.SendStatus(fiber.StatusOK)
}

func unfollow(c *fiber.Ctx) error {
	target := c.Params("username")
	c.SendStatus(fiber.StatusOK)
}