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

	"github.com/warrant-dev/warrant/pkg/wookie"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/service"
)

type Service interface {
	Create(ctx context.Context, objectSpec CreateObjectSpec) (*ObjectSpec, error)
	GetByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (*ObjectSpec, error)
	BatchGetByObjectTypeAndIds(ctx context.Context, objectType string, objectIds []string) ([]ObjectSpec, error)
	List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]ObjectSpec, *service.Cursor, *service.Cursor, error)
	UpdateByObjectTypeAndId(ctx context.Context, objectType string, objectId string, updateSpec UpdateObjectSpec) (*ObjectSpec, error)
	DeleteByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (*wookie.Token, error)
}

type ObjectService struct {
	service.BaseService
	repository ObjectRepository
}

func NewService(env service.Env, repository ObjectRepository) *ObjectService {
	return &ObjectService{
		BaseService: service.NewBaseService(env),
		repository:  repository,
	}
}

func (svc ObjectService) Create(ctx context.Context, objectSpec CreateObjectSpec) (*ObjectSpec, error) {
	if objectSpec.ObjectId == "" {
		// generate an id for the object if one isn't supplied
		generatedUUID, err := uuid.NewV7()
		if err != nil {
			return nil, errors.New("unable to generate random UUID for object")
		}
		objectSpec.ObjectId = generatedUUID.String()
	}

	var createdObject Model
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		newObject, err := objectSpec.ToObject()
		if err != nil {
			return err
		}

		newObjectId, err := svc.repository.Create(txCtx, newObject)
		if err != nil {
			return err
		}

		createdObject, err = svc.repository.GetById(txCtx, newObjectId)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdObject.ToObjectSpec()
}

func (svc ObjectService) GetByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (*ObjectSpec, error) {
	object, err := svc.repository.GetByObjectTypeAndId(ctx, objectType, objectId)
	if err != nil {
		return nil, err
	}

	return object.ToObjectSpec()
}

func (svc ObjectService) BatchGetByObjectTypeAndIds(ctx context.Context, objectType string, objectIds []string) ([]ObjectSpec, error) {
	objects, err := svc.repository.BatchGetByObjectTypeAndIds(ctx, objectType, objectIds)
	if err != nil {
		return nil, err
	}

	objectSpecs := make([]ObjectSpec, 0)
	for _, object := range objects {
		objectSpec, err := object.ToObjectSpec()
		if err != nil {
			return nil, err
		}

		objectSpecs = append(objectSpecs, *objectSpec)
	}

	return objectSpecs, nil
}

func (svc ObjectService) List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]ObjectSpec, *service.Cursor, *service.Cursor, error) {
	objectSpecs := make([]ObjectSpec, 0)
	objects, prevCursor, nextCursor, err := svc.repository.List(ctx, filterOptions, listParams)
	if err != nil {
		return objectSpecs, prevCursor, nextCursor, err
	}

	for _, object := range objects {
		objectSpec, err := object.ToObjectSpec()
		if err != nil {
			return objectSpecs, prevCursor, nextCursor, err
		}

		objectSpecs = append(objectSpecs, *objectSpec)
	}

	return objectSpecs, prevCursor, nextCursor, nil
}

func (svc ObjectService) UpdateByObjectTypeAndId(ctx context.Context, objectType string, objectId string, updateSpec UpdateObjectSpec) (*ObjectSpec, error) {
	var updatedObject Model
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		currentObject, err := svc.repository.GetByObjectTypeAndId(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		err = currentObject.SetMeta(updateSpec.Meta)
		if err != nil {
			return err
		}

		err = svc.repository.UpdateByObjectTypeAndId(txCtx, objectType, objectId, currentObject)
		if err != nil {
			return err
		}

		updatedObject, err = svc.repository.GetByObjectTypeAndId(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedObject.ToObjectSpec()
}

func (svc ObjectService) DeleteByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (*wookie.Token, error) {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		_, err := svc.GetByObjectTypeAndId(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		err = svc.repository.DeleteByObjectTypeAndId(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		// Delete existing warrants where this object is the object
		err = svc.repository.DeleteWarrantsMatchingObject(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		// Delete existing warrants where this object is the subject
		err = svc.repository.DeleteWarrantsMatchingSubject(txCtx, objectType, objectId)
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
