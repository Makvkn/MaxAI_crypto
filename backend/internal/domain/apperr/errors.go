// Package apperr defines the stable application error vocabulary. Every layer
// maps its failures into this type: adapters translate provider errors (§28),
// repositories translate SQL errors, and the HTTP layer renders the result as
// the single documented error envelope (§105, §156).
package apperr

import (
	"errors"
	"fmt"
	"maps"
)

// Error is an application error carrying a stable code, a user-safe message,
// optional structured details and an optional wrapped internal cause. The
// cause is for logs and never reaches the client.
type Error struct {
	Code    Code
	Message string
	Details map[string]any

	cause error
}

// New builds an error with the code's default message.
func New(code Code) *Error {
	return &Error{Code: code, Message: metaFor(code).message}
}

// Newf builds an error with a custom user-safe message.
func Newf(code Code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Wrap builds an error that carries an internal cause for observability.
func Wrap(code Code, cause error) *Error {
	return &Error{Code: code, Message: metaFor(code).message, cause: cause}
}

// WithMessage replaces the user-facing message.
func (e *Error) WithMessage(message string) *Error {
	clone := e.clone()
	clone.Message = message
	return clone
}

// WithDetail attaches one structured detail, such as a rate-limit budget or a
// field-level validation failure.
func (e *Error) WithDetail(key string, value any) *Error {
	clone := e.clone()
	if clone.Details == nil {
		clone.Details = map[string]any{}
	}
	clone.Details[key] = value
	return clone
}

// WithDetails attaches several structured details at once.
func (e *Error) WithDetails(details map[string]any) *Error {
	clone := e.clone()
	if clone.Details == nil {
		clone.Details = make(map[string]any, len(details))
	}
	maps.Copy(clone.Details, details)
	return clone
}

// WithCause attaches an internal cause.
func (e *Error) WithCause(cause error) *Error {
	clone := e.clone()
	clone.cause = cause
	return clone
}

func (e *Error) clone() *Error {
	clone := &Error{Code: e.Code, Message: e.Message, cause: e.cause}
	if e.Details != nil {
		clone.Details = maps.Clone(e.Details)
	}
	return clone
}

// Error implements error. The internal cause is included so logs keep the full
// chain; transport code renders Message instead.
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap exposes the internal cause to errors.Is and errors.As.
func (e *Error) Unwrap() error { return e.cause }

// Category reports the transport category of this error.
func (e *Error) Category() Category { return CategoryOf(e.Code) }

// HTTPStatus reports the HTTP status this error maps to.
func (e *Error) HTTPStatus() int { return StatusOf(e.Code) }

// From converts any error into an *Error. Unrecognised errors become internal
// errors so that implementation details never leak to the client (§156).
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return Wrap(CodeInternal, err)
}

// Is reports whether err is an application error with the given code.
func Is(err error, code Code) bool {
	var appErr *Error
	return errors.As(err, &appErr) && appErr.Code == code
}

// ErrNotImplemented marks a skeleton implementation that has not been built
// yet. It exists so unfinished slices fail loudly instead of returning
// plausible-looking empty data.
var ErrNotImplemented = New(CodeNotImplemented)
