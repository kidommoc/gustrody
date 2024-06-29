package db

import "errors"

var ErrNoConn = errors.New("NoConnection")
var ErrConnClosed = errors.New("ConnectionClosed")
var ErrNotFound = errors.New("NotFound")
var ErrDbInternal = errors.New("DbInternal")
