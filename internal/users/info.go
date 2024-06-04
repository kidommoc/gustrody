package users

import (
	"github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/utils"
)

var site = "localhost:8000" // should load from .env

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"preferredUsername"`
	Nickname string `json:"name"`
}

type UserProfile struct {
	UserInfo
	Summary    string `json:"summary"`
	Follows    uint   `json:"follows"`
	Followings string `json:"followings"`
	Followed   uint   `json:"followed"`
	Followers  string `json:"followers"`
}

func generateID(username string) string {
	return site + "/users/" + username
}

func GetInfo(username string) (info UserInfo, err utils.Err) {
	u, e := db.QueryUser(username)
	if e != nil {
		return info, utils.NewErr(ErrNotFound)
	}
	info.ID = generateID(u.Username)
	info.Nickname = u.Nickname
	return info, nil
}

func GetProfile(username string) (info UserProfile, err utils.Err) {
	u, e := db.QueryUser(username)
	if e != nil {
		return info, utils.NewErr(ErrNotFound)
	}
	info.ID = generateID(u.Username)
	info.Username = u.Username
	info.Nickname = u.Nickname
	info.Summary = u.Summary
	if info.Follows, e = db.QueryUserFollows(username); e != nil {
		// should not reach here now
	}
	if info.Followed, e = db.QueryUserFollowed(username); e != nil {
		// should not reach here now
	}
	info.Followings = info.ID + "/followings"
	info.Followers = info.ID + "/followers"
	return info, nil
}

func GetFollowings(username string) (list []*UserInfo, err utils.Err) {
	list = make([]*UserInfo, 0)
	l, e := db.QueryUserFollowings(username)
	if e != nil {
		return list, utils.NewErr(ErrNotFound)
	}
	for _, u := range l {
		list = append(list, &UserInfo{
			ID:       generateID(u.Username),
			Username: u.Username,
			Nickname: u.Nickname,
		})
	}
	return list, nil
}

func GetFollowers(username string) (list []*UserInfo, err utils.Err) {
	list = make([]*UserInfo, 0)
	l, e := db.QueryUserFollowers(username)
	if e != nil {
		return list, utils.NewErr(ErrNotFound)
	}
	for _, u := range l {
		list = append(list, &UserInfo{
			ID:       generateID(u.Username),
			Username: u.Username,
			Nickname: u.Nickname,
		})
	}
	return list, nil
}
