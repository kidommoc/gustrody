package users

import (
	"errors"

	"github.com/kidommoc/gustrody/internal/db"
)

var site = "localhost:8000" // should load from .env

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type UserProfile struct {
	UserInfo
	Bio        string `json:"bio"`
	Follows    uint   `json:"follows"`
	Followings string `json:"followings"`
	Followed   uint   `json:"followed"`
	Followers  string `json:"followers"`
}

func generateID(username string) string {
	return site + "/users/" + username
}

func GetProfile(username string) (info UserProfile, err error) {
	u, err := db.QueryUser(username)
	if err != nil {
		return info, errors.New("user not found")
	}
	info.Username = u.Username
	info.Bio = u.Bio
	if info.Follows, err = db.QueryUserFollows(username); err != nil {
		return info, errors.New("user not found")
	}
	if info.Followed, err = db.QueryUserFollowed(username); err != nil {
		return info, errors.New("user not found")
	}
	info.ID = site + "/users/" + username
	info.Followings = info.ID + "/followings"
	info.Followers = info.ID + "/followers"
	return info, nil
}

func GetFollowings(username string) (list []*UserInfo, err error) {
	list = make([]*UserInfo, 0)
	l, err := db.QueryUserFollowings(username)
	if err != nil {
		return list, errors.New("user not found")
	}
	for _, e := range l {
		list = append(list, &UserInfo{
			ID:       generateID(e.Username),
			Username: e.Username,
		})
	}
	return list, nil
}

func GetFollowers(username string) (list []*UserInfo, err error) {
	list = make([]*UserInfo, 0)
	l, err := db.QueryUserFollowers(username)
	if err != nil {
		return list, errors.New("user not found")
	}
	for _, e := range l {
		list = append(list, &UserInfo{
			ID:       generateID(e.Username),
			Username: e.Username,
		})
	}
	return list, nil
}
