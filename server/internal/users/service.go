package users

import (
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/models"
)

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type UserProfile struct {
	UserInfo
	Summary  string `json:"summary"`
	Follows  int64  `json:"follows"`
	Followed int64  `json:"followed"`
}

// user service

type UserService struct {
	site string
	db   models.IUserDb
}

func NewService(db models.IUserDb, c ...config.Config) *UserService {
	var cfg config.Config
	if len(c) == 0 {
		cfg = config.Get()
	} else {
		cfg = c[0]
	}
	return &UserService{
		site: cfg.Site,
		db:   db,
	}
}

func (service *UserService) generateID(username string) string {
	return service.site + "/users/" + username
}
