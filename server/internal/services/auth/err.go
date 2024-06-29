package auth

import "errors"

var ErrInvalid = errors.New("Invalid")
var ErrExpired = errors.New("Expired")
var ErrWrongSession = errors.New("WrongSession")
var ErrWrongPassword = errors.New("WrongPassword")
var ErrUserNotFound = errors.New("UserNotFound")
