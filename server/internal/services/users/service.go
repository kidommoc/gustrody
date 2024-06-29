package users

import (
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/logging"
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
	Summary     string `json:"summary"`
	Follows     int64  `json:"follows"`
	Followed    int64  `json:"followed"`
	PublicKey   string `json:"publicKey,omitempty"`
	PrivateKey  string `json:"privateKey,omitempty"`
	Preferences string `json:"preferences,omitempty"`
}

// user service

type UserDbs struct {
	Account models.IUserAccount
	Info    models.IUserInfo
	Follow  models.IUserFollow
	Auth    models.IAuthDb
}

type UserService struct {
	lg   logging.Logger
	site string
	db   UserDbs
}

func NewService(dbs UserDbs, cfg config.Config, lg logging.Logger) *UserService {
	return &UserService{
		lg:   lg,
		site: cfg.Site,
		db:   dbs,
	}
}

func (service *UserService) generateID(username string) string {
	return service.site + "/users/" + username
}
