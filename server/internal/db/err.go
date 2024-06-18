package db

import "github.com/kidommoc/gustrody/internal/utils"

const (
	ErrNoConn utils.ErrCode = iota
	ErrConnClosed
	ErrDbInternal
	ErrNotFound
)

type DbErr struct {
	utils.Err
}

func newErr(c utils.ErrCode, m ...string) DbErr {
	return DbErr{
		Err: utils.NewErr(c, m...),
	}
}

func (e DbErr) CodeString() string {
	switch e.Code() {
	case ErrNoConn:
		return "NoConnection"
	case ErrConnClosed:
		return "ConnectionClosed"
	case ErrDbInternal:
		return "DatabaseInternalError"
	case ErrNotFound:
		return "NotFound"
	}
	return "Unknown"
}
