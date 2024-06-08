package posts

import (
	"github.com/kidommoc/gustrody/internal/database"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *PostService) Share(username string, postID string) utils.Err {
	if err := service.db.SetShare(username, service.fullID(postID)); err != nil {
		switch {
		case err.Code() == database.ErrNotFound:
			return utils.NewErr(ErrPostNotFound)
		}
	}
	return nil
}

func (service *PostService) Unshare(username string, postID string) utils.Err {
	if err := service.db.RemoveShare(username, service.fullID(postID)); err != nil {
		switch err.Code() {
		case database.ErrNotFound:
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
