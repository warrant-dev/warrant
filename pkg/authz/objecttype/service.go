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

package authz

import (
	"context"

	"github.com/warrant-dev/warrant/pkg/service"
	"github.com/warrant-dev/warrant/pkg/wookie"
)

type Service interface {
	Create(ctx context.Context, spec CreateObjectTypeSpec) (*ObjectTypeSpec, *wookie.Token, error)
	GetByTypeId(ctx context.Context, typeId string) (*ObjectTypeSpec, error)
	List(ctx context.Context, listParams service.ListParams) ([]ObjectTypeSpec, *service.Cursor, *service.Cursor, error)
	UpdateByTypeId(ctx context.Context, typeId string, spec UpdateObjectTypeSpec) (*ObjectTypeSpec, *wookie.Token, error)
	DeleteByTypeId(ctx context.Context, typeId string) (*wookie.Token, error)
}

type ObjectTypeService struct {
	service.BaseService
	repository ObjectTypeRepository
}

func NewService(env service.Env, repository ObjectTypeRepository) *ObjectTypeService {
	return &ObjectTypeService{
		BaseService: service.NewBaseService(env),
		repository:  repository,
	}
}

func (svc ObjectTypeService) Create(ctx context.Context, spec CreateObjectTypeSpec) (*ObjectTypeSpec, *wookie.Token, error) {
	var newObjectTypeSpec *ObjectTypeSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		objectType, err := spec.ToObjectType()
		if err != nil {
			return err
		}

		newObjectTypeId, err := svc.repository.Create(txCtx, objectType)
		if err != nil {
			return err
		}

		newObjectType, err := svc.repository.GetById(txCtx, newObjectTypeId)
		if err != nil {
			return err
		}

		newObjectTypeSpec, err = newObjectType.ToObjectTypeSpec()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return newObjectTypeSpec, nil, nil
}

func (svc ObjectTypeService) GetByTypeId(ctx context.Context, typeId string) (*ObjectTypeSpec, error) {
	objectType, err := svc.repository.GetByTypeId(ctx, typeId)
	if err != nil {
		return nil, err
	}

	objectTypeSpec, err := objectType.ToObjectTypeSpec()
	if err != nil {
		return nil, err
	}

	return objectTypeSpec, nil
}

func (svc ObjectTypeService) List(ctx context.Context, listParams service.ListParams) ([]ObjectTypeSpec, *service.Cursor, *service.Cursor, error) {
	objectTypeSpecs := make([]ObjectTypeSpec, 0)
	objectTypes, prevCursor, nextCursor, err := svc.repository.List(ctx, listParams)
	if err != nil {
		return objectTypeSpecs, prevCursor, nextCursor, err
	}

	for _, objectType := range objectTypes {
		objectTypeSpec, err := objectType.ToObjectTypeSpec()
		if err != nil {
			return nil, prevCursor, nextCursor, err
		}

		objectTypeSpecs = append(objectTypeSpecs, *objectTypeSpec)
	}

	return objectTypeSpecs, prevCursor, nextCursor, nil
}

func (svc ObjectTypeService) UpdateByTypeId(ctx context.Context, typeId string, spec UpdateObjectTypeSpec) (*ObjectTypeSpec, *wookie.Token, error) {
	var updatedObjectTypeSpec *ObjectTypeSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		currentObjectType, err := svc.repository.GetByTypeId(txCtx, typeId)
		if err != nil {
			return err
		}

		updateTo, err := spec.ToObjectType(typeId)
		if err != nil {
			return err
		}
		currentObjectType.SetDefinition(updateTo.Definition)

		err = svc.repository.UpdateByTypeId(txCtx, typeId, currentObjectType)
		if err != nil {
			return err
		}

		updatedObjectTypeSpec, err = svc.GetByTypeId(txCtx, typeId)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return updatedObjectTypeSpec, nil, nil
}

func (svc ObjectTypeService) DeleteByTypeId(ctx context.Context, typeId string) (*wookie.Token, error) {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		err := svc.repository.DeleteByTypeId(txCtx, typeId)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	//nolint:nilnil
	return nil, nil
}
