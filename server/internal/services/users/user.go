package users

import (
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *UserService) IsUserExist(username string) bool {
	return service.db.Info.IsUserExist(username)
}

func (service *UserService) IsFollowing(username, target string) bool {
	return service.db.Follow.IsFollowing(models.NewUD(username), models.NewUD(target))
}

func (service *UserService) GetInfo(username string) (info UserInfo, err error) {
	logger := service.lg
	if utils.UdReg.MatchString(username) {
		// !!foreign
	} else {
		// local
		u, e := service.db.Info.QueryUser(username)
		if e != nil {
			logger.Error("[User] when GetInfo", e)
			return info, ErrUserNotFound
		}
		info.ID = utils.GenerateUserID(u.Username.Username, service.site)
		info.Username = u.Username.Username
		info.Nickname = u.Nickname
		// Avatar
	}
	return info, nil
}

func (service *UserService) GetProfile(username string) (pf *UserProfile, err error) {
	logger := service.lg
	if utils.UdReg.MatchString(username) {
		// !!foreign
	} else {
		// local
		u, e := service.db.Info.QueryUser(username)
		if e != nil {
			logger.Error("[Users] Error when GetProfile", e)
			return pf, ErrUserNotFound
		}
		pf = &UserProfile{
			UserInfo: UserInfo{
				ID:       utils.GenerateUserID(u.Username.Username, service.site),
				Username: u.Username.Username,
				Nickname: u.Nickname,
				// Avatar
			},
			Summary: u.Summary,
		}
		pf.Follows, pf.Followed, e = service.db.Follow.QueryUserFollowInfo(username)
		if e != nil {
			logger.Error("[Users] Error when GetProfile", e)
		}
	}
	return pf, nil
}

func (service *UserService) GetFollowings(username string) (list []*UserInfo, err error) {
	logger := service.lg
	if utils.UdReg.MatchString(username) {
		// !!foreign
	} else {
		// local
		l, e := service.db.Follow.QueryUserFollowings(username)
		if e != nil {
			logger.Error("[User] when GetFollowings", e)
			return list, ErrUserNotFound
		}

		list = make([]*UserInfo, 0, len(l))
		for _, u := range l {
			info, err := service.GetInfo(u.String())
			if err != nil {
				// handle error
				continue
			}
			list = append(list, &info)
		}
	}
	return list, nil
}

func (service *UserService) GetFollowers(username string) (list []*UserInfo, err error) {
	logger := service.lg
	if utils.UdReg.MatchString(username) {
		// foreign
	} else {
		// local
		l, e := service.db.Follow.QueryUserFollowers(username)
		if e != nil {
			logger.Error("[User] when GetFollowers", e)
			return list, ErrUserNotFound
		}

		list = make([]*UserInfo, 0, len(l))
		for _, u := range l {
			info, err := service.GetInfo(u.String())
			if err != nil {
				// handle error
				continue
			}
			list = append(list, &info)
		}
	}
	return list, nil
}
