package router

import (
	"fmt"
	"strings"

	"github.com/kidommoc/gustrody/internal/auth"

	"github.com/gofiber/fiber/v2"
)

func routeAuth(router fiber.Router) {
	router.Post("/login", login)
	router.Post("/token", refreshToken)
}

func Auth(c *fiber.Ctx) error {
	session := c.Get("Session")
	bearer := c.Get("Authorization")
	// parse bearer to token
	token := strings.SplitN(bearer, " ", 2)[1]
	username, err := auth.VerifyToken(token, session)
	if err != nil {
		// handle error
		c.Status(fiber.StatusUnauthorized)
		return c.SendString(err.Error())
	}
	c.Locals("username", username)
	return c.Next()
}

type loginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshBody struct {
	Session string `json:"session"`
	Refresh string `json:"refresh"`
}

func login(c *fiber.Ctx) error {
	body := new(loginBody)
	if err := c.BodyParser(body); err != nil {
		return c.SendStatus(400)
	}
	fmt.Printf("[AUTH]LOGIN: username \"%s\", password \"%s\"\n", body.Username, body.Password)

	session, oauth, err := auth.Login(body.Username, body.Password)
	if err != nil {
		// handle err
		c.Status(401)
		return c.SendString(err.Error())
	}

	fmt.Printf("LOGIN: succeed. session: %s\n", session)
	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"session": session,
		"token":   oauth.Token,
		"refresh": oauth.Refresh,
	})
}

func refreshToken(c *fiber.Ctx) error {
	body := new(refreshBody)
	if err := c.BodyParser(body); err != nil {
		return c.SendStatus(400)
	}
	fmt.Printf("[AUTH]REFRESH: session \"%s\", refresh token \"%s\"\n", body.Session, body.Refresh)

	oauth, err := auth.RefreshToken(body.Session, body.Refresh)
	if err != nil {
		// handle err
		c.Status(401)
		return c.SendString(err.Error())
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"token":   oauth.Token,
		"refresh": oauth.Refresh,
	})
}
