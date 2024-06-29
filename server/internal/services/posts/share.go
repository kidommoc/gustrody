package posts

import (
	"fmt"
	"time"

	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/services/users"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *PostService) GetShares(user, postID string) (list []*users.UserInfo, err error) {
	logger := service.lg
	result, owner, vsb, e := service.db.Share.QueryShares(postID)
	if e != nil && e != models.ErrNotFound {
		msg := fmt.Sprintf("[Posts.Share] Cannot get shares of %s", postID)
		logger.Error(msg, err)
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
			msg := fmt.Sprintf("[Posts.Share] Cannot get info of %s", u)
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

func (service *PostService) Share(username, postID string, vsb string) error {
	logger := service.lg
	v, ok := utils.GetVsb(vsb)
	if !ok {
		pf, err := service.user.GetPreferences(username)
		if err != nil {
			logger.Error("[Posts.Share] Cannot get user preferences.", err)
			return ErrInternal
		}
		v = pf.ShareVsb
	}

	if err := service.db.Share.SetShare(username, postID, time.Now(), v); err != nil {
		switch {
		case err == models.ErrNotFound:
			return ErrPostNotFound
		default:
			msg := fmt.Sprintf("[Posts.Share] Cannot set %s's share to %s", username, postID)
			logger.Error(msg, err)
			return ErrInternal
		}
	}
	return nil
}

func (service *PostService) Unshare(username, postID string) error {
	logger := service.lg
	if err := service.db.Share.RemoveShare(username, postID); err != nil {
		switch err {
		case models.ErrNotFound:
			switch err.Error() {
			case "post":
				return ErrPostNotFound
			case "share":
				return ErrShareNotFound
			}
		default:
			msg := fmt.Sprintf("[Posts.Share] Cannot remove share of %s to %s", username, postID)
			logger.Error(msg, err)
			return ErrInternal
		}
	}
	return nil
}
