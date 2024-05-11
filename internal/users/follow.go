package users

import (
	"errors"

	"github.com/kidommoc/gustrody/internal/db"
)

func Follow(actor string, target string) error {
	if actor == target {
		return errors.New("try to self-follow")
	}
	if err := db.SetFollow(actor, target); err != nil {
		switch {
		case err.Error() == "from not found":
			return errors.New("acting user not found")
		case err.Error() == "to not found":
			return errors.New("target user not found")
		default:
			return err
		}
	}
	return nil
}

func Unfollow(actor string, target string) error {
	if actor == target {
		return errors.New("try to self-unfollow")
	}
	if err := db.UnsetFollow(actor, target); err != nil {
		switch {
		case err.Error() == "from not found":
			return errors.New("acting user not found")
		case err.Error() == "to not found":
			return errors.New("target user not found")
		default:
			return err
		}
	}
	return nil
}
