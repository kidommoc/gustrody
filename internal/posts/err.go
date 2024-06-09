package posts

import "github.com/kidommoc/gustrody/internal/utils"

const (
	ErrPostNotFound utils.ErrCode = iota
	ErrUserNotFound
	ErrLikeNotFound
	ErrShareNotFound
	ErrContent
	ErrOwner
)

type PostErr struct {
	utils.Err
}

func newErr(c utils.ErrCode, m ...string) PostErr {
	return PostErr{
		Err: utils.NewErr(c, m...),
	}
}

func (e PostErr) CodeString() string {
	switch e.Code() {
	case ErrPostNotFound:
		return "PostNotFound"
	case ErrUserNotFound:
		return "UserNotFound"
	case ErrLikeNotFound:
		return "LikeNotFound"
	case ErrShareNotFound:
		return "ShareNotFound"
	case ErrContent:
		return "Content"
	case ErrOwner:
		return "Owner"
	}
	return "Unknown"
}
