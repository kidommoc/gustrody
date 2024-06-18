package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/posts"
	"github.com/kidommoc/gustrody/internal/users"
)

func routeUsers(router *gin.RouterGroup) {
	router.GET("/:username", getUserProfile)
	router.GET("/:username/posts", getUserPosts)
	router.GET("/:username/followings", getUserFollowings)
	router.GET("/:username/followers", getUserFollowers)
	router.PUT("/follow/:username", mAuth, follow)
	router.DELETE("/follow/:username", mAuth, unfollow)
}

func getUserProfile(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.String(http.StatusBadRequest, "Acquire username")
		return
	}

	userService := users.NewService(models.UserInstance())
	profile, err := userService.GetProfile(username)
	if err != nil {
		switch err.Code() {
		case users.ErrNotFound:
			c.String(http.StatusNotFound, "User not found.")
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]GET: request for %s", username)
	logger.Info(msg)
	c.JSON(http.StatusOK, profile)
}

func getUserPosts(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.String(http.StatusBadRequest, "Acquire username")
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	list, err := postService.GetByUser(username)
	if err != nil {
		switch err.Code() {
		case posts.ErrUserNotFound:
			c.String(http.StatusNotFound, "User not found.")
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]GET: request for posts of %s", username)
	logger.Info(msg)
	c.JSON(http.StatusOK, list)
}

func getUserFollowings(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.String(http.StatusBadRequest, "Acquire username")
	}

	userService := users.NewService(models.UserInstance())
	list, err := userService.GetFollowings(username)
	if err != nil {
		switch err.Code() {
		case users.ErrNotFound:
			c.String(http.StatusNotFound, "User not found.")
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]GET: request for followings of %s", username)
	logger.Info(msg)
	c.JSON(http.StatusOK, gin.H{
		"list": list,
	})
}

func getUserFollowers(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.String(http.StatusBadRequest, "Acquire username")
	}

	userService := users.NewService(models.UserInstance())
	list, err := userService.GetFollowers(username)
	if err != nil {
		switch err.Code() {
		case users.ErrNotFound:
			c.String(http.StatusNotFound, "User not found.")
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]GET: request for followers of %s", username)
	logger.Info(msg)
	c.JSON(http.StatusOK, gin.H{
		"list": list,
	})
}

func follow(c *gin.Context) {
	target := c.Param("username")
	if target == "" {
		c.String(http.StatusBadRequest, "Acquire target username")
	}

	var username string
	if u, ok := c.Get("username"); !ok {
		c.Status(http.StatusUnauthorized)
		return
	} else {
		username, ok = u.(string)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	userService := users.NewService(models.UserInstance())
	err := userService.Follow(strings.Clone(username), strings.Clone(target))
	if err != nil {
		switch err.Code() {
		case users.ErrSelfFollow:
			c.String(http.StatusBadRequest, "Try to self-follow")
			return
		case users.ErrNotFound:
			c.String(http.StatusNotFound, fmt.Sprintf("User not found: %s", err.Error()))
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]FOLLOW: %s follows %s", username, target)
	logger.Info(msg)
	c.Status(http.StatusOK)
}

func unfollow(c *gin.Context) {
	target := c.Param("username")
	if target == "" {
		c.String(http.StatusBadRequest, "Acquire target username")
	}
	var username string
	if u, ok := c.Get("username"); !ok {
		c.Status(http.StatusUnauthorized)
		return
	} else {
		username, ok = u.(string)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	userService := users.NewService(models.UserInstance())
	err := userService.Unfollow(strings.Clone(username), strings.Clone(target))
	if err != nil {
		switch err.Code() {
		case users.ErrSelfFollow:
			c.String(http.StatusBadRequest, "Try to self-unfollow")
			return
		case users.ErrNotFound:
			c.String(http.StatusNotFound, fmt.Sprintf("User not found: %s", err.Error()))
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]UNFOLLOW: %s unfollows %s", username, target)
	logger.Info(msg)
	c.Status(http.StatusOK)
}
