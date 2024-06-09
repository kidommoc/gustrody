package database

import "github.com/kidommoc/gustrody/internal/utils"

const (
	ErrNotFound utils.ErrCode = iota
	ErrDunplicate
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
	case ErrNotFound:
		return "NotFound"
	case ErrDunplicate:
		return "Dunplicate"
	}
	return "Unknown"
}
