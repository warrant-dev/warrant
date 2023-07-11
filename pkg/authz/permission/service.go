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
	"context"

	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypePermission = "permission"

type PermissionService struct {
	service.BaseService
	Repository PermissionRepository
	EventSvc   event.Service
	ObjectSvc  *object.ObjectService
}

func NewService(env service.Env, repository PermissionRepository, eventSvc event.Service, objectSvc *object.ObjectService) *PermissionService {
	return &PermissionService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
		EventSvc:    eventSvc,
		ObjectSvc:   objectSvc,
	}
}

func (svc PermissionService) Create(ctx context.Context, permissionSpec PermissionSpec) (*PermissionSpec, error) {
	var newPermission Model
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		createdObject, err := svc.ObjectSvc.Create(txCtx, *permissionSpec.ToObjectSpec())
		if err != nil {
			return err
		}

		_, err = svc.Repository.GetByPermissionId(txCtx, permissionSpec.PermissionId)
		if err == nil {
			return service.NewDuplicateRecordError("Permission", permissionSpec.PermissionId, "A permission with the given permissionId already exists")
		}

		newPermissionId, err := svc.Repository.Create(txCtx, permissionSpec.ToPermission(createdObject.ID))
		if err != nil {
			return err
		}

		newPermission, err = svc.Repository.GetById(txCtx, newPermissionId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceCreated(txCtx, ResourceTypePermission, newPermission.GetPermissionId(), newPermission.ToPermissionSpec())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newPermission.ToPermissionSpec(), nil
}

func (svc PermissionService) GetByPermissionId(ctx context.Context, permissionId string) (*PermissionSpec, error) {
	permission, err := svc.Repository.GetByPermissionId(ctx, permissionId)
	if err != nil {
		return nil, err
	}

	return permission.ToPermissionSpec(), nil
}

func (svc PermissionService) List(ctx context.Context, listParams service.ListParams) ([]PermissionSpec, error) {
	permissionSpecs := make([]PermissionSpec, 0)

	permissions, err := svc.Repository.List(ctx, listParams)
	if err != nil {
		return permissionSpecs, nil
	}

	for _, permission := range permissions {
		permissionSpecs = append(permissionSpecs, *permission.ToPermissionSpec())
	}

	return permissionSpecs, nil
}

func (svc PermissionService) UpdateByPermissionId(ctx context.Context, permissionId string, permissionSpec UpdatePermissionSpec) (*PermissionSpec, error) {
	currentPermission, err := svc.Repository.GetByPermissionId(ctx, permissionId)
	if err != nil {
		return nil, err
	}

	currentPermission.SetName(permissionSpec.Name)
	currentPermission.SetDescription(permissionSpec.Description)
	err = svc.Repository.UpdateByPermissionId(ctx, permissionId, currentPermission)
	if err != nil {
		return nil, err
	}

	updatedPermission, err := svc.Repository.GetByPermissionId(ctx, permissionId)
	if err != nil {
		return nil, err
	}

	updatedPermissionSpec := updatedPermission.ToPermissionSpec()
	err = svc.EventSvc.TrackResourceUpdated(ctx, ResourceTypePermission, updatedPermission.GetPermissionId(), updatedPermissionSpec)
	if err != nil {
		return nil, err
	}

	return updatedPermissionSpec, nil
}

func (svc PermissionService) DeleteByPermissionId(ctx context.Context, permissionId string) (*wookie.Token, error) {
	var newWookie *wookie.Token
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		err := svc.Repository.DeleteByPermissionId(txCtx, permissionId)
		if err != nil {
			return err
		}

		newWookie, err = svc.ObjectSvc.DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypePermission, permissionId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceDeleted(txCtx, ResourceTypePermission, permissionId, nil)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newWookie, nil
}
