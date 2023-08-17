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

	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
)

type UserSpec struct {
	UserId    string    `json:"userId" validate:"omitempty,valid_object_id"`
	Email     *string   `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func (spec UserSpec) ToUser(objectId int64) *User {
	return &User{
		ObjectId: objectId,
		UserId:   spec.UserId,
		Email:    spec.Email,
	}
}

func (spec UserSpec) ToCreateObjectSpec() *object.CreateObjectSpec {
	return &object.CreateObjectSpec{
		ObjectType: objecttype.ObjectTypeUser,
		ObjectId:   spec.UserId,
	}
}

type UpdateUserSpec struct {
	Email *string `json:"email"`
}
