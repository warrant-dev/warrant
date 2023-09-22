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
	"strings"
	"time"
)

type FilterParams struct {
	ObjectType      []string
	ObjectId        []string
	Relation        []string
	SubjectType     []string
	SubjectId       []string
	SubjectRelation []string
	Policy          Policy
}

func (fp FilterParams) String() string {
	return fmt.Sprintf(
		"objectType: '%s' objectId: '%s' relation: '%s' subjectType: '%s' subjectId: '%s' subjectRelation: '%s' policy: '%s'",
		strings.Join(fp.ObjectType, ", "),
		strings.Join(fp.ObjectId, ", "),
		strings.Join(fp.Relation, ", "),
		strings.Join(fp.SubjectType, ", "),
		strings.Join(fp.SubjectId, ", "),
		strings.Join(fp.SubjectRelation, ", "),
		fp.Policy,
	)
}

type WarrantListParamParser struct{}

func (parser WarrantListParamParser) GetDefaultSortBy() string {
	return "createdAt"
}

func (parser WarrantListParamParser) GetSupportedSortBys() []string {
	return []string{"createdAt"}
}

func (parser WarrantListParamParser) ParseValue(val string, sortBy string) (interface{}, error) {
	switch sortBy {
	case "createdAt":
		afterValue, err := time.Parse(time.RFC3339, val)
		if err != nil || afterValue.Equal(time.Time{}) {
			return nil, fmt.Errorf("must be a valid time in the format %s", time.RFC3339)
		}

		return &afterValue, nil
	default:
		return nil, fmt.Errorf("must match type of selected sortBy attribute %s", sortBy)
	}
}
