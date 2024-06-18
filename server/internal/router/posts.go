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

func routePosts(router *gin.RouterGroup) {
	router.GET("/:postID", getPost)
	router.GET("/:postID/likes", getPostLikes)
	router.GET("/:postID/shares", getPostShares)
	router.PUT("/", mAuth, newPost)
	router.POST("/:postID", mAuth, editPost)
	router.DELETE("/:postID", mAuth, removePost)

	router.PUT("/:postID/like", mAuth, likePost)
	router.DELETE("/:postID/like", mAuth, unlikePost)
	router.PUT("/:postID/share", mAuth, sharePost)
	router.DELETE("/:postID/share", mAuth, unsharePost)
	router.PUT("/:postID/reply", mAuth, replyPost)
}

func getPost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.String(http.StatusBadRequest, "Acquire post id.")
		return
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	post, err := postService.Get(postID)
	if err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.String(http.StatusNotFound, "Post not found.")
			return
		case posts.ErrOwner:
			c.String(http.StatusNotFound, "Post owner not found.")
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]GET: request for %s", postID)
	logger.Info(msg)
	c.JSON(http.StatusOK, post)
}

func getPostLikes(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.String(http.StatusBadRequest, "Acquire post id.")
		return
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	list, err := postService.GetLikes(postID)
	if err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.String(http.StatusNotFound, "Post not found.")
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]GET: request for likes of %s", postID)
	logger.Info(msg)
	c.JSON(http.StatusOK, list)
}

func getPostShares(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.String(http.StatusBadRequest, "Acquire post id.")
		return
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	list, err := postService.GetShares(postID)
	if err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.String(http.StatusNotFound, "Post not found.")
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]GET: request for shares of %s", postID)
	logger.Info(msg)
	c.JSON(http.StatusOK, list)
}

type contentBody struct {
	Content string `json:"content"`
}

func newPost(c *gin.Context) {
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

	content := contentBody{}
	if err := c.ShouldBind(&content); err != nil {
		c.String(http.StatusBadRequest, "Acquire content to post.")
		return
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.New(strings.Clone(username), strings.Clone(content.Content)); err != nil {
		switch err.Code() {
		case posts.ErrUserNotFound:
			c.String(http.StatusNotFound, "Actor(User) not found.")
			return
		case posts.ErrContent:
			switch err.Error() {
			case "empty":
				c.String(http.StatusBadRequest, "Acquire content to post.")
				return
			case "long":
				c.String(http.StatusBadRequest, "Content is too long to post.")
				return
			}
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]NEW: %s posts a new post", username)
	logger.Info(msg)
	c.Status(http.StatusOK)
}

func editPost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.String(http.StatusBadRequest, "Acquire post id.")
		return
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
	content := contentBody{}
	if err := c.ShouldBind(&content); err != nil {
		c.String(http.StatusBadRequest, "Acquire content to post.")
		return
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.Edit(username, postID, strings.Clone(content.Content)); err != nil {
		switch err.Code() {
		case posts.ErrOwner:
			c.String(http.StatusForbidden, "Not post owner.")
			return
		case posts.ErrPostNotFound:
			c.String(http.StatusNotFound, "Post not found.")
			return
		case posts.ErrContent:
			switch err.Error() {
			case "empty":
				c.String(http.StatusBadRequest, "Acquire content to post.")
				return
			case "long":
				c.String(http.StatusBadRequest, "Content is too long to post.")
				return
			}
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]EDIT: %s edits %s", username, postID)
	logger.Info(msg)
	c.Status(http.StatusOK)
}

func removePost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.String(http.StatusBadRequest, "Acquire post id")
		return
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
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.Remove(username, postID); err != nil {
		switch err.Code() {
		case posts.ErrOwner:
			c.String(http.StatusForbidden, "Not post owner.")
			return
		case posts.ErrPostNotFound:
			c.String(http.StatusNotFound, "Post not found.")
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]REMOVE: %s removes %s", username, postID)
	logger.Info(msg)
	c.Status(http.StatusOK)
}

func likePost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.String(http.StatusBadRequest, "Acquire post id.")
		return
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
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.Like(strings.Clone(username), postID); err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.String(http.StatusNotFound, "Post not found.")
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]LIKE: %s likes %s", username, postID)
	logger.Info(msg)
	c.Status(http.StatusOK)
}

func unlikePost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.String(http.StatusBadRequest, "Acquire post id.")
		return
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
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.Unlike(username, postID); err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.String(http.StatusNotFound, "Post not found.")
			return
		case posts.ErrLikeNotFound:
			c.String(http.StatusNotFound, "Like not found.")
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]UNLIKE: %s unlikes %s", username, postID)
	logger.Info(msg)
	c.Status(http.StatusOK)
}

func sharePost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.String(http.StatusBadRequest, "Acquire post id.")
		return
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
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.Share(strings.Clone(username), postID); err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.String(http.StatusNotFound, "Post not found.")
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]SHARE: %s shares %s", username, postID)
	logger.Info(msg)
	c.Status(http.StatusOK)
}

func unsharePost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.String(http.StatusBadRequest, "Acquire post id.")
		return
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
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.Unshare(username, postID); err != nil {
		switch err.Code() {
		case posts.ErrPostNotFound:
			c.String(http.StatusNotFound, "Post not found.")
			return
		case posts.ErrShareNotFound:
			c.String(http.StatusNotFound, "Share not found.")
			return
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]UNSHARE: %s unshares %s", username, postID)
	logger.Info(msg)
	c.Status(http.StatusOK)
}

func replyPost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.String(http.StatusBadRequest, "Acquire post id")
		return
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

	content := contentBody{}
	if err := c.ShouldBind(content); err != nil {
		c.String(http.StatusBadRequest, "Acquire content to post.")
		return
	}

	userService := users.NewService(models.UserInstance())
	postService := posts.NewService(models.PostInstance(), userService)
	if err := postService.Reply(strings.Clone(username), postID, strings.Clone(content.Content)); err != nil {
		switch err.Code() {
		case posts.ErrUserNotFound:
			c.String(http.StatusNotFound, "Actor(User) not found.")
			return
		case posts.ErrPostNotFound:
			c.String(http.StatusNotFound, "Post not found.")
			return
		case posts.ErrContent:
			switch err.Error() {
			case "empty":
				c.String(http.StatusBadRequest, "Acquire content to post.")
				return
			case "long":
				c.String(http.StatusBadRequest, "Content is too long to post.")
				return
			}
		default:
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	logger := logging.Get()
	msg := fmt.Sprintf("[POSTS]REPLY: %s replies %s", username, postID)
	logger.Info(msg)
	c.Status(http.StatusOK)
}
