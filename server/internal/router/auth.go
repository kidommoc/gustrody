package router

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/services"
	"github.com/kidommoc/gustrody/internal/services/auth"

	"github.com/gofiber/fiber/v2"
)

func routeAuth(router fiber.Router) {
	router.Post("/login", login)
	router.Post("/token", mAuth, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
}

func mAuth(c *fiber.Ctx) error {
	forced, ok := c.Locals("forced").(bool)
	if !ok {
		forced = true
	}
	session := c.Get("Session")
	authroization := c.Get("Authorization")
	if session == "" || authroization == "" {
		if forced {
			c.Status(fiber.StatusUnauthorized)
			return c.SendString("Missing HTTP header: Session and/or Authorization")
		} else {
			return c.Next()
		}
	}
	// parse bearer to token
	bearer := strings.Split(authroization, " ")
	if len(bearer) != 2 || bearer[0] != "Bearer" {
		c.Status(fiber.StatusUnauthorized)
		return c.SendString("Invalid header: Authorization.\nShould be Bearer token in JWT.")
	}

	var authService *auth.OauthService
	err := services.Get(reflect.ValueOf(&authService).Elem())
	if err != nil {
		// ?
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	username, err := authService.VerifyToken(bearer[1], session)
	if err != nil {
		c.Status(fiber.StatusUnauthorized)
		switch err {
		case auth.ErrExpired:
			return c.SendString("Token expired.")
		case auth.ErrInvalid:
			return c.SendString("Invalid token.")
		case auth.ErrWrongSession:
			return c.SendString("Session can't match.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	c.Locals("username", username)
	oauth := auth.NewOauth(username, session)
	c.Set("Token", oauth.Token)
	c.Set("Refresh", oauth.Refresh)
	// logger := logging.Get()
	// msg := fmt.Sprintf("[AUTH]MAUTH: %s at session %s", username, session)
	// logger.Info(msg)
	return c.Next()
}

type loginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func login(c *fiber.Ctx) error {
	body := new(loginBody)
	if err := c.BodyParser(body); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	var authService *auth.OauthService
	err := services.Get(reflect.ValueOf(&authService).Elem())
	if err != nil {
		// ?
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	session, oauth, err := authService.Login(body.Username, body.Password)
	if err != nil {
		c.Status(fiber.StatusUnauthorized)
		switch err {
		case auth.ErrUserNotFound:
			return c.SendString("User not found.")
		case auth.ErrWrongPassword:
			return c.SendString("Incorrect password.")
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[AUTH]LOGIN: %s succeed. session: %s", body.Username, session)
	logger.Info(msg)
	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"session": session,
		"token":   oauth.Token,
		"refresh": oauth.Refresh,
	})
}
