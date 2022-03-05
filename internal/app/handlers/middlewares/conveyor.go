// Package middlewares consist methods for parse http request
package middlewares

import "net/http"

// Middleware type for execute middlewares on handlers request
type Middleware func(http.Handler) http.Handler

// Conveyor service handlers
func Conveyor(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}
