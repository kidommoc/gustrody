package posts

import (
	"github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/utils"
)

func Share(username string, postID string) utils.Err {
	if err := db.SetShare(username, fullID(postID)); err != nil {
		switch {
		case err.Code() == db.ErrNotFound:
			return utils.NewErr(ErrPostNotFound)
		}
	}
	return nil
}

func Unshare(username string, postID string) utils.Err {
	if err := db.RemoveShare(username, fullID(postID)); err != nil {
		switch err.Code() {
		case db.ErrNotFound:
			switch err.Error() {
			case "post":
				return utils.NewErr(ErrPostNotFound)
			case "share":
				return utils.NewErr(ErrShareNotFound)
			}
		}
	}
	return nil
}
