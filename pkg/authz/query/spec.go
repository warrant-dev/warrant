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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/pkg/service"
)

type Query struct {
	Expand         bool
	SelectSubjects *SelectSubjects
	SelectObjects  *SelectObjects
	Context        warrant.PolicyContext
}

func (q *Query) WithContext(contextString string) error {
	var context warrant.PolicyContext
	err := json.Unmarshal([]byte(contextString), &context)
	if err != nil {
		return errors.Wrap(err, "query: error parsing query context")
	}

	q.Context = context
	return nil
}

func (q *Query) String() string {
	var str string
	if q.Expand {
		str = "select"
	} else {
		str = "select explicit"
	}

	if q.SelectObjects != nil {
		return fmt.Sprintf("%s %s %s", str, q.SelectObjects.String(), q.Context.String())
	} else if q.SelectSubjects != nil {
		return fmt.Sprintf("%s %s %s", str, q.SelectSubjects.String(), q.Context.String())
	}

	return ""
}

type SelectSubjects struct {
	ForObject    *Resource
	Relations    []string
	SubjectTypes []string
}

func (s SelectSubjects) String() string {
	str := fmt.Sprintf("%s of type %s", strings.Join(s.Relations, ", "), strings.Join(s.SubjectTypes, ", "))
	if s.ForObject != nil {
		str = fmt.Sprintf("%s for %s", str, s.ForObject.String())
	}

	return str
}

type SelectObjects struct {
	ObjectTypes  []string
	Relations    []string
	WhereSubject *Resource
}

func (s SelectObjects) String() string {
	str := strings.Join(s.ObjectTypes, ", ")
	if s.WhereSubject != nil {
		str = fmt.Sprintf("%s where %s", str, s.WhereSubject.String())
	}

	return fmt.Sprintf("%s is %s", str, strings.Join(s.Relations, ", "))
}

type Resource struct {
	Type string
	Id   string
}

func (res Resource) String() string {
	return fmt.Sprintf("%s:%s", res.Type, res.Id)
}

type QueryHaving struct {
	ObjectType  string `json:"objectType,omitempty"`
	ObjectId    string `json:"objectId,omitempty"`
	Relation    string `json:"relation,omitempty"`
	SubjectType string `json:"subjectType,omitempty"`
	SubjectId   string `json:"subjectId,omitempty"`
}

type QueryResult struct {
	ObjectType string                 `json:"objectType"`
	ObjectId   string                 `json:"objectId"`
	Relation   string                 `json:"relation"`
	Warrant    warrant.WarrantSpec    `json:"warrant"`
	IsImplicit bool                   `json:"isImplicit"`
	Meta       map[string]interface{} `json:"meta,omitempty"`
}

type QueryResponseV1 struct {
	Results []QueryResult `json:"results"`
	LastId  string        `json:"lastId,omitempty"`
}

type QueryResponseV2 struct {
	Results    []QueryResult   `json:"results"`
	PrevCursor *service.Cursor `json:"prevCursor,omitempty"`
	NextCursor *service.Cursor `json:"nextCursor,omitempty"`
}
