package models

import "errors"

var ErrNotFound = errors.New("NotFound")
var ErrDunplicate = errors.New("Dunplicate")
var ErrFormat = errors.New("Syntax")
var ErrDbInternal = errors.New("DbInternal")
var ErrMaxRetries = errors.New("MaxRetries")
var ErrNoEnoughPages = errors.New("NoEnoughPages")
var ErrInconsistent = errors.New("Inconsistent")
