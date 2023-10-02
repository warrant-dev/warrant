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

	"github.com/pkg/errors"
)

const HeaderName = "Warrant-Token"
const Latest = "latest"

type warrantTokenCtxKey struct{}
type wookieCtxKey struct{}

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

// Returns true if ctx contains wookie set to 'latest', false otherwise.
func ContainsLatest(ctx context.Context) bool {
	if val, ok := ctx.Value(warrantTokenCtxKey{}).(string); ok {
		if val == Latest {
			return true
		}
	}
	return false
}

func GetWookieFromContext(ctx context.Context) (*Token, error) {
	wookieCtxVal := ctx.Value(wookieCtxKey{})
	if wookieCtxVal == nil {
		return nil, errors.New("wookie not found in context")
	}

	wookieToken, ok := wookieCtxVal.(Token)
	if !ok {
		return nil, errors.New("error fetching wookie from context")
	}

	return &wookieToken, nil
}

// Return a context with Warrant-Token set to 'latest'.
func WithLatest(parent context.Context) context.Context {
	return context.WithValue(parent, warrantTokenCtxKey{}, Latest)
}

// Return context with wookie set to specified Token.
func WithWookie(parent context.Context, wookie *Token) context.Context {
	return context.WithValue(parent, wookieCtxKey{}, *wookie)
}

func AddAsResponseHeader(w http.ResponseWriter, token *Token) {
	if token != nil {
		w.Header().Set(HeaderName, token.String())
	}
}
