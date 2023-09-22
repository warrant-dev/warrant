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
	baseWarrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
)

const Wildcard = "*"

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
