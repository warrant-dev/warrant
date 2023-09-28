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

type WarrantTokenCtxKey struct{}
type ClientPassedWookieCtxKey struct{}
type ServerCreatedWookieCtxKey struct{}

func WarrantTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerVal := r.Header.Get(HeaderName)
		if headerVal != "" {
			wookieCtx := context.WithValue(r.Context(), WarrantTokenCtxKey{}, headerVal)
			next.ServeHTTP(w, r.WithContext(wookieCtx))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Returns true if ctx contains wookie set to 'latest', false otherwise.
func ContainsLatest(ctx context.Context) bool {
	if val, ok := ctx.Value(WarrantTokenCtxKey{}).(string); ok {
		if val == Latest {
			return true
		}
	}
	return false
}

func GetServerCreatedWookieFromRequestContext(ctx context.Context) (*Token, error) {
	wookieCtxVal := ctx.Value(ServerCreatedWookieCtxKey{})
	if wookieCtxVal == nil {
		//nolint:nilnil
		return nil, nil
	}

	wookieString, ok := wookieCtxVal.(*Token)
	if !ok {
		return nil, errors.New("error fetching server created wookie from request context")
	}

	return wookieString, nil
}

func GetClientPassedWookieFromRequestContext(ctx context.Context) (string, error) {
	wookieCtxVal := ctx.Value(ClientPassedWookieCtxKey{})
	if wookieCtxVal == nil {
		return "", nil
	}

	wookieString, ok := wookieCtxVal.(string)
	if !ok {
		return "", errors.New("error fetching client passed wookie from request context")
	}

	return wookieString, nil
}

// Return a context with wookie set to 'latest'.
func WithLatest(parent context.Context) context.Context {
	return context.WithValue(parent, WarrantTokenCtxKey{}, Latest)
}

func AddAsResponseHeader(w http.ResponseWriter, token *Token) {
	if token != nil {
		w.Header().Set(HeaderName, token.String())
	}
}
