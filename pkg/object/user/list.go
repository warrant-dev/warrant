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

package object

import (
	"fmt"
	"net/mail"
	"time"

	"github.com/pkg/errors"
)

type UserListParamParser struct{}

func (parser UserListParamParser) GetDefaultSortBy() string {
	return "userId"
}

func (parser UserListParamParser) GetSupportedSortBys() []string {
	return []string{"createdAt", "userId", "email"}
}

func (parser UserListParamParser) ParseValue(val string, sortBy string) (interface{}, error) {
	switch sortBy {
	case "createdAt":
		value, err := time.Parse(time.RFC3339, val)
		if err != nil || value.Equal(time.Time{}) {
			return nil, fmt.Errorf("must be a valid time in the format %s", time.RFC3339)
		}

		return &value, nil
	case "email":
		if val == "" {
			return "", nil
		}

		afterValue, err := mail.ParseAddress(val)
		if err != nil {
			return nil, errors.New("must be a valid email")
		}

		return afterValue.Address, nil
	case "userId":
		if val == "" {
			return nil, errors.New("must not be empty")
		}

		return val, nil
	default:
		return nil, errors.New(fmt.Sprintf("must match type of selected sortBy attribute %s", sortBy))
	}
}
