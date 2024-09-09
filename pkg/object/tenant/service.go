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

package tenant

import (
	"context"

	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	object "github.com/warrant-dev/warrant/pkg/object"
	"github.com/warrant-dev/warrant/pkg/service"
)

type TenantService struct {
	service.BaseService
	objectSvc object.Service
}

func NewService(env service.Env, objectSvc object.Service) *TenantService {
	return &TenantService{
		BaseService: service.NewBaseService(env),
		objectSvc:   objectSvc,
	}
}

func (svc TenantService) Create(ctx context.Context, tenantSpec TenantSpec) (*TenantSpec, error) {
	var createdTenantSpec *TenantSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		objectSpec, err := tenantSpec.ToCreateObjectSpec()
		if err != nil {
			return err
		}

		createdObjectSpec, err := svc.objectSvc.Create(txCtx, *objectSpec)
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
	objectSpec, err := svc.objectSvc.GetByObjectTypeAndId(ctx, objecttype.ObjectTypeTenant, tenantId)
	if err != nil {
		return nil, err
	}

	return NewTenantSpecFromObjectSpec(objectSpec)
}

func (svc TenantService) List(ctx context.Context, listParams service.ListParams) ([]TenantSpec, error) {
	tenantSpecs := make([]TenantSpec, 0)
	objectSpecs, _, _, err := svc.objectSvc.List(ctx, &object.FilterOptions{ObjectType: objecttype.ObjectTypeTenant}, listParams)
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
		updatedObjectSpec, err := svc.objectSvc.UpdateByObjectTypeAndId(txCtx, objecttype.ObjectTypeTenant, tenantId, *tenantSpec.ToUpdateObjectSpec())
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
	_, err := svc.objectSvc.DeleteByObjectTypeAndId(ctx, objecttype.ObjectTypeTenant, tenantId)
	if err != nil {
		return err
	}

	return nil
}
