package core

import "fmt"

// AppError wraps an underlying error with a user-friendly message.
type AppError struct {
	Msg string
	Err error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

func (e *AppError) Unwrap() error {
	return e.Err
}
