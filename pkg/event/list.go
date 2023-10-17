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

package event

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

const (
	DefaultEpochMicroseconds  = 1136160000000
	QueryParamType            = "type"
	QueryParamSource          = "source"
	QueryParamResourceType    = "resourceType"
	QueryParamResourceId      = "resourceId"
	QueryParamLastId          = "lastId"
	QueryParamSince           = "since"
	QueryParamUntil           = "until"
	QueryParamObjectType      = "objectType"
	QueryParamObjectId        = "objectId"
	QueryParamRelation        = "relation"
	QueryParamSubjectType     = "subjectType"
	QueryParamSubjectId       = "subjectId"
	QueryParamSubjectRelation = "subjectRelation"
)

type ResourceEventFilterParams struct {
	Type         string
	Source       string
	ResourceType string
	ResourceId   string
	Since        time.Time
	Until        time.Time
}

type ResourceEventListParamParser struct{}

func (parser ResourceEventListParamParser) GetDefaultSortBy() string {
	//nolint:goconst
	return "createdAt"
}

func (parser ResourceEventListParamParser) GetSupportedSortBys() []string {
	return []string{"createdAt"}
}

func (parser ResourceEventListParamParser) ParseValue(val string, sortBy string) (interface{}, error) {
	switch sortBy {
	case "createdAt":
		value, err := time.Parse(time.RFC3339, val)
		if err != nil || value.Equal(time.Time{}) {
			return nil, errors.New(fmt.Sprintf("must be a valid time in the format %s", time.RFC3339))
		}

		return value, nil
	default:
		return nil, errors.New(fmt.Sprintf("must match type of selected sortBy attribute %s", sortBy))
	}
}

type AccessEventFilterParams struct {
	Type            string
	Source          string
	ObjectType      string
	ObjectId        string
	Relation        string
	SubjectType     string
	SubjectId       string
	SubjectRelation string
	Since           time.Time
	Until           time.Time
}

type AccessEventListParamParser struct{}

func (parser AccessEventListParamParser) GetDefaultSortBy() string {
	return "createdAt"
}

func (parser AccessEventListParamParser) GetSupportedSortBys() []string {
	return []string{"createdAt"}
}

func (parser AccessEventListParamParser) ParseValue(val string, sortBy string) (interface{}, error) {
	switch sortBy {
	case "createdAt":
		value, err := time.Parse(time.RFC3339, val)
		if err != nil || value.Equal(time.Time{}) {
			return nil, errors.New(fmt.Sprintf("must be a valid time in the format %s", time.RFC3339))
		}

		return &value, nil
	default:
		return nil, errors.New(fmt.Sprintf("must match type of selected sortBy attribute %s", sortBy))
	}
}
