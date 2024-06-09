package auth

import "github.com/kidommoc/gustrody/internal/utils"

const (
	ErrInvalid utils.ErrCode = iota
	ErrExpired
	ErrWrongSession
	ErrUserNotFound
	ErrWrongPassword
)

type AuthErr struct {
	utils.Err
}

func newErr(c utils.ErrCode, m ...string) AuthErr {
	return AuthErr{
		Err: utils.NewErr(c, m...),
	}
}

func (e AuthErr) CodeString() string {
	switch e.Code() {
	case ErrInvalid:
		return "Invalid"
	case ErrExpired:
		return "Expired"
	case ErrWrongSession:
		return "WrongSession"
	case ErrUserNotFound:
		return "UserNotFound"
	case ErrWrongPassword:
		return "WrongPassword"
	}
	return "Unknown"
}
