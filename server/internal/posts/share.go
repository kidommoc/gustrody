package posts

import (
	"fmt"
	"time"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/users"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *PostService) GetShares(user string, postID string) (list []*users.UserInfo, err utils.Error) {
	result, owner, vsb, e := service.db.QueryShares(postID)
	if e != nil && e.Code() != models.ErrNotFound {
		logger := logging.Get()
		logger.Error("[Posts.Share]", err)
		return nil, newErr(ErrInternal)
	}

	if !service.checkPermission(user, owner, postID, vsb) {
		return nil, newErr(
			ErrNotPermitted,
			fmt.Sprintf("%s is not allowed to visit %s", user, postID),
		)
	}

	if e != nil {
		return list, newErr(ErrPostNotFound, postID)
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
			logger := logging.Get()
			logger.Error("[Posts.Like] Cannot get user info", e)
			return nil
		}
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

func (service *PostService) Share(username string, postID string, vsb string) utils.Error {
	v, ok := utils.GetVsb(vsb)
	if !ok {
		// get default
	}

	if err := service.db.SetShare(username, postID, time.Now(), v); err != nil {
		switch {
		case err.Code() == models.ErrNotFound:
			return newErr(ErrPostNotFound, postID)
		default:
			logger := logging.Get()
			logger.Error("[Posts.Share]", err)
			return newErr(ErrInternal)
		}
	}
	return nil
}

func (service *PostService) Unshare(username string, postID string) utils.Error {
	if err := service.db.RemoveShare(username, postID); err != nil {
		switch err.Code() {
		case models.ErrNotFound:
			switch err.Error() {
			case "post":
				return newErr(ErrPostNotFound, postID)
			case "share":
				return newErr(ErrShareNotFound, username)
			}
		default:
			logger := logging.Get()
			logger.Error("[Posts.Share]", err)
			return newErr(ErrInternal)
		}
	}
	return nil
}
