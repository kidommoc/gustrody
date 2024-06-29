package users

import "errors"

var ErrUserNotFound = errors.New("UserNotFound")
var ErrExist = errors.New("Exist")
var ErrSyntax = errors.New("Syntax")
var ErrFollowFromNotFound = errors.New("FollowFromNotFound")
var ErrFollowToNotFound = errors.New("FollowToNotFound")
var ErrSelfFollow = errors.New("SelfFollow")
var ErrInternal = errors.New("Internal")
