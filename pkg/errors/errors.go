package errors

import pkg "github.com/pkg/errors"

func New(message string) error {
	return pkg.New(message)
}

func Errorf(fmt string, args ...interface{}) error {
	return pkg.Errorf(fmt, args...)
}

func Wrap(err error, msg string) error {
	return pkg.Wrap(err, msg)
}

func Wrapf(err error, format string, args ...interface{}) error {
	return pkg.Wrapf(err, format, args...)
}

//type wrapper interface {
//	Unwrap() error
//}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//	type causer interface {
//	       Cause() error
//	}
//
// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(err error) error {
	return pkg.Cause(err)
}
