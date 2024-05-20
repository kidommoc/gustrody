package utils

type Err interface {
	Code() uint
	Error() string
}

type err struct {
	code uint
	msg  string
}

func (e err) Error() string {
	return e.msg
}

func (e err) Code() uint {
	return e.code
}

func NewErr(code uint, msg ...string) Err {
	if len(msg) == 0 {
		msg = []string{""}
	}
	return err{
		code: code,
		msg:  msg[0],
	}
}
