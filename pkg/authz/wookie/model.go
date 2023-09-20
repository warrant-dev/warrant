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

package authz

import "time"

type Model interface {
	GetID() int64
	GetVersion() int64
	GetCreatedAt() time.Time
	ToToken() *Token
}

type Wookie struct {
	ID        int64     `mysql:"id" postgres:"id" sqlite:"id"`
	Version   int64     `mysql:"ver" postgres:"ver" sqlite:"ver"`
	CreatedAt time.Time `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
}

func (w Wookie) GetID() int64 {
	return w.ID
}

func (w Wookie) GetVersion() int64 {
	return w.Version
}

func (w Wookie) GetCreatedAt() time.Time {
	return w.CreatedAt
}

func (w Wookie) ToToken() *Token {
	return &Token{
		ID:        w.ID,
		Version:   w.Version,
		Timestamp: w.CreatedAt,
	}
}
