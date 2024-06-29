package models

import "errors"

var ErrNotFound = errors.New("NotFound")
var ErrDunplicate = errors.New("Dunplicate")
var ErrSyntax = errors.New("Syntax")
var ErrDbInternal = errors.New("DbInternal")
