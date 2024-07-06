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

	// !!may update foreign

	post, err := service.db.Query.QueryPostByID(postID)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			return nil, ErrPostNotFound
		default:
			msg := fmt.Sprintf("[Posts.Share] Cannot query %s", postID)
			logger.Error(msg, err)
			return nil, ErrInternal
		}
	}
	if !service.checkPermission(user, &post) {
		return nil, ErrNotPermitted
	}

	result, err := service.db.Share.QueryShares(postID)
	if err != nil && err != models.ErrNotFound {
		msg := fmt.Sprintf("[Posts.Share] Cannot get shares of %s", postID)
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
			msg := fmt.Sprintf("[Posts.Share] Cannot get info of %s", u)
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

func (service *PostService) Share(username, postID string, date time.Time, vsb string) error {
	logger := service.lg
	v, ok := utils.GetVsb(vsb)
	if !utils.UdReg.MatchString(username) {
		// local
		if !ok {
			pf, err := service.user.GetPreferences(username)
			if err != nil {
				logger.Error("[Posts.Share] Cannot get user preferences.", err)
				return ErrInternal
			}
			v = pf.ShareVsb
		}
	}

	if err := service.db.Share.SetShare(models.NewUD(username), postID, date, v); err != nil {
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
	if err := service.db.Share.RemoveShare(models.NewUD(username), postID); err != nil {
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
