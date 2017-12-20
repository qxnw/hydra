package context

import (
	"errors"
	"fmt"
)

type StateError struct {
	code int
	error
}

func (a *StateError) Code() int {
	return a.code
}

func NewError(code int, err ...interface{}) *StateError {
	r := &StateError{code: code}
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
