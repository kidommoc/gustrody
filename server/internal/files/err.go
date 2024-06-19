package files

import "github.com/kidommoc/gustrody/internal/utils"

const (
	ErrFsInternal utils.ErrCode = iota
	ErrFile
)

type FileErr struct {
	utils.Err
}

func newErr(c utils.ErrCode, m ...string) FileErr {
	return FileErr{
		Err: utils.NewErr(c, m...),
	}
}

func (e FileErr) CodeString() string {
	switch e.Code() {
	case ErrFsInternal:
		return "FsInternal"
	case ErrFile:
		return "File"
	}
	return "Unknown"
}
