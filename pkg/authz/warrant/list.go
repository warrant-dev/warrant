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

	"github.com/pkg/errors"
)

type FilterParams struct {
	ObjectType      []string `json:"objectType,omitempty"`
	ObjectId        []string `json:"objectId,omitempty"`
	Relation        []string `json:"relation,omitempty"`
	SubjectType     []string `json:"subjectType,omitempty"`
	SubjectId       []string `json:"subjectId,omitempty"`
	SubjectRelation []string `json:"subjectRelation,omitempty"`
	Policy          Policy   `json:"policy,omitempty"`
}

func (fp FilterParams) String() string {
	s := ""
	if len(fp.ObjectType) > 0 {
		s = fmt.Sprintf("%s&objectType=%s", s, strings.Join(fp.ObjectType, ","))
	}

	if len(fp.ObjectId) > 0 {
		s = fmt.Sprintf("%s&objectId=%s", s, strings.Join(fp.ObjectId, ","))
	}

	if len(fp.Relation) > 0 {
		s = fmt.Sprintf("%s&relation=%s", s, strings.Join(fp.Relation, ","))
	}

	if len(fp.SubjectType) > 0 {
		s = fmt.Sprintf("%s&subjectType=%s", s, strings.Join(fp.SubjectType, ","))
	}

	if len(fp.SubjectId) > 0 {
		s = fmt.Sprintf("%s&subjectId=%s", s, strings.Join(fp.SubjectId, ","))
	}

	if len(fp.SubjectRelation) > 0 {
		s = fmt.Sprintf("%s&subjectRelation=%s", s, strings.Join(fp.SubjectRelation, ","))
	}

	if fp.Policy != "" {
		s = fmt.Sprintf("%s&policy=%s", s, fp.Policy)
	}

	return strings.TrimPrefix(s, "&")
}

const PrimarySortKey = "id"

type WarrantListParamParser struct{}

func (parser WarrantListParamParser) GetDefaultSortBy() string {
	//nolint:goconst
	return "createdAt"
}

func (parser WarrantListParamParser) GetSupportedSortBys() []string {
	return []string{"createdAt"}
}

func (parser WarrantListParamParser) ParseValue(val string, sortBy string) (interface{}, error) {
	// TODO: add support for more sortBy columns
	switch sortBy {
	case "createdAt":
		value, err := time.Parse(time.RFC3339, val)
		if err != nil || value.Equal(time.Time{}) {
			return nil, fmt.Errorf("must be a valid time in the format %s", time.RFC3339)
		}

		return &value, nil
	default:
		return nil, errors.New(fmt.Sprintf("must match type of selected sortBy attribute %s", sortBy))
	}
}
