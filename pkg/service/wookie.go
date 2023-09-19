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

package service

import (
	"context"
	"net/http"

	"github.com/warrant-dev/warrant/pkg/wookie"
)

const WarrantTokenHeaderName = "Warrant-Token"

func WookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerVal := r.Header.Get(WarrantTokenHeaderName)
		if headerVal == wookie.Latest {
			latestCtx := WithLatest(r.Context())
			next.ServeHTTP(w, r.WithContext(latestCtx))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Return a context with wookie set to 'latest'.
func WithLatest(parent context.Context) context.Context {
	return context.WithValue(parent, wookie.WookieCtxKey{}, wookie.Latest)
}

func AddAsResponseHeader(w http.ResponseWriter, token string) {
	if token != "" {
		w.Header().Set(WarrantTokenHeaderName, token)
	}
}
