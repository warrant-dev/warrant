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

package wookie_test

import (
	"context"
	"testing"

	"github.com/warrant-dev/warrant/pkg/service"
	"github.com/warrant-dev/warrant/pkg/wookie"
)

func TestBasicSerialization(t *testing.T) {
	t.Parallel()
	ctx := service.WithLatest(context.Background())
	if !wookie.ContainsLatest(ctx) {
		t.Fatalf("expected ctx to contain 'latest' wookie")
	}

	ctx = context.Background()
	if wookie.ContainsLatest(ctx) {
		t.Fatalf("expected ctx to not contain 'latest' wookie")
	}
}
