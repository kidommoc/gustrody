package users

import (
	"github.com/kidommoc/gustrody/internal/database"
)

var site = "127.0.0.1:8000" // should load from .env

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

// user service

type UserService struct {
	db database.IUsersDb
}

func NewService(db database.IUsersDb) *UserService {
	return &UserService{
		db: db,
	}
}
