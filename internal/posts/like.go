package posts

import (
	"github.com/kidommoc/gustrody/internal/database"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *PostService) Like(username string, postID string) utils.Err {
	postID = service.fullID(postID)
	if err := service.db.SetLike(username, postID); err != nil {
		switch {
		case err.Code() == database.ErrNotFound:
			return newErr(ErrPostNotFound, postID)
		}
	}
	return nil
}

func (service *PostService) Unlike(username string, postID string) utils.Err {
	postID = service.fullID(postID)
	if err := service.db.RemoveLike(username, postID); err != nil {
		switch err.Code() {
		case database.ErrNotFound:
			switch err.Error() {
			case "post":
				return newErr(ErrPostNotFound, postID)
			case "like":
				return newErr(ErrLikeNotFound, username)
			}
		}
	}
	return nil
}
