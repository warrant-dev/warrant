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
)

type Model interface {
	GetID() int64
	GetObjectId() int64
	GetTenantId() string
	GetName() *string
	SetName(newName *string)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
	ToTenantSpec() *TenantSpec
}

type Tenant struct {
	ID        int64      `mysql:"id" postgres:"id" sqlite:"id"`
	ObjectId  int64      `mysql:"objectId" postgres:"object_id" sqlite:"objectId"`
	TenantId  string     `mysql:"tenantId" postgres:"tenant_id" sqlite:"tenantId"`
	Name      *string    `mysql:"name" postgres:"name" sqlite:"name"`
	CreatedAt time.Time  `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
	UpdatedAt time.Time  `mysql:"updatedAt" postgres:"updated_at" sqlite:"updatedAt"`
	DeletedAt *time.Time `mysql:"deletedAt" postgres:"deleted_at" sqlite:"deletedAt"`
}

func (tenant Tenant) GetID() int64 {
	return tenant.ID
}

func (tenant Tenant) GetObjectId() int64 {
	return tenant.ObjectId
}

func (tenant Tenant) GetTenantId() string {
	return tenant.TenantId
}

func (tenant Tenant) GetName() *string {
	return tenant.Name
}

func (tenant *Tenant) SetName(newName *string) {
	tenant.Name = newName
}

func (tenant Tenant) GetCreatedAt() time.Time {
	return tenant.CreatedAt
}

func (tenant Tenant) GetUpdatedAt() time.Time {
	return tenant.UpdatedAt
}

func (tenant Tenant) GetDeletedAt() *time.Time {
	return tenant.DeletedAt
}

func (tenant Tenant) ToTenantSpec() *TenantSpec {
	return &TenantSpec{
		TenantId:  tenant.TenantId,
		Name:      tenant.Name,
		CreatedAt: tenant.CreatedAt,
	}
}
