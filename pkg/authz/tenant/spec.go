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
	"time"

	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
)

type TenantSpec struct {
	TenantId  string    `json:"tenantId" validate:"omitempty,valid_object_id"`
	Name      *string   `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

func (spec TenantSpec) ToTenant(objectId int64) *Tenant {
	return &Tenant{
		ObjectId: objectId,
		TenantId: spec.TenantId,
		Name:     spec.Name,
	}
}

func (spec TenantSpec) ToObjectSpec() *object.ObjectSpec {
	return &object.ObjectSpec{
		ObjectType: objecttype.ObjectTypeTenant,
		ObjectId:   spec.TenantId,
	}
}

type UpdateTenantSpec struct {
	Name *string `json:"name"`
}
