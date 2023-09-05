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
	"context"

	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypeTenant = "tenant"

type TenantService struct {
	service.BaseService
	EventSvc  event.Service
	ObjectSvc *object.ObjectService
}

func NewService(env service.Env, eventSvc event.Service, objectSvc *object.ObjectService) *TenantService {
	return &TenantService{
		BaseService: service.NewBaseService(env),
		EventSvc:    eventSvc,
		ObjectSvc:   objectSvc,
	}
}

func (svc TenantService) Create(ctx context.Context, tenantSpec TenantSpec) (*TenantSpec, error) {
	var createdTenantSpec *TenantSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		objectSpec, err := tenantSpec.ToCreateObjectSpec()
		if err != nil {
			return err
		}

		createdObjectSpec, err := svc.ObjectSvc.Create(txCtx, *objectSpec)
		if err != nil {
			return err
		}

		createdTenantSpec, err = NewTenantSpecFromObjectSpec(createdObjectSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdTenantSpec, nil
}

func (svc TenantService) GetByTenantId(ctx context.Context, tenantId string) (*TenantSpec, error) {
	objectSpec, err := svc.ObjectSvc.GetByObjectTypeAndId(ctx, objecttype.ObjectTypeTenant, tenantId)
	if err != nil {
		return nil, err
	}

	return NewTenantSpecFromObjectSpec(objectSpec)
}

func (svc TenantService) List(ctx context.Context, listParams service.ListParams) ([]TenantSpec, error) {
	tenantSpecs := make([]TenantSpec, 0)
	objectSpecs, err := svc.ObjectSvc.List(ctx, &object.FilterOptions{ObjectType: objecttype.ObjectTypeTenant}, listParams)
	if err != nil {
		return tenantSpecs, err
	}

	for i := range objectSpecs {
		tenantSpec, err := NewTenantSpecFromObjectSpec(&objectSpecs[i])
		if err != nil {
			return tenantSpecs, err
		}

		tenantSpecs = append(tenantSpecs, *tenantSpec)
	}

	return tenantSpecs, nil
}

func (svc TenantService) UpdateByTenantId(ctx context.Context, tenantId string, tenantSpec UpdateTenantSpec) (*TenantSpec, error) {
	var updatedTenantSpec *TenantSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		updatedObjectSpec, err := svc.ObjectSvc.UpdateByObjectTypeAndId(txCtx, objecttype.ObjectTypeTenant, tenantId, *tenantSpec.ToUpdateObjectSpec())
		if err != nil {
			return err
		}

		updatedTenantSpec, err = NewTenantSpecFromObjectSpec(updatedObjectSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return updatedTenantSpec, nil
}

func (svc TenantService) DeleteByTenantId(ctx context.Context, tenantId string) error {
	err := svc.ObjectSvc.DeleteByObjectTypeAndId(ctx, objecttype.ObjectTypeTenant, tenantId)
	if err != nil {
		return err
	}

	return nil
}
