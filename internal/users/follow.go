package users

import (
	"github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/utils"
)

func Follow(actor string, target string) utils.Err {
	if actor == target {
		return utils.NewErr(ErrSelfFollow)
	}
	if err := db.SetFollow(actor, target); err != nil {
		switch {
		case err.Code() == db.ErrNotFound && err.Error() == "from":
			return utils.NewErr(ErrNotFound, "from")
		case err.Code() == db.ErrNotFound && err.Error() == "to":
			return utils.NewErr(ErrNotFound, "to")
		default:
			return err
		}
	}
	return nil
}

func Unfollow(actor string, target string) utils.Err {
	if actor == target {
		return utils.NewErr(ErrSelfFollow)
	}
	if err := db.UnsetFollow(actor, target); err != nil {
		switch {
		case err.Code() == db.ErrNotFound && err.Error() == "from":
			return utils.NewErr(ErrNotFound, "from")
		case err.Code() == db.ErrNotFound && err.Error() == "to":
			return utils.NewErr(ErrNotFound, "to")
		default:
			return err
		}
	}
	return nil
}
