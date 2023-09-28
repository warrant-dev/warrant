// Copyright 2023 Forerunner Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authz

import (
	"context"
	"net/http"

	"github.com/warrant-dev/warrant/pkg/service"
	"github.com/warrant-dev/warrant/pkg/wookie"
)

func GenerateWookieMiddleware(wookieSvc *WookieService) service.Middleware {
	return func(next http.Handler) http.Handler {
		return wookieMiddleware(next, wookieSvc)
	}
}

func wookieMiddleware(next http.Handler, wookieSvc *WookieService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := wookie.GetClientPassedWookieFromRequestContext(r.Context())
		if err == nil {
			next.ServeHTTP(w, r)
			return
		}

		headerVal := r.Header.Get(wookie.HeaderName)

		switch headerVal {
		case wookie.Latest, "":
			token, err := wookieSvc.GetLatestWookie(r.Context())
			if err != nil {
				service.SendErrorResponse(w, service.NewInvalidRequestError("invalid warrant token"))
				return
			}

			ctxWithToken := context.WithValue(r.Context(), wookie.ClientPassedWookieCtxKey{}, *token)
			next.ServeHTTP(w, r.WithContext(ctxWithToken))
		default:
			token, err := wookie.FromString(headerVal)
			if err != nil {
				service.SendErrorResponse(w, service.NewInvalidRequestError("invalid warrant token"))
				return
			}

			ctxWithToken := context.WithValue(r.Context(), wookie.ClientPassedWookieCtxKey{}, token)
			next.ServeHTTP(w, r.WithContext(ctxWithToken))
		}
	})
}
