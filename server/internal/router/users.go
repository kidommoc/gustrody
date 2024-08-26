package router

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/services"
	"github.com/kidommoc/gustrody/internal/services/posts"
	"github.com/kidommoc/gustrody/internal/services/users"
)

func routeUsers(router fiber.Router) {
	router.Put("/", registerUser)
	router.Post("/password", mAuth, changePassword)
	router.Get("/profile", mAuth, getUserProfile)
	router.Post("/profile", mAuth, editUserProfile)
	router.Get("/preferences", mAuth, getUserPreferences)
	router.Post("/preferences", mAuth, editUserPreferences)

	router.Get("/:username", switchByAccept(AcpMap{
		"application/json":          getUserProfile,
		"application/activity+json": getUserProfileFederal,
	}))
	router.Get("/:username/posts", func(c *fiber.Ctx) error {
		c.Locals("forced", false)
		return c.Next()
	}, mAuth, getUserPosts)
	router.Get("/:username/followings", switchByAccept(AcpMap{
		"application/json":          getUserFollowings,
		"application/activity+json": getUserFollowingsFederal,
	}))
	router.Get("/:username/followers", switchByAccept(AcpMap{
		"application/json":          getUserFollowers,
		"application/activity+json": getUserFollowersFederal,
	}))
	router.Put("/follow/:username", mAuth, follow)
	router.Delete("/follow/:username", mAuth, unfollow)
}

type registerBody struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

func registerUser(c *fiber.Ctx) error {
	body := new(registerBody)
	c.BodyParser(body)
	if body.Username == "" || body.Nickname == "" || body.Password == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Require username, nickname and password to register.")
	}

	var userService *users.UserService
	services.Get(reflect.ValueOf(&userService).Elem())
	if err := userService.Register(
		body.Username, body.Nickname, body.Password,
	); err != nil {
		switch err {
		// handle error
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]REGISTER: %s", body.Username)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}

func changePassword(c *fiber.Ctx) error {
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	body := new(registerBody)
	c.BodyParser(body)
	if body.Password == "" {
		c.Status(fiber.StatusBadRequest)
		return c.SendString("Require new password.")
	}

	var userService *users.UserService
	services.Get(reflect.ValueOf(&userService).Elem())
	if err := userService.UpdatePassword(username, body.Password); err != nil {
		switch err {
		// handle error
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]UPDATE: password of %s", username)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}

func editUserProfile(c *fiber.Ctx) error {
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	body := new(users.ProfileBody)
	c.BodyParser(body)

	var userService *users.UserService
	services.Get(reflect.ValueOf(&userService).Elem())
	if err := userService.UpdateProfile(username, body); err != nil {
		switch err {
		// handle error
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]EDIT: profile of %s", username)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}

func getUserPreferences(c *fiber.Ctx) error {
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	var userService *users.UserService
	services.Get(reflect.ValueOf(&userService).Elem())
	preferences, err := userService.GetPreferences(username)
	switch err {
	// handle error
	case nil:
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]GET: preferences of %s", username)
	logger.Info(msg)
	c.Status(fiber.StatusOK)
	return c.JSON(preferences)
}

func editUserPreferences(c *fiber.Ctx) error {
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	body := new(users.PreferenceBody)
	c.BodyParser(body)

	var userService *users.UserService
	services.Get(reflect.ValueOf(&userService).Elem())
	if err := userService.UpdatePreferences(username, body); err != nil {
		switch err {
		// handle error
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]EDIT: preferences of %s", username)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}

func getUserProfile(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire username")
	}

	var userService *users.UserService
	services.Get(reflect.ValueOf(&userService).Elem())
	profile, err := userService.GetProfile(username)
	switch err {
	case users.ErrUserNotFound:
		c.Status(fiber.StatusNotFound)
		return c.SendString("User not found.")
	case nil:
	default:
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]GET: request for %s", username)
	logger.Info(msg)
	c.Status(fiber.StatusOK)
	return c.JSON(profile)
}

func getUserPosts(c *fiber.Ctx) error {
	username, ok := c.Locals("username").(string)
	if !ok {
		username = ""
	}
	target := c.Params("username")
	if target == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire username")
	}
	maxDate := time.Now()
	mdstring := c.Params("maxdate")
	if mdstring != "" {
		if mdunix, err := strconv.ParseInt(mdstring, 0, 64); err == nil {
			maxDate = time.Unix(mdunix, 0)
		}
	}

	var postService *posts.PostService
	services.Get(reflect.ValueOf(&postService).Elem())
	list, err := postService.GetByUser(username, target, maxDate)
	switch err {
	case posts.ErrUserNotFound:
		c.Status(fiber.StatusNotFound)
		return c.SendString("User not found.")
	case nil:
	default:
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]GET: request for posts of %s", username)
	logger.Info(msg)
	c.Status(fiber.StatusOK)
	return c.JSON(list)
}

func getUserFollowings(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		c.Status(fiber.StatusBadRequest)
		c.SendString("Acquire username")
	}

	var userService *users.UserService
	services.Get(reflect.ValueOf(&userService).Elem())
	list, err := userService.GetFollowings(username)
	switch err {
	case users.ErrUserNotFound:
		c.Status(fiber.StatusNotFound)
		return c.SendString("User not found.")
	case nil:
	default:
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]GET: request for followings of %s", username)
	logger.Info(msg)
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

	var userService *users.UserService
	services.Get(reflect.ValueOf(&userService).Elem())
	list, err := userService.GetFollowers(username)
	switch err {
	case users.ErrUserNotFound:
		c.Status(fiber.StatusNotFound)
		return c.SendString("User not found.")
	case nil:
	default:
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]GET: request for followers of %s", username)
	logger.Info(msg)
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

	var userService *users.UserService
	services.Get(reflect.ValueOf(&userService).Elem())
	if err := userService.Follow(username, target); err != nil {
		switch err {
		case users.ErrSelfFollow:
			c.Status(fiber.StatusBadRequest)
			return c.SendString("Try to self-follow")
		case users.ErrFollowFromNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString(fmt.Sprintf("User not found: %s", username))
		case users.ErrFollowToNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString(fmt.Sprintf("User not found: %s", target))
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]FOLLOW: %s follows %s", username, target)
	logger.Info(msg)
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

	var userService *users.UserService
	services.Get(reflect.ValueOf(&userService).Elem())
	if err := userService.Unfollow(username, target); err != nil {
		switch err {
		case users.ErrSelfFollow:
			c.Status(fiber.StatusBadRequest)
			return c.SendString("Try to self-unfollow")
		case users.ErrFollowFromNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString(fmt.Sprintf("User not found: %s", username))
		case users.ErrFollowToNotFound:
			c.Status(fiber.StatusNotFound)
			return c.SendString(fmt.Sprintf("User not found: %s", target))
		default:
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[USERS]UNFOLLOW: %s unfollows %s", username, target)
	logger.Info(msg)
	return c.SendStatus(fiber.StatusOK)
}
