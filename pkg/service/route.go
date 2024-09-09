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

package service

import "net/http"

type Route interface {
	GetPattern() string
	GetMethod() string
	GetHandler() http.Handler
	GetOverrideAuthMiddlewareFunc() AuthMiddlewareFunc
}

type WarrantRoute struct {
	Pattern                    string
	Method                     string
	Handler                    http.Handler
	OverrideAuthMiddlewareFunc AuthMiddlewareFunc
}

func (route WarrantRoute) GetPattern() string {
	return route.Pattern
}

func (route WarrantRoute) GetMethod() string {
	return route.Method
}

func (route WarrantRoute) GetHandler() http.Handler {
	return route.Handler
}

func (route WarrantRoute) GetOverrideAuthMiddlewareFunc() AuthMiddlewareFunc {
	return route.OverrideAuthMiddlewareFunc
}
