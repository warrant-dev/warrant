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

package tenant

import (
	"errors"
	"time"

	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
)

type TenantSpec struct {
	TenantId  string    `json:"tenantId" validate:"omitempty,valid_object_id"`
	Name      *string   `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewTenantSpecFromObjectSpec(objectSpec *object.ObjectSpec) (*TenantSpec, error) {
	var name *string

	if objectSpec.Meta != nil {
		if _, exists := objectSpec.Meta["name"]; exists {
			nameStr, ok := objectSpec.Meta["name"].(string)
			if !ok {
				return nil, errors.New("tenant name has invalid type in object meta")
			}
			name = &nameStr
		}
	}

	return &TenantSpec{
		TenantId:  objectSpec.ObjectId,
		Name:      name,
		CreatedAt: objectSpec.CreatedAt,
	}, nil
}

func (spec TenantSpec) ToCreateObjectSpec() (*object.CreateObjectSpec, error) {
	createObjectSpec := object.CreateObjectSpec{
		ObjectType: objecttype.ObjectTypeTenant,
		ObjectId:   spec.TenantId,
	}

	meta := make(map[string]interface{})
	if spec.Name != nil {
		meta["name"] = spec.Name
	}

	if len(meta) > 0 {
		createObjectSpec.Meta = meta
	}

	return &createObjectSpec, nil
}

type UpdateTenantSpec struct {
	Name *string `json:"name"`
}

func (updateSpec UpdateTenantSpec) ToUpdateObjectSpec() *object.UpdateObjectSpec {
	meta := make(map[string]interface{})

	if updateSpec.Name != nil {
		meta["name"] = updateSpec.Name
	}

	return &object.UpdateObjectSpec{
		Meta: meta,
	}
}
