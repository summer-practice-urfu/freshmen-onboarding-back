package storages

type DoubleOperError struct {
	msg  string
	oper rune
}

func (e *DoubleOperError) Error() string {
	return e.msg + string(e.oper)
}

func NewDoubleOperError(oper rune, msg *string) *DoubleOperError {
	defaultMsg := "double oper "
	if msg == nil {
		msg = &defaultMsg
	}
	return &DoubleOperError{
		msg:  *msg,
		oper: oper,
	}
}
