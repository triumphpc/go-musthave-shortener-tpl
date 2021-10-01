// Package errors consist package level errors
package errors

import "errors"

// ErrBadResponse bad http request
var ErrBadResponse = errors.New("bad request")

// ErrUnknownURL if can't find url in storage
var ErrUnknownURL = errors.New("unknown url")

// ErrInternalError some internal errors
var ErrInternalError = errors.New("internal error")

// ErrNoContent no records in database
var ErrNoContent = errors.New("no content")
