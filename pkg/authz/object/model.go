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
	"time"
)

type Model interface {
	GetID() int64
	GetObjectType() string
	GetObjectId() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
	ToObjectSpec() *ObjectSpec
}

type Object struct {
	ID         int64      `mysql:"id" postgres:"id" sqlite:"id"`
	ObjectType string     `mysql:"objectType" postgres:"object_type" sqlite:"objectType"`
	ObjectId   string     `mysql:"objectId" postgres:"object_id" sqlite:"objectId"`
	CreatedAt  time.Time  `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
	UpdatedAt  time.Time  `mysql:"updatedAt" postgres:"updated_at" sqlite:"updatedAt"`
	DeletedAt  *time.Time `mysql:"deletedAt" postgres:"deleted_at" sqlite:"deletedAt"`
}

func (object Object) GetID() int64 {
	return object.ID
}

func (object Object) GetObjectType() string {
	return object.ObjectType
}

func (object Object) GetObjectId() string {
	return object.ObjectId
}

func (object Object) GetCreatedAt() time.Time {
	return object.CreatedAt
}

func (object Object) GetUpdatedAt() time.Time {
	return object.UpdatedAt
}

func (object Object) GetDeletedAt() *time.Time {
	return object.DeletedAt
}

func (object Object) ToObjectSpec() *ObjectSpec {
	return &ObjectSpec{
		ID:         object.ID,
		ObjectType: object.ObjectType,
		ObjectId:   object.ObjectId,
		CreatedAt:  object.CreatedAt,
	}
}
