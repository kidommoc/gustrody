package posts

import (
	"github.com/kidommoc/gustrody/internal/database"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *PostService) Like(username string, postID string) utils.Err {
	if err := service.db.SetLike(username, service.fullID(postID)); err != nil {
		switch {
		case err.Code() == database.ErrNotFound:
			return utils.NewErr(ErrPostNotFound)
		}
	}
	return nil
}

func (service *PostService) Unlike(username string, postID string) utils.Err {
	if err := service.db.RemoveLike(username, service.fullID(postID)); err != nil {
		switch err.Code() {
		case database.ErrNotFound:
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
