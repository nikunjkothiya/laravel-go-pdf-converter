package errors

import (
	"encoding/json"
	"fmt"
)

// ErrorCode represents specific error types for structured handling
type ErrorCode string

const (
	ErrFileNotFound      ErrorCode = "FILE_NOT_FOUND"
	ErrInvalidFormat     ErrorCode = "INVALID_FORMAT"
	ErrCorruptFile       ErrorCode = "CORRUPT_FILE"
	ErrConversionFailed  ErrorCode = "CONVERSION_FAILED"
	ErrMemoryLimit       ErrorCode = "MEMORY_LIMIT"
	ErrTimeout           ErrorCode = "TIMEOUT"
	ErrUnsupportedFormat ErrorCode = "UNSUPPORTED_FORMAT"
	ErrWriteFailed       ErrorCode = "WRITE_FAILED"
	ErrParseFailed       ErrorCode = "PARSE_FAILED"
)

// ConversionError is a structured error with JSON output for Laravel parsing
type ConversionError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	File    string    `json:"file,omitempty"`
	Details string    `json:"details,omitempty"`
}

func (e *ConversionError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ToJSON returns the error as a JSON string for Laravel to parse
func (e *ConversionError) ToJSON() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// New creates a new ConversionError
func New(code ErrorCode, message string) *ConversionError {
	return &ConversionError{
		Code:    code,
		Message: message,
	}
}

// NewWithFile creates an error with file context
func NewWithFile(code ErrorCode, message, file string) *ConversionError {
	return &ConversionError{
		Code:    code,
		Message: message,
		File:    file,
	}
}

// NewWithDetails creates an error with additional details
func NewWithDetails(code ErrorCode, message, file, details string) *ConversionError {
	return &ConversionError{
		Code:    code,
		Message: message,
		File:    file,
		Details: details,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string) *ConversionError {
	return &ConversionError{
		Code:    code,
		Message: message,
		Details: err.Error(),
	}
}
