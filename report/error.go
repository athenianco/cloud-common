package report

import "fmt"

type Err interface {
	error
	ErrorFormat() string
}

type errf struct {
	format string
	a      []interface{}
}

func (e *errf) Error() string {
	return fmt.Errorf(e.format, e.a...).Error()
}

func (e *errf) ErrorFormat() string {
	return e.format
}

func Errorf(format string, a ...interface{}) Err {
	return &errf{
		format: format,
		a:      a,
	}
}
