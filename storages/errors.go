package storages

type DoubleOperError struct {
	msg  string
	Data any
	oper rune
}

func (e DoubleOperError) Error() string {
	return e.msg + string(e.oper)
}

func NewDoubleOperError(oper rune, msg *string, data any) DoubleOperError {
	defaultMsg := "double oper "
	if msg == nil {
		msg = &defaultMsg
	}
	return DoubleOperError{
		msg:  *msg,
		oper: oper,
		Data: data,
	}
}

type InvalidOperError struct {
	msg  string
	oper rune
}

func (e InvalidOperError) Error() string {
	return e.msg + " " + string(e.oper)
}

func NewInvalidOperError(oper rune, msg *string) InvalidOperError {
	err := InvalidOperError{
		oper: oper,
		msg:  "invalid operation",
	}
	if msg == nil {
		err.msg = *msg
	}
	return err
}
