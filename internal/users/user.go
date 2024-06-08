package users

import (
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *UserService) IsUserExist(username string) bool {
	return service.db.IsUserExist(username)
}

func (service *UserService) GetInfo(username string) (info UserInfo, err utils.Err) {
	u, e := service.db.QueryUser(username)
	if e != nil {
		return info, utils.NewErr(ErrNotFound)
	}
	info.ID = service.generateID(u.Username)
	info.Username = u.Username
	info.Nickname = u.Nickname
	return info, nil
}

func (service *UserService) GetProfile(username string) (info UserProfile, err utils.Err) {
	u, e := service.db.QueryUser(username)
	if e != nil {
		return info, utils.NewErr(ErrNotFound)
	}
	info = UserProfile{
		UserInfo: UserInfo{
			ID:       service.generateID(u.Username),
			Username: u.Username,
			Nickname: u.Nickname,
		},
		Summary:    u.Summary,
		Followings: info.ID + "/followings",
		Followers:  info.ID + "/followers",
	}
	info.Follows, _ = service.db.QueryUserFollows(username)
	info.Followed, _ = service.db.QueryUserFollowed(username)
	return info, nil
}

func (service *UserService) GetFollowings(username string) (list []*UserInfo, err utils.Err) {
	list = make([]*UserInfo, 0)
	l, e := service.db.QueryUserFollowings(username)
	if e != nil {
		return list, utils.NewErr(ErrNotFound)
	}
	for _, u := range l {
		list = append(list, &UserInfo{
			ID:       service.generateID(u.Username),
			Username: u.Username,
			Nickname: u.Nickname,
		})
	}
	return list, nil
}

func (service *UserService) GetFollowers(username string) (list []*UserInfo, err utils.Err) {
	list = make([]*UserInfo, 0)
	l, e := service.db.QueryUserFollowers(username)
	if e != nil {
		return list, utils.NewErr(ErrNotFound)
	}
	for _, u := range l {
		list = append(list, &UserInfo{
			ID:       service.generateID(u.Username),
			Username: u.Username,
			Nickname: u.Nickname,
		})
	}
	return list, nil
}
