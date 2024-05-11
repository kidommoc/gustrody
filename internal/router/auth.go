package router

import (
	"fmt"
	"strings"

	"github.com/kidommoc/gustrody/internal/auth"

	"github.com/gofiber/fiber/v2"
)

func routeAuth(router fiber.Router) {
	router.Post("/login", login)
	router.Post("/token", mAuth, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
}

func mAuth(c *fiber.Ctx) error {
	session := c.Get("Session")
	authroization := c.Get("Authorization")
	if session == "" || authroization == "" {
		c.Status(fiber.StatusUnauthorized)
		return c.SendString("missing http header: session and authorization")
	}
	// parse bearer to token
	bearer := strings.Split(authroization, " ")
	if len(bearer) != 2 || bearer[0] != "Bearer" {
		c.Status(fiber.StatusUnauthorized)
		return c.SendString("invalid header: authorization")
	}
	username, err := auth.VerifyToken(bearer[1], session)
	if err != nil {
		c.Status(fiber.StatusUnauthorized)
		return c.SendString(err.Error())
	}
	fmt.Printf("[AUTH]mAuth: %s at session %s\n", username, session)

	c.Locals("username", username)
	oauth := auth.NewOauth(username, session)
	c.Set("Token", oauth.Token)
	c.Set("Refresh", oauth.Refresh)
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
	fmt.Printf("[AUTH]LOGIN: username \"%s\", password \"%s\"\n", body.Username, body.Password)

	session, oauth, err := auth.Login(body.Username, body.Password)
	if err != nil {
		c.Status(401)
		return c.SendString(err.Error())
	}

	fmt.Printf("[AUTH]LOGIN: succeed. session: %s\n", session)
	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"session": session,
		"token":   oauth.Token,
		"refresh": oauth.Refresh,
	})
}
