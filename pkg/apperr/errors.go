package apperr

import "fmt"

// Code is a typed integer for machine-readable error codes.
type Code int

const (
	CodeInvalidInput    Code = 400
	CodeUnauthorized    Code = 401
	CodeForbidden       Code = 403
	CodeNotFound        Code = 404
	CodeConflict        Code = 409
	CodeRateLimited     Code = 429
	CodeInternal        Code = 500

	// Domain-specific codes
	CodeInsufficientBal Code = 4001 // not enough balance
	CodeDuplicateOrder  Code = 4002 // same orderID submitted twice
	CodePositionLimit   Code = 4003 // position size exceeded
	CodeSelfMatch       Code = 4004 // buyer == seller (not allowed)
)

// AppError is the single error type used across all services.
type AppError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"` // internal cause — never sent to client
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New creates an AppError without a cause.
// Use when the error is a business rule violation.
// Example: apperr.New(apperr.CodeInvalidInput, "price must be > 0")
func New(code Code, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

// Wrap creates an AppError with an internal cause.
// Use when wrapping a DB or external error.
// Example: apperr.Wrap(apperr.CodeInternal, "failed to save order", err)
func Wrap(code Code, msg string, err error) *AppError {
	return &AppError{Code: code, Message: msg, Err: err}
}