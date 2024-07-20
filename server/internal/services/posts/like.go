package posts

import (
	"fmt"

	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/services/users"
)

func (service *PostService) GetLikes(user, postID string) (list []*users.UserInfo, err error) {
	logger := service.lg

	post, err := service.db.Query.QueryPostByID(postID)
	switch err {
	case models.ErrNotFound:
		return nil, ErrPostNotFound
	case nil:
	default:
		msg := fmt.Sprintf("[Posts.Share] Cannot query %s", postID)
		logger.Error(msg, err)
		return nil, ErrInternal
	}
	if !service.checkPermission(user, &post) {
		return nil, ErrNotPermitted
	}

	result, err := service.db.Like.QueryLikes(postID)
	if err != nil && err != models.ErrNotFound {
		msg := fmt.Sprintf("[Posts.Like] Cannot get likes of %s", postID)
		logger.Error(msg, err)
		return nil, ErrInternal
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
		info := gu(u.String())
		if info != nil {
			list = append(list, info)
		}
	}

	return list, nil
}

func (service *PostService) Like(username, postID string) error {
	logger := service.lg
	if err := service.db.Like.SetLike(models.NewUD(username), postID); err != nil {
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
	if err := service.db.Like.RemoveLike(models.NewUD(username), postID); err != nil {
		switch err {
		case models.ErrNotFound:
			return ErrPostNotFound
		default:
			msg := fmt.Sprintf("[Posts.Like] Cannot remove like of %s to %s", username, postID)
			logger.Error(msg, err)
			return ErrInternal
		}
	}
	return nil
}
