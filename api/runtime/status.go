package runtime

import "fmt"

type status struct {
	Code    int32
	Message string
}

type WarpError struct {
	e *status
}

func new(c int32, msg string) *status {
	return &status{Code: c, Message: msg}
}

func (e *WarpError) Error() string {
	return e.e.Message
}

func (e *WarpError) Code() int32 {
	return e.e.Code
}

func (s *status) Err() error {
	return &WarpError{e: s}
}

func Error(c int32, msg string) error {
	return new(c, msg).Err()
}

func Errorf(c int32, format string, a ...interface{}) error {
	return Error(c, fmt.Sprintf(format, a...))
}

func Code(c int32) error {
	return Error(c, "")
}
