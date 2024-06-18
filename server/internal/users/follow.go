package users

import (
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *UserService) Follow(actor string, target string) utils.Error {
	if actor == target {
		return newErr(ErrSelfFollow, actor)
	}
	if err := service.db.SetFollow(actor, target); err != nil {
		switch {
		case err.Code() == models.ErrNotFound && err.Error() == "from":
			return newErr(ErrNotFound, "from "+actor)
		case err.Code() == models.ErrNotFound && err.Error() == "to":
			return newErr(ErrNotFound, "to "+target)
		default:
			return err
		}
	}
	return nil
}

func (service *UserService) Unfollow(actor string, target string) utils.Error {
	if actor == target {
		return newErr(ErrSelfFollow, actor)
	}
	if err := service.db.RemoveFollow(actor, target); err != nil {
		switch {
		case err.Code() == models.ErrNotFound && err.Error() == "from":
			return newErr(ErrNotFound, "from "+actor)
		case err.Code() == models.ErrNotFound && err.Error() == "to":
			return newErr(ErrNotFound, "to "+target)
		default:
			return err
		}
	}
	return nil
}
