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
	"time"

	"github.com/pkg/errors"
)

const PrimarySortKey = "id"

type QueryListParamParser struct{}

func (parser QueryListParamParser) GetDefaultSortBy() string {
	return "id"
}

func (parser QueryListParamParser) GetSupportedSortBys() []string {
	return []string{"id", "createdAt"}
}

func (parser QueryListParamParser) ParseValue(val string, sortBy string) (interface{}, error) {
	switch sortBy {
	//nolint:goconst
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

type ByObjectTypeAndObjectIdAndRelationAsc []QueryResult

func (res ByObjectTypeAndObjectIdAndRelationAsc) Len() int      { return len(res) }
func (res ByObjectTypeAndObjectIdAndRelationAsc) Swap(i, j int) { res[i], res[j] = res[j], res[i] }
func (res ByObjectTypeAndObjectIdAndRelationAsc) Less(i, j int) bool {
	if res[i].ObjectType == res[j].ObjectType {
		if res[i].ObjectId == res[j].ObjectId {
			return res[i].Relation < res[j].Relation
		}
		return res[i].ObjectId < res[j].ObjectId
	}
	return res[i].ObjectType < res[j].ObjectType
}

type ByObjectTypeAndObjectIdAndRelationDesc []QueryResult

func (res ByObjectTypeAndObjectIdAndRelationDesc) Len() int      { return len(res) }
func (res ByObjectTypeAndObjectIdAndRelationDesc) Swap(i, j int) { res[i], res[j] = res[j], res[i] }
func (res ByObjectTypeAndObjectIdAndRelationDesc) Less(i, j int) bool {
	if res[i].ObjectType == res[j].ObjectType {
		if res[i].ObjectId == res[j].ObjectId {
			return res[i].Relation > res[j].Relation
		}
		return res[i].ObjectId > res[j].ObjectId
	}
	return res[i].ObjectType > res[j].ObjectType
}

type ByCreatedAtAsc []QueryResult

func (res ByCreatedAtAsc) Len() int      { return len(res) }
func (res ByCreatedAtAsc) Swap(i, j int) { res[i], res[j] = res[j], res[i] }
func (res ByCreatedAtAsc) Less(i, j int) bool {
	return res[i].Warrant.CreatedAt.Before(res[j].Warrant.CreatedAt)
}

type ByCreatedAtDesc []QueryResult

func (res ByCreatedAtDesc) Len() int      { return len(res) }
func (res ByCreatedAtDesc) Swap(i, j int) { res[i], res[j] = res[j], res[i] }
func (res ByCreatedAtDesc) Less(i, j int) bool {
	return res[i].Warrant.CreatedAt.After(res[j].Warrant.CreatedAt)
}
