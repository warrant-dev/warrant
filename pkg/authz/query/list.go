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

import (
	"fmt"
)

type QueryListParamParser struct{}

func (parser QueryListParamParser) GetDefaultSortBy() string {
	return "objectType"
}

func (parser QueryListParamParser) GetSupportedSortBys() []string {
	return []string{"objectType", "objectId"}
}

func (parser QueryListParamParser) ParseValue(val string, sortBy string) (interface{}, error) {
	return nil, fmt.Errorf("must match type of selected sortBy attribute %s", sortBy)
}
