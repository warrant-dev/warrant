package middleware

import "net/http"

// Middleware defines the type of all middleware
type Middleware func(http.Handler) http.Handler

// ChainMiddleware a top-level middleware which applies the given middlewares in order from inner to outer (order of execution)
func ChainMiddleware(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
