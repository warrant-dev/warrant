// Copyright 2024 WorkOS, Inc.
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
const OrgIdHeaderName = "x-org-id"
const Latest = "latest"
const OrgIdKey = "orgId"

type warrantTokenCtxKey struct{}

func WarrantTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerVal := r.Header.Get(HeaderName)
		if headerVal != "" {
			warrantTokenCtx := context.WithValue(r.Context(), warrantTokenCtxKey{}, headerVal)
			next.ServeHTTP(w, r.WithContext(warrantTokenCtx))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func OrgIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgId := r.Header.Get(OrgIdHeaderName)
		if orgId == "" {
			orgId = r.URL.Query().Get("orgId")
		}
		if orgId == "" {
			http.Error(w, "no orgId found in header[X-org-id] or query[orgId]", http.StatusUnauthorized)
			return
		}
		orgIdCtx := context.WithValue(r.Context(), OrgIdKey, orgId)
		next.ServeHTTP(w, r.WithContext(orgIdCtx))
	})
}

// Returns true if ctx contains wookie set to 'latest', false otherwise.
func ContainsLatest(ctx context.Context) bool {
	if val, ok := ctx.Value(warrantTokenCtxKey{}).(string); ok {
		if val == Latest {
			return true
		}
	}
	return false
}

// Return a context with Warrant-Token set to 'latest'.
func WithLatest(parent context.Context) context.Context {
	return context.WithValue(parent, warrantTokenCtxKey{}, Latest)
}
