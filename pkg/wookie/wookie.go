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

type WookieCtxKey struct{}
type WookieKey struct{}

// Returns true if ctx contains wookie set to 'latest', false otherwise.
func ContainsLatest(ctx context.Context) bool {
	if val, ok := ctx.Value(WookieCtxKey{}).(string); ok {
		if val == Latest {
			return true
		}
	}
	return false
}

func GetWookieFromRequestContext(ctx context.Context) (*Token, error) {
	wookie, ok := ctx.Value(WookieKey{}).(Token)
	if !ok {
		return nil, errors.New("error getting wookie from request context")
	}
	return &wookie, nil
}

// Return a context with wookie set to 'latest'.
func WithLatest(parent context.Context) context.Context {
	return context.WithValue(parent, WookieCtxKey{}, Latest)
}

func AddAsResponseHeader(w http.ResponseWriter, token *Token) {
	if token != nil {
		w.Header().Set(HeaderName, token.String())
	}
}
