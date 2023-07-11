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
		w.Header().Set(WarrantTokenHeaderName, token.String())
	}
}
