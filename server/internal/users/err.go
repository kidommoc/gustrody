package users

import "github.com/kidommoc/gustrody/internal/utils"

const (
	ErrNotFound utils.ErrCode = iota
	ErrSelfFollow
)

type UserErr struct {
	utils.Err
}

func newErr(c utils.ErrCode, m ...string) UserErr {
	return UserErr{
		Err: utils.NewErr(c, m...),
	}
}

func (e UserErr) CodeString() string {
	switch e.Code() {
	case ErrNotFound:
		return "NotFound"
	case ErrSelfFollow:
		return "SelfFollow"
	}
	return "Unknown"
}
