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

package object

import (
	"errors"
	"time"

	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	object "github.com/warrant-dev/warrant/pkg/object"
)

type UserSpec struct {
	UserId    string    `json:"userId"    validate:"omitempty,valid_object_id"`
	Email     *string   `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewUserSpecFromObjectSpec(objectSpec *object.ObjectSpec) (*UserSpec, error) {
	var email *string

	if objectSpec.Meta != nil {
		if _, exists := objectSpec.Meta["email"]; exists {
			emailStr, ok := objectSpec.Meta["email"].(string)
			if !ok {
				return nil, errors.New("user email has invalid type in object meta")
			}
			email = &emailStr
		}
	}

	return &UserSpec{
		UserId:    objectSpec.ObjectId,
		Email:     email,
		CreatedAt: objectSpec.CreatedAt,
	}, nil
}

func (spec UserSpec) ToCreateObjectSpec() (*object.CreateObjectSpec, error) {
	createObjectSpec := object.CreateObjectSpec{
		ObjectType: objecttype.ObjectTypeUser,
		ObjectId:   spec.UserId,
	}

	meta := make(map[string]interface{})
	if spec.Email != nil {
		meta["email"] = spec.Email
	}

	if len(meta) > 0 {
		createObjectSpec.Meta = meta
	}

	return &createObjectSpec, nil
}

type UpdateUserSpec struct {
	Email *string `json:"email"`
}

func (updateSpec UpdateUserSpec) ToUpdateObjectSpec() *object.UpdateObjectSpec {
	meta := make(map[string]interface{})

	if updateSpec.Email != nil {
		meta["email"] = updateSpec.Email
	}

	return &object.UpdateObjectSpec{
		Meta: meta,
	}
}
