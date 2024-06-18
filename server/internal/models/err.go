package models

import "github.com/kidommoc/gustrody/internal/utils"

const (
	ErrNotFound utils.ErrCode = iota
	ErrDunplicate
	ErrDbInternal
	ErrSyntax
)

type ModelErr struct {
	utils.Err
}

func newErr(c utils.ErrCode, m ...string) ModelErr {
	return ModelErr{
		Err: utils.NewErr(c, m...),
	}
}

func (e ModelErr) CodeString() string {
	switch e.Code() {
	case ErrNotFound:
		return "NotFound"
	case ErrDunplicate:
		return "Dunplicate"
	case ErrDbInternal:
		return "DbInternal"
	case ErrSyntax:
		return "Syntax"
	}
	return "Unknown"
}
