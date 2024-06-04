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
