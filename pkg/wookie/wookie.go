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
)

const HeaderName = "Warrant-Token"
const Latest = "latest"

type wookieCtxKey struct{}

func ClientHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerVal := r.Header.Get(HeaderName)
		if headerVal == Latest {
			latestCtx := context.WithValue(r.Context(), wookieCtxKey{}, Latest)
			next.ServeHTTP(w, r.WithContext(latestCtx))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Return a context with wookie set to 'latest'
func WithLatest(parent context.Context) context.Context {
	return context.WithValue(parent, wookieCtxKey{}, Latest)
}

// Returns true if ctx contains wookie set to 'latest', false otherwise
func ContainsLatest(ctx context.Context) bool {
	if val, ok := ctx.Value(wookieCtxKey{}).(string); ok {
		if val == Latest {
			return true
		}
	}
	return false
}
