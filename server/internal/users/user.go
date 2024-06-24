package users

import (
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *UserService) IsUserExist(username string) bool {
	return service.db.IsUserExist(username)
}

func (service *UserService) IsFollowing(username string, target string) bool {
	return service.db.IsFollowing(username, target)
}

func (service *UserService) GetInfo(username string) (info UserInfo, err utils.Error) {
	logger := logging.Get()
	u, e := service.db.QueryUser(username)
	if e != nil {
		logger.Error("[User] when GetInfo", e)
		return info, newErr(ErrNotFound, username)
	}
	info.ID = service.generateID(u.Username)
	info.Username = u.Username
	info.Nickname = u.Nickname
	return info, nil
}

func (service *UserService) GetProfile(username string) (info UserProfile, err utils.Error) {
	logger := logging.Get()
	u, e := service.db.QueryUser(username)
	if e != nil {
		logger.Error("[Users] Error when GetProfile", e)
		return info, newErr(ErrNotFound, username)
	}
	info = UserProfile{
		UserInfo: UserInfo{
			ID:       service.generateID(u.Username),
			Username: u.Username,
			Nickname: u.Nickname,
		},
		Summary: u.Summary,
	}
	info.Follows, info.Followed, e = service.db.QueryUserFollowInfo(username)
	if e != nil {
		logger.Error("[Users] Error when GetProfile", e)
	}
	return info, nil
}

func (service *UserService) GetFollowings(username string) (list []*UserInfo, err utils.Error) {
	logger := logging.Get()
	l, e := service.db.QueryUserFollowings(username)
	if e != nil {
		logger.Error("[User] when GetFollowings", e)
		return list, newErr(ErrNotFound, username)
	}

	list = make([]*UserInfo, len(l))
	for i, u := range l {
		list[i] = &UserInfo{
			ID:       service.generateID(u.Username),
			Username: u.Username,
			Nickname: u.Nickname,
		}
	}
	return list, nil
}

func (service *UserService) GetFollowers(username string) (list []*UserInfo, err utils.Error) {
	logger := logging.Get()
	l, e := service.db.QueryUserFollowers(username)
	if e != nil {
		logger.Error("[User] when GetFollowers", e)
		return list, newErr(ErrNotFound, username)
	}

	list = make([]*UserInfo, len(l))
	for i, u := range l {
		list[i] = &UserInfo{
			ID:       service.generateID(u.Username),
			Username: u.Username,
			Nickname: u.Nickname,
		}
	}
	return list, nil
}
