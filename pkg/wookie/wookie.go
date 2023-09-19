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
)

const Latest = "latest"

type WookieCtxKey struct{}

// Returns true if ctx contains wookie set to 'latest', false otherwise.
func ContainsLatest(ctx context.Context) bool {
	if val, ok := ctx.Value(WookieCtxKey{}).(string); ok {
		if val == Latest {
			return true
		}
	}
	return false
}
