package utils

type ErrCode uint

type Error interface {
	Code() ErrCode
	Error() string
	CodeString() string
}

type Err struct {
	code ErrCode
	msg  string
}

func (e Err) Error() string {
	return e.msg
}

func (e Err) Code() ErrCode {
	return e.code
}

func NewErr(code ErrCode, msg ...string) Err {
	if len(msg) == 0 {
		msg = []string{""}
	}
	return Err{
		code: code,
		msg:  msg[0],
	}
}
