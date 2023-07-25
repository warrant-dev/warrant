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
	"fmt"

	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypeObjectType = "object-type"

type ObjectTypeMap struct {
	objectTypes map[string]ObjectTypeSpec
}

func (m ObjectTypeMap) GetByTypeId(typeId string) (*ObjectTypeSpec, error) {
	if val, ok := m.objectTypes[typeId]; ok {
		return &val, nil
	}
	return nil, fmt.Errorf("no object type with typeId %s exists", typeId)
}

type ObjectTypeService struct {
	service.BaseService
	Repository ObjectTypeRepository
	EventSvc   event.Service
	WookieSvc  *wookie.WookieService
}

func NewService(env service.Env, repository ObjectTypeRepository, eventSvc event.Service, wookieSvc *wookie.WookieService) *ObjectTypeService {
	return &ObjectTypeService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
		EventSvc:    eventSvc,
		WookieSvc:   wookieSvc,
	}
}

func (svc ObjectTypeService) Create(ctx context.Context, objectTypeSpec ObjectTypeSpec) (*ObjectTypeSpec, *wookie.Token, error) {
	_, err := svc.Repository.GetByTypeId(ctx, objectTypeSpec.Type)
	if err == nil {
		return nil, nil, service.NewDuplicateRecordError("ObjectType", objectTypeSpec.Type, "An objectType with the given type already exists")
	}

	objectType, err := objectTypeSpec.ToObjectType()
	if err != nil {
		return nil, nil, err
	}

	var newObjectTypeSpec *ObjectTypeSpec
	newWookie, e := svc.WookieSvc.WithWookieUpdate(ctx, func(txCtx context.Context) error {
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
	if e != nil {
		return nil, nil, e
	}
	return newObjectTypeSpec, newWookie, nil
}

func (svc ObjectTypeService) GetByTypeId(ctx context.Context, typeId string) (*ObjectTypeSpec, *wookie.Token, error) {
	var objectTypeSpec *ObjectTypeSpec
	newWookie, e := svc.WookieSvc.WookieSafeRead(ctx, func(wkCtx context.Context) error {
		objectType, err := svc.Repository.GetByTypeId(wkCtx, typeId)
		if err != nil {
			return err
		}

		objectTypeSpec, err = objectType.ToObjectTypeSpec()
		if err != nil {
			return err
		}
		return nil
	})
	if e != nil {
		return nil, nil, e
	}
	return objectTypeSpec, newWookie, e
}

func (svc ObjectTypeService) GetTypeMap(ctx context.Context) (*ObjectTypeMap, *wookie.Token, error) {
	typeMap := make(map[string]ObjectTypeSpec)
	newWookie, e := svc.WookieSvc.WookieSafeRead(ctx, func(wkCtx context.Context) error {
		objectTypes, err := svc.Repository.ListAll(wkCtx)
		if err != nil {
			return err
		}

		for _, objectType := range objectTypes {
			objectTypeSpec, err := objectType.ToObjectTypeSpec()
			if err != nil {
				return err
			}

			typeMap[objectTypeSpec.Type] = *objectTypeSpec
		}

		return nil
	})
	if e != nil {
		return nil, nil, e
	}
	return &ObjectTypeMap{
		objectTypes: typeMap,
	}, newWookie, nil
}

func (svc ObjectTypeService) List(ctx context.Context, listParams service.ListParams) ([]ObjectTypeSpec, *wookie.Token, error) {
	objectTypeSpecs := make([]ObjectTypeSpec, 0)
	newWookie, e := svc.WookieSvc.WookieSafeRead(ctx, func(wkCtx context.Context) error {
		objectTypes, err := svc.Repository.List(wkCtx, listParams)
		if err != nil {
			return err
		}

		for _, objectType := range objectTypes {
			objectTypeSpec, err := objectType.ToObjectTypeSpec()
			if err != nil {
				return err
			}

			objectTypeSpecs = append(objectTypeSpecs, *objectTypeSpec)
		}

		return nil
	})
	if e != nil {
		return nil, nil, e
	}
	return objectTypeSpecs, newWookie, nil
}

func (svc ObjectTypeService) UpdateByTypeId(ctx context.Context, typeId string, objectTypeSpec ObjectTypeSpec) (*ObjectTypeSpec, *wookie.Token, error) {
	currentObjectType, err := svc.Repository.GetByTypeId(ctx, typeId)
	if err != nil {
		return nil, nil, err
	}
	updateTo, err := objectTypeSpec.ToObjectType()
	if err != nil {
		return nil, nil, err
	}
	currentObjectType.SetDefinition(updateTo.Definition)
	var updatedObjectTypeSpec *ObjectTypeSpec
	newWookie, e := svc.WookieSvc.WithWookieUpdate(ctx, func(txCtx context.Context) error {
		err := svc.Repository.UpdateByTypeId(txCtx, typeId, currentObjectType)
		if err != nil {
			return err
		}

		updatedObjectTypeSpec, _, err = svc.GetByTypeId(txCtx, typeId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceUpdated(txCtx, ResourceTypeObjectType, typeId, updatedObjectTypeSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if e != nil {
		return nil, nil, e
	}
	return updatedObjectTypeSpec, newWookie, nil
}

func (svc ObjectTypeService) DeleteByTypeId(ctx context.Context, typeId string) (*wookie.Token, error) {
	newWookie, e := svc.WookieSvc.WithWookieUpdate(ctx, func(txCtx context.Context) error {
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
	if e != nil {
		return nil, e
	}
	return newWookie, nil
}
