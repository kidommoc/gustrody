package users

import (
	"fmt"

	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/utils"
)

type Preferences struct {
	PostVsb  utils.Vsb `json:"postVsb"`
	ShareVsb utils.Vsb `json:"shareVsb"`
}

// DB: Account, Auth
func (service *UserService) Register(username, nickname, password string) error {
	logger := service.lg

	account := models.User{
		Username: username,
		Nickname: nickname,
		Keys:     models.KeyPair{},
	}
	account.Keys.Pub, account.Keys.Pri = utils.NewKeyPair()

	if err := service.db.Account.SetUser(&account); err != nil {
		switch err {
		case models.ErrDunplicate:
			return ErrExist
		default:
			msg := fmt.Sprintf("[Users.Account] Failed to register new user: %s.", username)
			logger.Error(msg, err)
			return ErrInternal
		}
	}

	password = string(utils.SHA256Hash(password))
	if err := service.db.Auth.SetUserPassword(username, password); err != nil {
		switch err {
		case models.ErrSyntax:
			return ErrSyntax
		default:
			msg := fmt.Sprintf("[Users.Account] Failed to update password of %s.", username)
			logger.Error(msg, err)
			return ErrInternal
		}
	}
	return nil
}

// DB: Info, Auth
func (service *UserService) UpdatePassword(username, newPwd string) error {
	logger := service.lg
	if !service.db.Info.IsUserExist(username) {
		return ErrUserNotFound
	}

	newPwd = string(utils.SHA256Hash(newPwd))
	if err := service.db.Auth.SetUserPassword(username, newPwd); err != nil {
		switch err {
		case models.ErrSyntax:
			return ErrSyntax
		default:
			msg := fmt.Sprintf("[Users.Account] Failed to update password of %s.", username)
			logger.Error(msg, err)
			return ErrInternal
		}
	}
	return nil
}

type ProfileBody struct {
	Nickname *string `json:"nickname,omitempty"`
	Summary  *string `json:"summary,omitempty"`
	Avatar   *struct {
		Type *string `json:"type,omitempty"`
		Url  *string `json:"url,omitempty"`
	} `json:"avatar,omitempty"`
}

// DB: Info
func (service *UserService) UpdateProfile(username string, body *ProfileBody) error {
	logger := service.lg

	pf, err := service.db.Info.QueryUser(username)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			return ErrUserNotFound
		default:
			msg := fmt.Sprintf("[Users.Account] Cannot get profile of %s.", username)
			logger.Error(msg, err)
			return ErrInternal
		}
	}

	if body.Nickname != nil {
		pf.Nickname = *body.Nickname
	}
	if body.Summary != nil {
		pf.Summary = *body.Summary
	}
	if body.Avatar != nil {
		// check type
		// check url

		if body.Avatar.Url != nil {
			pf.Avatar = *body.Avatar.Url
		}
	}

	if err := service.db.Info.UpdateUser(&pf); err != nil {
		msg := fmt.Sprintf("[Users.Account] Failed to update profile of %s.", username)
		logger.Error(msg, err)
		return ErrInternal
	}

	return nil
}

func (service *UserService) GetPreferences(username string) (pf Preferences, err error) {
	logger := service.lg
	mpf, err := service.db.Account.QueryUserPreferences(username)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			return pf, ErrUserNotFound
		default:
			msg := fmt.Sprintf("[Users.Account] Cannot get preferences of %s.", username)
			logger.Error(msg, err)
			return pf, ErrInternal
		}
	}
	pf.PostVsb, _ = utils.GetVsb(mpf.PostVsb)
	pf.ShareVsb, _ = utils.GetVsb(mpf.ShareVsb)
	return pf, nil
}

type PreferenceBody struct {
	PostVsb  *string `json:"postVsb,omitempty"`
	ShareVsb *string `json:"shareVsb,omitempty"`
}

func (service *UserService) UpdatePreferences(username string, body *PreferenceBody) error {
	logger := service.lg

	pf, err := service.db.Account.QueryUserPreferences(username)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			return ErrUserNotFound
		default:
			msg := fmt.Sprintf("[Users.Account] Cannot get preferences of %s.", username)
			logger.Error(msg, err)
			return ErrInternal
		}
	}

	if body.PostVsb != nil {
		v, ok := utils.GetVsb(*body.PostVsb)
		if ok {
			pf.PostVsb = v.String()
		}
	}
	if body.ShareVsb != nil {
		v, ok := utils.GetVsb(*body.ShareVsb)
		if ok {
			pf.ShareVsb = v.String()
		}
	}

	if err := service.db.Account.UpdateUserPreferences(username, pf); err != nil {
		msg := fmt.Sprintf("[Users.Account] Failed to update preferences of %s.", username)
		logger.Error(msg, err)
		return ErrInternal
	}
	return nil
}
