package wookie

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/hlog"
)

const WarrantTokenHeaderName = "Warrant-Token"

func ClientTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerVal := r.Header.Get(WarrantTokenHeaderName)
		if headerVal != "" {
			clientWookie, err := FromString(headerVal)
			if err != nil {
				hlog.FromRequest(r).Warn().Msgf("invalid client-supplied wookie header: %s", headerVal)
				next.ServeHTTP(w, r)
				return
			}
			newContext := context.WithValue(r.Context(), ClientTokenKey{}, clientWookie)
			next.ServeHTTP(w, r.WithContext(newContext))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AddAsResponseHeader(w http.ResponseWriter, token *Token) {
	if token != nil {
		w.Header().Set(WarrantTokenHeaderName, token.AsString())
	}
}
