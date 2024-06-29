package posts

import "errors"

var ErrPostNotFound = errors.New("PostNotFound")
var ErrUserNotFound = errors.New("UserNotFound")
var ErrLikeNotFound = errors.New("LikeNotFound")
var ErrShareNotFound = errors.New("ShareNotFound")
var ErrContentTooLong = errors.New("ContentTooLong")
var ErrContentEmpty = errors.New("ContentEmpty")
var ErrOwner = errors.New("Owner")
var ErrNotPermitted = errors.New("NotPermitted")
var ErrInternal = errors.New("Internal")
