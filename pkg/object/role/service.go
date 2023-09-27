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
	"context"

	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/event"
	object "github.com/warrant-dev/warrant/pkg/object"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypeRole = "role"

type RoleService struct {
	service.BaseService
	EventSvc  event.Service
	ObjectSvc object.Service
}

func NewService(env service.Env, eventSvc event.Service, objectSvc object.Service) *RoleService {
	return &RoleService{
		BaseService: service.NewBaseService(env),
		EventSvc:    eventSvc,
		ObjectSvc:   objectSvc,
	}
}

func (svc RoleService) Create(ctx context.Context, roleSpec RoleSpec) (*RoleSpec, error) {
	var createdRoleSpec *RoleSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		objectSpec, err := roleSpec.ToCreateObjectSpec()
		if err != nil {
			return err
		}

		createdObjectSpec, err := svc.ObjectSvc.Create(txCtx, *objectSpec)
		if err != nil {
			return err
		}

		createdRoleSpec, err = NewRoleSpecFromObjectSpec(createdObjectSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdRoleSpec, nil
}

func (svc RoleService) GetByRoleId(ctx context.Context, roleId string) (*RoleSpec, error) {
	objectSpec, err := svc.ObjectSvc.GetByObjectTypeAndId(ctx, objecttype.ObjectTypeRole, roleId)
	if err != nil {
		return nil, err
	}

	return NewRoleSpecFromObjectSpec(objectSpec)
}

func (svc RoleService) List(ctx context.Context, listParams service.ListParams) ([]RoleSpec, error) {
	roleSpecs := make([]RoleSpec, 0)
	objectSpecs, err := svc.ObjectSvc.List(ctx, &object.FilterOptions{ObjectType: objecttype.ObjectTypeRole}, listParams)
	if err != nil {
		return roleSpecs, err
	}

	for i := range objectSpecs {
		roleSpec, err := NewRoleSpecFromObjectSpec(&objectSpecs[i])
		if err != nil {
			return roleSpecs, err
		}

		roleSpecs = append(roleSpecs, *roleSpec)
	}

	return roleSpecs, nil
}

func (svc RoleService) UpdateByRoleId(ctx context.Context, roleId string, roleSpec UpdateRoleSpec) (*RoleSpec, error) {
	var updatedRoleSpec *RoleSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		updatedObjectSpec, err := svc.ObjectSvc.UpdateByObjectTypeAndId(txCtx, objecttype.ObjectTypeRole, roleId, *roleSpec.ToUpdateObjectSpec())
		if err != nil {
			return err
		}

		updatedRoleSpec, err = NewRoleSpecFromObjectSpec(updatedObjectSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return updatedRoleSpec, nil
}

func (svc RoleService) DeleteByRoleId(ctx context.Context, roleId string) error {
	err := svc.ObjectSvc.DeleteByObjectTypeAndId(ctx, objecttype.ObjectTypeRole, roleId)
	if err != nil {
		return err
	}

	return nil
}
