package federal

import "errors"

var ErrNotFound = errors.New("NotFound")
var ErrDbInternal = errors.New("DbInternal")
var ErrFile = errors.New("File")
var ErrSyntax = errors.New("Syntax")
var ErrRequest = errors.New("Request")
var ErrResponse = errors.New("Response")
