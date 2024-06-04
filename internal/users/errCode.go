package users

import "github.com/kidommoc/gustrody/internal/utils"

const (
	ErrNotFound utils.ErrCode = iota
	ErrSelfFollow
)
