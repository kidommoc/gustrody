package utils

import "strconv"

type ErrCode uint

type Err interface {
	Code() ErrCode
	CodeString() string
	Error() string
}

type err struct {
	code ErrCode
	msg  string
}

func (e err) Error() string {
	return e.msg
}

func (e err) Code() ErrCode {
	return e.code
}

func (e err) CodeString() string {
	return strconv.Itoa(int(e.code))
}

func NewErr(code ErrCode, msg ...string) Err {
	if len(msg) == 0 {
		msg = []string{""}
	}
	return err{
		code: code,
		msg:  msg[0],
	}
}
