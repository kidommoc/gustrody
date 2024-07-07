package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kidommoc/gustrody/internal/logging"
)

func routeFederal(router fiber.Router) {
	router.Get("/.well-known/webfinger", webfinger)
	router.Post("/inbox", sigVerify, inbox)
	router.Post("/users/:username/inbox", sigVerify, inbox)
}

func sigVerify(c *fiber.Ctx) error {
	return c.Next()
}

func webfinger(c *fiber.Ctx) error {
	logger := logging.Get()
	logger.Info("[FEDERAL] webfinger.")
	return c.SendStatus(fiber.StatusOK)
}

func inbox(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func getUserProfileFederal(c *fiber.Ctx) error {
	logger := logging.Get()
	logger.Info("[FEDERAL] get profile")
	return c.SendStatus(fiber.StatusOK)
}

func getUserFollowingsFederal(c *fiber.Ctx) error {
	logger := logging.Get()
	logger.Info("[FEDERAL] get followings")
	return c.SendStatus(fiber.StatusOK)
}

func getUserFollowersFederal(c *fiber.Ctx) error {
	logger := logging.Get()
	logger.Info("[FEDERAL] get followers")
	return c.SendStatus(fiber.StatusOK)
}

func getPostFederal(c *fiber.Ctx) error {
	logger := logging.Get()
	logger.Info("[FEDERAL] get post")
	return c.SendStatus(fiber.StatusOK)
}
