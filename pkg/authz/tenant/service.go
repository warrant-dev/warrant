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

	"github.com/google/uuid"
	"github.com/pkg/errors"
	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypeTenant = "tenant"

type TenantService struct {
	service.BaseService
	Repository TenantRepository
	EventSvc   event.Service
	ObjectSvc  *object.ObjectService
}

func NewService(env service.Env, repository TenantRepository, eventSvc event.Service, objectSvc *object.ObjectService) *TenantService {
	return &TenantService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
		EventSvc:    eventSvc,
		ObjectSvc:   objectSvc,
	}
}

func (svc TenantService) Create(ctx context.Context, tenantSpec TenantSpec) (*TenantSpec, error) {
	if tenantSpec.TenantId == "" {
		// generate an id for the tenant if one isn't supplied
		generatedUUID, err := uuid.NewRandom()
		if err != nil {
			return nil, errors.New("unable to generate random UUID for tenant")
		}
		tenantSpec.TenantId = generatedUUID.String()
	}

	var newTenant Model
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		createdObject, err := svc.ObjectSvc.Create(txCtx, *tenantSpec.ToObjectSpec())
		if err != nil {
			switch err.(type) {
			case *service.DuplicateRecordError:
				return service.NewDuplicateRecordError("Tenant", tenantSpec.TenantId, "A tenant with the given tenantId already exists")
			default:
				return err
			}
		}

		_, err = svc.Repository.GetByTenantId(txCtx, tenantSpec.TenantId)
		if err == nil {
			return service.NewDuplicateRecordError("Tenant", tenantSpec.TenantId, "A tenant with the given tenantId already exists")
		}

		newTenantId, err := svc.Repository.Create(txCtx, tenantSpec.ToTenant(createdObject.ID))
		if err != nil {
			return err
		}

		newTenant, err = svc.Repository.GetById(txCtx, newTenantId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceCreated(txCtx, ResourceTypeTenant, newTenant.GetTenantId(), newTenant.ToTenantSpec())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newTenant.ToTenantSpec(), nil
}

func (svc TenantService) GetByTenantId(ctx context.Context, tenantId string) (*TenantSpec, error) {
	tenant, err := svc.Repository.GetByTenantId(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	return tenant.ToTenantSpec(), nil
}

func (svc TenantService) List(ctx context.Context, listParams service.ListParams) ([]TenantSpec, error) {
	tenantSpecs := make([]TenantSpec, 0)

	tenants, err := svc.Repository.List(ctx, listParams)
	if err != nil {
		return tenantSpecs, nil
	}

	for _, tenant := range tenants {
		tenantSpecs = append(tenantSpecs, *tenant.ToTenantSpec())
	}

	return tenantSpecs, nil
}

func (svc TenantService) UpdateByTenantId(ctx context.Context, tenantId string, tenantSpec UpdateTenantSpec) (*TenantSpec, error) {
	var updatedTenantSpec *TenantSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		currentTenant, err := svc.Repository.GetByTenantId(txCtx, tenantId)
		if err != nil {
			return err
		}

		currentTenant.SetName(tenantSpec.Name)
		err = svc.Repository.UpdateByTenantId(txCtx, tenantId, currentTenant)
		if err != nil {
			return err
		}

		updatedTenant, err := svc.Repository.GetByTenantId(txCtx, tenantId)
		if err != nil {
			return err
		}

		updatedTenantSpec = updatedTenant.ToTenantSpec()
		err = svc.EventSvc.TrackResourceUpdated(txCtx, ResourceTypeTenant, updatedTenant.GetTenantId(), updatedTenantSpec)
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
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		err := svc.Repository.DeleteByTenantId(txCtx, tenantId)
		if err != nil {
			return err
		}

		err = svc.ObjectSvc.DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypeTenant, tenantId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceDeleted(txCtx, ResourceTypeTenant, tenantId, nil)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
