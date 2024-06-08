package users

import (
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/database"
)

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

// user service

type UserService struct {
	site string
	db   database.IUsersDb
}

func NewService(db database.IUsersDb) *UserService {
	cfg := config.Get()
	return &UserService{
		site: cfg.Site,
		db:   db,
	}
}

func (service *UserService) generateID(username string) string {
	return service.site + "/users/" + username
}
