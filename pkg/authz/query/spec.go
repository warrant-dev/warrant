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
	"encoding/base64"
	"encoding/json"

	"github.com/pkg/errors"
	baseWarrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
)

type Query struct {
	Expand         bool
	SelectSubjects *SelectSubjects
	SelectObjects  *SelectObjects
	Context        *baseWarrant.PolicyContext
}

type SelectSubjects struct {
	ForObject    *Resource
	Relations    []string
	SubjectTypes []string
}

type SelectObjects struct {
	ObjectTypes  []string
	Relations    []string
	WhereSubject *Resource
}

type Resource struct {
	Type string
	Id   string
}

type QueryHaving struct {
	ObjectType  string `json:"objectType,omitempty"`
	ObjectId    string `json:"objectId,omitempty"`
	Relation    string `json:"relation,omitempty"`
	SubjectType string `json:"subjectType,omitempty"`
	SubjectId   string `json:"subjectId,omitempty"`
}

type QueryResult struct {
	ObjectType string                  `json:"objectType"`
	ObjectId   string                  `json:"objectId"`
	Warrant    baseWarrant.WarrantSpec `json:"warrant"`
	IsImplicit bool                    `json:"isImplicit"`
	Meta       map[string]interface{}  `json:"meta,omitempty"`
}

type Result struct {
	Results []QueryResult `json:"results"`
	LastId  string        `json:"lastId,omitempty"`
}

type ByObjectTypeAsc []QueryResult

func (res ByObjectTypeAsc) Len() int      { return len(res) }
func (res ByObjectTypeAsc) Swap(i, j int) { res[i], res[j] = res[j], res[i] }
func (res ByObjectTypeAsc) Less(i, j int) bool {
	if res[i].ObjectType == res[j].ObjectType {
		return res[i].ObjectId < res[j].ObjectId
	}
	return res[i].ObjectType < res[j].ObjectType
}

type ByObjectTypeDesc []QueryResult

func (res ByObjectTypeDesc) Len() int      { return len(res) }
func (res ByObjectTypeDesc) Swap(i, j int) { res[i], res[j] = res[j], res[i] }
func (res ByObjectTypeDesc) Less(i, j int) bool {
	if res[i].ObjectType == res[j].ObjectType {
		return res[i].ObjectId > res[j].ObjectId
	}
	return res[i].ObjectType > res[j].ObjectType
}

type ByObjectIdAsc []QueryResult

func (res ByObjectIdAsc) Len() int           { return len(res) }
func (res ByObjectIdAsc) Swap(i, j int)      { res[i], res[j] = res[j], res[i] }
func (res ByObjectIdAsc) Less(i, j int) bool { return res[i].ObjectId < res[j].ObjectId }

type ByObjectIdDesc []QueryResult

func (res ByObjectIdDesc) Len() int           { return len(res) }
func (res ByObjectIdDesc) Swap(i, j int)      { res[i], res[j] = res[j], res[i] }
func (res ByObjectIdDesc) Less(i, j int) bool { return res[i].ObjectId > res[j].ObjectId }

type LastIdSpec struct {
	ObjectType string `json:"objectType"`
	ObjectId   string `json:"objectId"`
}

func LastIdSpecToString(lastIdSpec LastIdSpec) (string, error) {
	jsonStr, err := json.Marshal(lastIdSpec)
	if err != nil {
		return "", errors.Wrapf(err, "error marshaling lastId %v", lastIdSpec)
	}

	return base64.StdEncoding.EncodeToString(jsonStr), nil
}

func StringToLastIdSpec(base64Str string) (*LastIdSpec, error) {
	var lastIdSpec LastIdSpec
	jsonStr, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, errors.Wrapf(err, "error base64 decoding lastId string %s", base64Str)
	}

	err = json.Unmarshal(jsonStr, &lastIdSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "error unmarshaling lastIdSpec %v", lastIdSpec)
	}

	return &lastIdSpec, nil
}
