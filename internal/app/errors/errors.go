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

// ErrURLNotFound error by package level
var ErrURLNotFound = errors.New("url not found")

// ErrAlreadyHasShort if exist
var ErrAlreadyHasShort = errors.New("already has short")

// ErrURLIsGone in storage
var ErrURLIsGone = errors.New("url is gone")
