package users

import (
	"github.com/kidommoc/gustrody/internal/database"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *UserService) Follow(actor string, target string) utils.Err {
	if actor == target {
		return utils.NewErr(ErrSelfFollow, actor)
	}
	if err := service.db.SetFollow(actor, target); err != nil {
		switch {
		case err.Code() == database.ErrNotFound && err.Error() == "from":
			return utils.NewErr(ErrNotFound, "from "+actor)
		case err.Code() == database.ErrNotFound && err.Error() == "to":
			return utils.NewErr(ErrNotFound, "to "+target)
		default:
			return err
		}
	}
	return nil
}

func (service *UserService) Unfollow(actor string, target string) utils.Err {
	if actor == target {
		return utils.NewErr(ErrSelfFollow, actor)
	}
	if err := service.db.RemoveFollow(actor, target); err != nil {
		switch {
		case err.Code() == database.ErrNotFound && err.Error() == "from":
			return utils.NewErr(ErrNotFound, "from "+actor)
		case err.Code() == database.ErrNotFound && err.Error() == "to":
			return utils.NewErr(ErrNotFound, "to "+target)
		default:
			return err
		}
	}
	return nil
}
