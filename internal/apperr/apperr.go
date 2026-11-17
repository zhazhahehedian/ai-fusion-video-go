package apperr

import "fmt"

// Error is a business error mapped to API {code,msg} and an HTTP status.
// Convention mirrors Java BusinessException: code often equals HTTP status (401/400/...).
type Error struct {
	Code       int
	Msg        string
	HTTPStatus int
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("apperr %d: %s", e.Code, e.Msg)
}

func New(httpStatus, code int, msg string) *Error {
	return &Error{Code: code, Msg: msg, HTTPStatus: httpStatus}
}

func BadRequest(msg string) *Error {
	return New(400, 400, msg)
}

func Unauthorized(msg string) *Error {
	return New(401, 401, msg)
}

func Forbidden(msg string) *Error {
	return New(403, 403, msg)
}

func NotFound(msg string) *Error {
	return New(404, 404, msg)
}

func Conflict(msg string) *Error {
	return New(409, 409, msg)
}
