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

package authz

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type FilterParams struct {
	ObjectType      string `json:"objectType,omitempty"`
	ObjectId        string `json:"objectId,omitempty"`
	Relation        string `json:"relation,omitempty"`
	SubjectType     string `json:"subjectType,omitempty"`
	SubjectId       string `json:"subjectId,omitempty"`
	SubjectRelation string `json:"subjectRelation,omitempty"`
}

func (fp FilterParams) String() string {
	s := ""
	if len(fp.ObjectType) > 0 {
		s = fmt.Sprintf("%s&objectType=%s", s, fp.ObjectType)
	}

	if len(fp.ObjectId) > 0 {
		s = fmt.Sprintf("%s&objectId=%s", s, fp.ObjectId)
	}

	if len(fp.Relation) > 0 {
		s = fmt.Sprintf("%s&relation=%s", s, fp.Relation)
	}

	if len(fp.SubjectType) > 0 {
		s = fmt.Sprintf("%s&subjectType=%s", s, fp.SubjectType)
	}

	if len(fp.SubjectId) > 0 {
		s = fmt.Sprintf("%s&subjectId=%s", s, fp.SubjectId)
	}

	if len(fp.SubjectRelation) > 0 {
		s = fmt.Sprintf("%s&subjectRelation=%s", s, fp.SubjectRelation)
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
