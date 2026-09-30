// Package mcpv2 — App-layer errors (Ultimate Go §8 — App defines codes, API maps to HTTP).
// "Once you handle it, it's not an error anymore" — handle = log once + inspect + respond.
package mcpv2

import "fmt"

// Code is protocol-agnostic (App layer). API maps to HTTP.
type Code string

const (
	CodeInvalidInput   Code = "INVALID_INPUT"
	CodeNotFound       Code = "NOT_FOUND"
	CodeLimitExceeded  Code = "LIMIT_EXCEEDED"
	CodeUnauthorized   Code = "UNAUTHORIZED"
	CodeForbidden      Code = "FORBIDDEN"
	CodeInternal       Code = "INTERNAL"
)

// AppError is App-layer error with code — never leak internals.
type AppError struct {
	Code Code
	Msg  string
	Err  error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Msg, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Msg)
}
func (e *AppError) Unwrap() error { return e.Err }

// Helpers
func NewInvalidInput(msg string, err error) *AppError { return &AppError{Code: CodeInvalidInput, Msg: msg, Err: err} }
func NewNotFound(msg string) *AppError               { return &AppError{Code: CodeNotFound, Msg: msg} }
func NewLimitExceeded(msg string) *AppError           { return &AppError{Code: CodeLimitExceeded, Msg: msg} }
func NewInternal(msg string, err error) *AppError     { return &AppError{Code: CodeInternal, Msg: msg, Err: err} }

// MapCodeToHTTP maps App Code -> HTTP (API layer responsibility, but helper here for handler).
func MapCodeToHTTP(c Code) int {
	switch c {
	case CodeInvalidInput:
		return 400
	case CodeUnauthorized:
		return 401
	case CodeForbidden:
		return 403
	case CodeNotFound:
		return 404
	case CodeLimitExceeded:
		return 429
	default:
		return 500
	}
}