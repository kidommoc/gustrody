package posts

import (
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/users"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *PostService) GetLikes(postID string) (list []*users.UserInfo, err utils.Error) {
	logger := logging.Get()
	postID = service.fullID(postID)
	result, e := service.db.QueryLikes(postID)
	if e != nil {
		switch {
		case e.Code() == models.ErrNotFound && e.Error() == "post":
			return list, newErr(ErrPostNotFound, postID)
			// default:
		default:
			logger.Error("[Posts.Like]", err)
			return nil, newErr(ErrInternal)
		}
	}

	list = make([]*users.UserInfo, 0, len(result))
	for i, u := range result {
		info, e := service.user.GetInfo(u)
		if e == nil {
			list[i] = &info
		} else {
			logger.Error("[Posts.Like] Cannot get user info", e)
		}
	}

	return list, nil
}

func (service *PostService) Like(username string, postID string) utils.Error {
	logger := logging.Get()
	postID = service.fullID(postID)
	if err := service.db.SetLike(username, postID); err != nil {
		switch {
		case err.Code() == models.ErrNotFound:
			return newErr(ErrPostNotFound, postID)
		default:
			logger.Error("[Posts.Like]", err)
			return newErr(ErrInternal)
		}
	}
	return nil
}

func (service *PostService) Unlike(username string, postID string) utils.Error {
	logger := logging.Get()
	postID = service.fullID(postID)
	if err := service.db.RemoveLike(username, postID); err != nil {
		switch err.Code() {
		case models.ErrNotFound:
			switch err.Error() {
			case "post":
				return newErr(ErrPostNotFound, postID)
			case "like":
				return newErr(ErrLikeNotFound, username)
			}
		default:
			logger.Error("[Posts.Like]", err)
			return newErr(ErrInternal)
		}
	}
	return nil
}
