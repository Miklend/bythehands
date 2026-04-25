package service

import (
	"errors"
	"fmt"
)

type ErrorKind string

const (
	KindValidation ErrorKind = "validation"
	KindNotFound   ErrorKind = "not_found"
	KindConflict   ErrorKind = "conflict"
	KindForbidden  ErrorKind = "forbidden"
	KindInternal   ErrorKind = "internal"
)

type Error struct {
	Kind    ErrorKind
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Kind, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Err)
}

func (e *Error) Unwrap() error { return e.Err }

func IsKind(err error, kind ErrorKind) bool {
	var se *Error
	if !errors.As(err, &se) {
		return false
	}
	return se.Kind == kind
}

func validation(msg string) error { return &Error{Kind: KindValidation, Message: msg} }
func notFound(msg string, err error) error {
	return &Error{Kind: KindNotFound, Message: msg, Err: err}
}
func conflict(msg string, err error) error {
	return &Error{Kind: KindConflict, Message: msg, Err: err}
}
func forbidden(msg string, err error) error {
	return &Error{Kind: KindForbidden, Message: msg, Err: err}
}
