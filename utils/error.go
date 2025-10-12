package utils

import "fmt"

// CustomError represents a custom error with HTTP status code
type CustomError struct {
	Message    string
	StatusCode int
	Err        error
}

func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewCustomError creates a new custom error
func NewCustomError(message string, statusCode int) error {
	return &CustomError{
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewCustomErrorWithTrace creates a new custom error with trace
func NewCustomErrorWithTrace(err error, message string, statusCode int) error {
	return &CustomError{
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// PanicIfError panics if error is not nil
func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

// PanicIfAppError panics with custom error if error is not nil
func PanicIfAppError(err error, message string, statusCode int) {
	if err != nil {
		panic(NewCustomErrorWithTrace(err, message, statusCode))
	}
}

// PanicAppError panics with custom error
func PanicAppError(message string, statusCode int) {
	panic(NewCustomError(message, statusCode))
}
