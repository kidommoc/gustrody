package posts

import (
	"github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/utils"
)

func Like(username string, postID string) utils.Err {
	if err := db.SetLike(username, fullID(postID)); err != nil {
		switch {
		case err.Code() == db.ErrNotFound:
			return utils.NewErr(ErrPostNotFound)
		}
	}
	return nil
}

func Unlike(username string, postID string) utils.Err {
	if err := db.RemoveLike(username, fullID(postID)); err != nil {
		switch err.Code() {
		case db.ErrNotFound:
			switch err.Error() {
			case "post":
				return utils.NewErr(ErrPostNotFound)
			case "like":
				return utils.NewErr(ErrLikeNotFound)
			}
		}
	}
	return nil
}
