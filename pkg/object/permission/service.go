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

package object

import (
	"context"

	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	object "github.com/warrant-dev/warrant/pkg/object"
	"github.com/warrant-dev/warrant/pkg/service"
)

type PermissionService struct {
	service.BaseService
	objectSvc object.Service
}

func NewService(env service.Env, objectSvc object.Service) *PermissionService {
	return &PermissionService{
		BaseService: service.NewBaseService(env),
		objectSvc:   objectSvc,
	}
}

func (svc PermissionService) Create(ctx context.Context, permissionSpec PermissionSpec) (*PermissionSpec, error) {
	var createdPermissionSpec *PermissionSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		objectSpec, err := permissionSpec.ToCreateObjectSpec()
		if err != nil {
			return err
		}

		createdObjectSpec, err := svc.objectSvc.Create(txCtx, *objectSpec)
		if err != nil {
			return err
		}

		createdPermissionSpec, err = NewPermissionSpecFromObjectSpec(createdObjectSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdPermissionSpec, nil
}

func (svc PermissionService) GetByPermissionId(ctx context.Context, permissionId string) (*PermissionSpec, error) {
	objectSpec, err := svc.objectSvc.GetByObjectTypeAndId(ctx, objecttype.ObjectTypePermission, permissionId)
	if err != nil {
		return nil, err
	}

	return NewPermissionSpecFromObjectSpec(objectSpec)
}

func (svc PermissionService) List(ctx context.Context, listParams service.ListParams) ([]PermissionSpec, error) {
	permissionSpecs := make([]PermissionSpec, 0)
	objectSpecs, _, _, err := svc.objectSvc.List(ctx, &object.FilterOptions{ObjectType: objecttype.ObjectTypePermission}, listParams)
	if err != nil {
		return permissionSpecs, err
	}

	for i := range objectSpecs {
		permissionSpec, err := NewPermissionSpecFromObjectSpec(&objectSpecs[i])
		if err != nil {
			return permissionSpecs, err
		}

		permissionSpecs = append(permissionSpecs, *permissionSpec)
	}

	return permissionSpecs, nil
}

func (svc PermissionService) UpdateByPermissionId(ctx context.Context, permissionId string, permissionSpec UpdatePermissionSpec) (*PermissionSpec, error) {
	var updatedPermissionSpec *PermissionSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		updatedObjectSpec, err := svc.objectSvc.UpdateByObjectTypeAndId(txCtx, objecttype.ObjectTypePermission, permissionId, *permissionSpec.ToUpdateObjectSpec())
		if err != nil {
			return err
		}

		updatedPermissionSpec, err = NewPermissionSpecFromObjectSpec(updatedObjectSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return updatedPermissionSpec, nil
}

func (svc PermissionService) DeleteByPermissionId(ctx context.Context, permissionId string) error {
	_, err := svc.objectSvc.DeleteByObjectTypeAndId(ctx, objecttype.ObjectTypePermission, permissionId)
	if err != nil {
		return err
	}

	return nil
}
