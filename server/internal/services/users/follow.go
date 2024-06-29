package users

import (
	"github.com/kidommoc/gustrody/internal/models"
)

func (service *UserService) Follow(actor, target string) error {
	if actor == target {
		return ErrSelfFollow
	}
	if !service.db.Info.IsUserExist(actor) {
		return ErrFollowFromNotFound
	}
	if !service.db.Info.IsUserExist(target) {
		return ErrFollowToNotFound
	}
	logger := service.lg
	if err := service.db.Follow.SetFollow(actor, target); err != nil {
		switch err {
		case models.ErrDunplicate:
			return nil
		default:
			logger.Error("[Users.Follow] Db error", err)
			return ErrInternal
		}
	}
	return nil
}

func (service *UserService) Unfollow(actor, target string) error {
	if actor == target {
		return ErrSelfFollow
	}
	if !service.db.Info.IsUserExist(actor) {
		return ErrFollowFromNotFound
	}
	if !service.db.Info.IsUserExist(target) {
		return ErrFollowToNotFound
	}
	logger := service.lg
	if err := service.db.Follow.RemoveFollow(actor, target); err != nil {
		switch err {
		case models.ErrNotFound:
			return nil
		default:
			logger.Error("[Users.Follow] Db error", err)
			return ErrInternal
		}
	}
	return nil
}
