package posts

import (
	"github.com/kidommoc/gustrody/internal/database"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *PostService) Share(username string, postID string) utils.Err {
	postID = service.fullID(postID)
	if err := service.db.SetShare(username, postID); err != nil {
		switch {
		case err.Code() == database.ErrNotFound:
			return newErr(ErrPostNotFound, postID)
		}
	}
	return nil
}

func (service *PostService) Unshare(username string, postID string) utils.Err {
	postID = service.fullID(postID)
	if err := service.db.RemoveShare(username, postID); err != nil {
		switch err.Code() {
		case database.ErrNotFound:
			switch err.Error() {
			case "post":
				return newErr(ErrPostNotFound, postID)
			case "share":
				return newErr(ErrShareNotFound, username)
			}
		}
	}
	return nil
}
