package models

import "errors"

var ErrNotFound = errors.New("NotFound")
var ErrDuplicated = errors.New("Duplicated")
var ErrFormat = errors.New("Format")
var ErrDbInternal = errors.New("DbInternal")
var ErrMaxRetries = errors.New("MaxRetries")
var ErrNoEnoughPages = errors.New("NoEnoughPages")
var ErrInconsistent = errors.New("Inconsistent")
