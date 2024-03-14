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

	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
)

type ResultSetNode struct {
	ObjectType string
	ObjectId   string
	Relation   string
	Warrant    warrant.WarrantSpec
	IsImplicit bool
	next       *ResultSetNode
}

func (node ResultSetNode) Next() *ResultSetNode {
	return node.next
}

type ResultSet struct {
	m    map[string]*ResultSetNode
	head *ResultSetNode
	tail *ResultSetNode
}

func (rs *ResultSet) List() *ResultSetNode {
	if rs == nil {
		return nil
	}

	return rs.head
}

func (rs *ResultSet) Add(objectType string, objectId string, relation string, warrant warrant.WarrantSpec, isImplicit bool) {
	existingRes, exists := rs.m[key(objectType, objectId, relation)]
	if !exists {
		newNode := ResultSetNode{
			ObjectType: objectType,
			ObjectId:   objectId,
			Relation:   relation,
			Warrant:    warrant,
			IsImplicit: isImplicit,
			next:       nil,
		}

		// Add warrant to list
		if rs.head == nil {
			rs.head = &newNode
		}

		if rs.tail != nil {
			rs.tail.next = &newNode
		}

		rs.tail = &newNode

		// Add result node to map for O(1) lookups
		rs.m[key(objectType, objectId, relation)] = &newNode
	} else if existingRes.IsImplicit && !isImplicit { // favor explicit results
		existingRes.IsImplicit = isImplicit
		existingRes.Warrant = warrant
	}
}

func (rs *ResultSet) Len() int {
	return len(rs.m)
}

func (rs *ResultSet) Get(objectType string, objectId string, relation string) *ResultSetNode {
	return rs.m[key(objectType, objectId, relation)]
}

func (rs *ResultSet) Has(objectType string, objectId string, relation string) bool {
	_, exists := rs.m[key(objectType, objectId, relation)]
	return exists
}

func (rs *ResultSet) Union(other *ResultSet) *ResultSet {
	resultSet := NewResultSet()
	for iter := rs.List(); iter != nil; iter = iter.Next() {
		resultSet.Add(iter.ObjectType, iter.ObjectId, iter.Relation, iter.Warrant, iter.IsImplicit)
	}

	for iter := other.List(); iter != nil; iter = iter.Next() {
		isImplicit := iter.IsImplicit
		if resultSet.Has(iter.ObjectType, iter.ObjectId, iter.Relation) {
			isImplicit = isImplicit && resultSet.Get(iter.ObjectType, iter.ObjectId, iter.Relation).IsImplicit
		}
		resultSet.Add(iter.ObjectType, iter.ObjectId, iter.Relation, iter.Warrant, isImplicit)
	}

	return resultSet
}

func (rs *ResultSet) Intersect(other *ResultSet) *ResultSet {
	resultSet := NewResultSet()
	for iter := rs.List(); iter != nil; iter = iter.Next() {
		isImplicit := iter.IsImplicit
		if other.Has(iter.ObjectType, iter.ObjectId, iter.Relation) {
			isImplicit = isImplicit || other.Get(iter.ObjectType, iter.ObjectId, iter.Relation).IsImplicit
			resultSet.Add(iter.ObjectType, iter.ObjectId, iter.Relation, iter.Warrant, isImplicit)
		}
	}

	return resultSet
}

func (rs *ResultSet) String() string {
	var strs []string
	for iter := rs.List(); iter != nil; iter = iter.Next() {
		strs = append(strs, fmt.Sprintf("%s => %s", key(iter.ObjectType, iter.ObjectId, iter.Relation), iter.Warrant.String()))
	}

	return strings.Join(strs, ", ")
}

func NewResultSet() *ResultSet {
	return &ResultSet{
		m:    make(map[string]*ResultSetNode),
		head: nil,
		tail: nil,
	}
}

func key(objectType string, objectId string, relation string) string {
	return fmt.Sprintf("%s:%s#%s", objectType, objectId, relation)
}
