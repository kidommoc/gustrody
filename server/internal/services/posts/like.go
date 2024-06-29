package posts

import (
	"fmt"

	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/services/users"
)

func (service *PostService) GetLikes(user, postID string) (list []*users.UserInfo, err error) {
	logger := service.lg
	result, owner, vsb, e := service.db.Like.QueryLikes(postID)
	if e != nil && e != models.ErrNotFound {
		msg := fmt.Sprintf("[Posts.Like] Cannot get likes of %s", postID)
		logger.Error(msg, e)
		return nil, ErrInternal
	}

	if !service.checkPermission(user, owner, postID, vsb) {
		return nil, ErrNotPermitted
	}

	if e != nil {
		return list, ErrPostNotFound
	}

	us := make(map[string]*users.UserInfo)
	gu := func(u string) *users.UserInfo {
		if u == "" {
			return nil
		}
		if us[u] != nil {
			return us[u]
		}
		info, e := service.user.GetInfo(u)
		if e != nil {
			msg := fmt.Sprintf("[Posts.Like] Cannot get info of %s", u)
			logger.Error(msg, e)
			return nil
		}
		us[u] = &info
		return &info
	}
	list = make([]*users.UserInfo, 0, len(result))
	for _, u := range result {
		info := gu(u)
		if info != nil {
			list = append(list, info)
		}
	}

	return list, nil
}

func (service *PostService) Like(username, postID string) error {
	logger := service.lg
	if err := service.db.Like.SetLike(username, postID); err != nil {
		switch {
		case err == models.ErrNotFound:
			return ErrPostNotFound
		default:
			msg := fmt.Sprintf("[Posts.Like] Cannot set %s's like to %s", username, postID)
			logger.Error(msg, err)
			return ErrInternal
		}
	}
	return nil
}

func (service *PostService) Unlike(username, postID string) error {
	logger := service.lg
	if err := service.db.Like.RemoveLike(username, postID); err != nil {
		switch err {
		case models.ErrNotFound:
			switch err.Error() {
			case "post":
				return ErrPostNotFound
			case "like":
				return ErrLikeNotFound
			}
		default:
			msg := fmt.Sprintf("[Posts.Like] Cannot remove like of %s to %s", username, postID)
			logger.Error(msg, err)
			return ErrInternal
		}
	}
	return nil
}
