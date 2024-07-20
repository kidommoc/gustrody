package router

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/services"
	"github.com/kidommoc/gustrody/internal/services/federal"
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
	acct := c.Query("resource")
	if acct == "" || !strings.HasPrefix(acct, "acct:") {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Wrong resource.")
	}
	username := strings.TrimPrefix(acct, "acct:")

	var federalService *federal.FederalService
	services.Get(reflect.ValueOf(&federalService).Elem())
	result, err := federalService.Webfinger(username)
	switch err {
	case federal.ErrNotFound:
		return c.SendStatus(fiber.StatusNotFound)
	case nil:
	}
	msg := fmt.Sprintf("[FEDERAL] Webfinger: %s.", username)
	logger.Info(msg)
	c.Set("Content-Type", "application/jrd+json")
	return c.JSON(result)
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
