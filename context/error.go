package context

import (
	"errors"
	"fmt"
)

type Error interface {
	error
	Code() int
}
type HydraError struct {
	code int
	error
}

func (a *HydraError) Code() int {
	return a.code
}

func NewError(code int, err ...interface{}) *HydraError {
	r := &HydraError{code: code}
	if len(err) == 0 {
		return r
	}
	if er, ok := err[0].(error); ok {
		r.error = er
		return r
	}
	r.error = errors.New(fmt.Sprint(err))
	return r
}
