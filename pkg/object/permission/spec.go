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

type PermissionSpec struct {
	PermissionId string    `json:"permissionId" validate:"required,valid_object_id"`
	Name         *string   `json:"name"`
	Description  *string   `json:"description"`
	CreatedAt    time.Time `json:"createdAt"`
}

func NewPermissionSpecFromObjectSpec(objectSpec *object.ObjectSpec) (*PermissionSpec, error) {
	var (
		name        *string
		description *string
	)

	if objectSpec.Meta != nil {
		if _, exists := objectSpec.Meta["name"]; exists {
			nameStr, ok := objectSpec.Meta["name"].(string)
			if !ok {
				return nil, errors.New("permission name has invalid type in object meta")
			}
			name = &nameStr
		}

		if _, exists := objectSpec.Meta["description"]; exists {
			descriptionStr, ok := objectSpec.Meta["description"].(string)
			if !ok {
				return nil, errors.New("permission description has invalid type in object meta")
			}
			description = &descriptionStr
		}
	}

	return &PermissionSpec{
		PermissionId: objectSpec.ObjectId,
		Name:         name,
		Description:  description,
		CreatedAt:    objectSpec.CreatedAt,
	}, nil
}

func (spec PermissionSpec) ToCreateObjectSpec() (*object.CreateObjectSpec, error) {
	createObjectSpec := object.CreateObjectSpec{
		ObjectType: objecttype.ObjectTypePermission,
		ObjectId:   spec.PermissionId,
	}

	meta := make(map[string]interface{})
	if spec.Name != nil {
		meta["name"] = spec.Name
	}

	if spec.Description != nil {
		meta["description"] = spec.Description
	}

	if len(meta) > 0 {
		createObjectSpec.Meta = meta
	}

	return &createObjectSpec, nil
}

type UpdatePermissionSpec struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (updateSpec UpdatePermissionSpec) ToUpdateObjectSpec() *object.UpdateObjectSpec {
	meta := make(map[string]interface{})

	if updateSpec.Name != nil {
		meta["name"] = updateSpec.Name
	}

	if updateSpec.Description != nil {
		meta["description"] = updateSpec.Description
	}

	return &object.UpdateObjectSpec{
		Meta: meta,
	}
}
