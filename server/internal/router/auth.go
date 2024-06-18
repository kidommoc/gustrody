package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kidommoc/gustrody/internal/auth"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/models"
)

func routeAuth(router *gin.RouterGroup) {
	router.POST("/login", login)
	router.POST("/token", mAuth, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
}

func mAuth(c *gin.Context) {
	session := c.Request.Header.Get("Session")
	authroization := c.Request.Header.Get("Authorization")
	if session == "" || authroization == "" {
		c.String(http.StatusUnauthorized, "Missing HTTP header: Session and/or Authorization")
		return
	}
	// parse bearer to token
	bearer := strings.Split(authroization, " ")
	if len(bearer) != 2 || bearer[0] != "Bearer" {
		c.String(http.StatusUnauthorized, "Invalid header: Authorization.\nShould be Bearer token in JWT.")
		return
	}

	db := models.AuthInstance()
	authService := auth.NewService(db)
	username, err := authService.VerifyToken(bearer[1], session)
	if err != nil {
		switch err.Code() {
		case auth.ErrExpired:
			c.String(http.StatusUnauthorized, "Token expired.")
			return
		case auth.ErrInvalid:
			c.String(http.StatusUnauthorized, "Invalid token.")
			return
		case auth.ErrWrongSession:
			c.String(http.StatusUnauthorized, "Session can't match.")
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.Set("username", username)
	oauth := auth.NewOauth(username, session)
	c.Set("Token", oauth.Token)
	c.Set("Refresh", oauth.Refresh)
	// logger := logging.Get()
	// msg := fmt.Sprintf("[AUTH]MAUTH: %s at session %s", username, session)
	// logger.Info(msg)
	c.Next()
}

type loginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func login(c *gin.Context) {
	body := new(loginBody)
	if err := c.ShouldBind(&body); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	db := models.AuthInstance()
	authService := auth.NewService(db)
	session, oauth, err := authService.Login(body.Username, body.Password)
	if err != nil {
		switch err.Code() {
		case auth.ErrUserNotFound:
			c.String(http.StatusUnauthorized, "User not found.")
		case auth.ErrWrongPassword:
			c.String(http.StatusUnauthorized, "Incorrect password.")
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[AUTH]LOGIN: %s succeed. session: %s", body.Username, session)
	logger.Info(msg)
	c.JSON(http.StatusOK, gin.H{
		"session": session,
		"token":   oauth.Token,
		"refresh": oauth.Refresh,
	})
}
