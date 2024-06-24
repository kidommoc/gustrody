package router

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/files"
	"github.com/kidommoc/gustrody/internal/logging"
)

func routeFiles(router fiber.Router) {
	cfg := config.Get()
	router.Put("/img", mAuth, uploadImg)
	router.Static("/imgs", cfg.ImgDir)
}

func uploadImg(c *fiber.Ctx) error {
	c.Accepts("multipart/form-data")
	logger := logging.Get()
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	var b []byte
	if fh, e := c.FormFile("image"); e != nil {
		// handle error
		logger.Error(e.Error(), nil)
		c.Status(fiber.StatusBadRequest)
		c.SendString(e.Error())
		return nil
	} else {
		if f, e := fh.Open(); e != nil {
			// handle error
			logger.Error(e.Error(), nil)
			c.Status(fiber.StatusBadRequest)
			c.SendString(e.Error())
			return nil
		} else {
			if b, e = io.ReadAll(f); e != nil {
				// handle error
				logger.Error(e.Error(), nil)
				c.Status(fiber.StatusBadRequest)
				c.SendString(e.Error())
				return nil
			}
		}
	}

	fileService := files.NewFileService()
	path, mediaType, e := fileService.StoreImage(username, b)
	if e != nil {
		c.Status(http.StatusInternalServerError)
		return nil
	}
	msg := fmt.Sprintf("[FILE]UPLOAD: Image from %s. Path: %s", username, path)
	logger.Info(msg)
	c.Status(fiber.StatusOK)
	c.JSON(fiber.Map{
		"type": mediaType,
		"url":  path,
	})
	return nil
}
