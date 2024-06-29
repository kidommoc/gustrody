package users

func (service *UserService) IsUserExist(username string) bool {
	return service.db.Info.IsUserExist(username)
}

func (service *UserService) IsFollowing(username, target string) bool {
	return service.db.Follow.IsFollowing(username, target)
}

func (service *UserService) GetInfo(username string) (info UserInfo, err error) {
	logger := service.lg
	u, e := service.db.Info.QueryUser(username)
	if e != nil {
		logger.Error("[User] when GetInfo", e)
		return info, ErrUserNotFound
	}
	info.ID = service.generateID(u.Username)
	info.Username = u.Username
	info.Nickname = u.Nickname
	return info, nil
}

func (service *UserService) GetProfile(username string) (info UserProfile, err error) {
	logger := service.lg
	u, e := service.db.Info.QueryUser(username)
	if e != nil {
		logger.Error("[Users] Error when GetProfile", e)
		return info, ErrUserNotFound
	}
	info = UserProfile{
		UserInfo: UserInfo{
			ID:       service.generateID(u.Username),
			Username: u.Username,
			Nickname: u.Nickname,
		},
		Summary: u.Summary,
	}
	info.Follows, info.Followed, e = service.db.Follow.QueryUserFollowInfo(username)
	if e != nil {
		logger.Error("[Users] Error when GetProfile", e)
	}
	return info, nil
}

func (service *UserService) GetFollowings(username string) (list []*UserInfo, err error) {
	logger := service.lg
	l, e := service.db.Follow.QueryUserFollowings(username)
	if e != nil {
		logger.Error("[User] when GetFollowings", e)
		return list, ErrUserNotFound
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

func (service *UserService) GetFollowers(username string) (list []*UserInfo, err error) {
	logger := service.lg
	l, e := service.db.Follow.QueryUserFollowers(username)
	if e != nil {
		logger.Error("[User] when GetFollowers", e)
		return list, ErrUserNotFound
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
