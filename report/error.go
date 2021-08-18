package report

import "fmt"

type Err interface {
	error
	ErrorFormat() string
}

type Errf struct {
	format string
	a      []interface{}
}

func (e *Errf) Error() string {
	return fmt.Sprintf(e.format, e.a...)
}

func (e *Errf) ErrorFormat() string {
	return e.format
}

func Errorf(format string, a ...interface{}) Err {
	return &Errf{
		format: format,
		a:      a,
	}
}
