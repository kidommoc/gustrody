package auth

import "github.com/kidommoc/gustrody/internal/utils"

const (
	ErrInvalid utils.ErrCode = iota
	ErrExpired
	ErrWrongSession
	ErrUserNotFound
	ErrWrongPassword
)
