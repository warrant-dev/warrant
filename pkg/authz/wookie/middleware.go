package wookie

import (
	"context"
	"net/http"
)

func WookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientWookie, err := Deserialize(r.Header.Get("Warrant-Token"))
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		newContext := context.WithValue(r.Context(), TokenKey{}, clientWookie)
		next.ServeHTTP(w, r.WithContext(newContext))
		// TODO: also set the value as a response header
	})
}
