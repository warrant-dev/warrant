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

	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
	"github.com/warrant-dev/warrant/pkg/wookie"
)

const ResourceTypeObjectType = "object-type"

type ObjectTypeMap map[string]ObjectTypeSpec

func (m ObjectTypeMap) GetByTypeId(typeId string) (*ObjectTypeSpec, error) {
	if val, ok := m[typeId]; ok {
		return &val, nil
	}

	return nil, service.NewRecordNotFoundError("ObjectType", typeId)
}

type Service interface {
	Create(ctx context.Context, objectTypeSpec ObjectTypeSpec) (*ObjectTypeSpec, *wookie.Token, error)
	GetByTypeId(ctx context.Context, typeId string) (*ObjectTypeSpec, error)
	GetTypeMap(ctx context.Context) (ObjectTypeMap, error)
	List(ctx context.Context, listParams service.ListParams) ([]ObjectTypeSpec, error)
	UpdateByTypeId(ctx context.Context, typeId string, objectTypeSpec ObjectTypeSpec) (*ObjectTypeSpec, *wookie.Token, error)
	DeleteByTypeId(ctx context.Context, typeId string) (*wookie.Token, error)
}

type ObjectTypeService struct {
	service.BaseService
	Repository ObjectTypeRepository
	EventSvc   event.Service
}

func NewService(env service.Env, repository ObjectTypeRepository, eventSvc event.Service) *ObjectTypeService {
	return &ObjectTypeService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
		EventSvc:    eventSvc,
	}
}

func (svc ObjectTypeService) Create(ctx context.Context, objectTypeSpec ObjectTypeSpec) (*ObjectTypeSpec, *wookie.Token, error) {
	var newObjectTypeSpec *ObjectTypeSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		objectType, err := objectTypeSpec.ToObjectType()
		if err != nil {
			return err
		}

		newObjectTypeId, err := svc.Repository.Create(txCtx, objectType)
		if err != nil {
			return err
		}

		newObjectType, err := svc.Repository.GetById(txCtx, newObjectTypeId)
		if err != nil {
			return err
		}

		newObjectTypeSpec, err = newObjectType.ToObjectTypeSpec()
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceCreated(txCtx, ResourceTypeObjectType, newObjectType.GetTypeId(), newObjectTypeSpec)
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
	objectType, err := svc.Repository.GetByTypeId(ctx, typeId)
	if err != nil {
		return nil, err
	}

	objectTypeSpec, err := objectType.ToObjectTypeSpec()
	if err != nil {
		return nil, err
	}

	return objectTypeSpec, nil
}

func (svc ObjectTypeService) GetTypeMap(ctx context.Context) (ObjectTypeMap, error) {
	objectTypes, err := svc.Repository.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	typeMap := make(map[string]ObjectTypeSpec)
	for _, objectType := range objectTypes {
		objectTypeSpec, err := objectType.ToObjectTypeSpec()
		if err != nil {
			return nil, err
		}

		typeMap[objectTypeSpec.Type] = *objectTypeSpec
	}

	return typeMap, nil
}

func (svc ObjectTypeService) List(ctx context.Context, listParams service.ListParams) ([]ObjectTypeSpec, error) {
	objectTypeSpecs := make([]ObjectTypeSpec, 0)

	objectTypes, err := svc.Repository.List(ctx, listParams)
	if err != nil {
		return objectTypeSpecs, err
	}

	for _, objectType := range objectTypes {
		objectTypeSpec, err := objectType.ToObjectTypeSpec()
		if err != nil {
			return nil, err
		}

		objectTypeSpecs = append(objectTypeSpecs, *objectTypeSpec)
	}

	return objectTypeSpecs, nil
}

func (svc ObjectTypeService) UpdateByTypeId(ctx context.Context, typeId string, objectTypeSpec ObjectTypeSpec) (*ObjectTypeSpec, *wookie.Token, error) {
	var updatedObjectTypeSpec *ObjectTypeSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		currentObjectType, err := svc.Repository.GetByTypeId(txCtx, typeId)
		if err != nil {
			return err
		}
		updateTo, err := objectTypeSpec.ToObjectType()
		if err != nil {
			return err
		}
		currentObjectType.SetDefinition(updateTo.Definition)

		err = svc.Repository.UpdateByTypeId(txCtx, typeId, currentObjectType)
		if err != nil {
			return err
		}

		updatedObjectTypeSpec, err = svc.GetByTypeId(txCtx, typeId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceUpdated(txCtx, ResourceTypeObjectType, typeId, updatedObjectTypeSpec)
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
		err := svc.Repository.DeleteByTypeId(txCtx, typeId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceDeleted(txCtx, ResourceTypeObjectType, typeId, nil)
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
