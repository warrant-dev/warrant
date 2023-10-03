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

	"github.com/warrant-dev/warrant/pkg/wookie"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

type Service interface {
	Create(ctx context.Context, objectSpec CreateObjectSpec) (*ObjectSpec, error)
	GetByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (*ObjectSpec, error)
	BatchGetByObjectTypeAndIds(ctx context.Context, objectType string, objectIds []string) ([]ObjectSpec, error)
	List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]ObjectSpec, error)
	UpdateByObjectTypeAndId(ctx context.Context, objectType string, objectId string, updateSpec UpdateObjectSpec) (*ObjectSpec, error)
	DeleteByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (*wookie.Token, error)
}

type ObjectService struct {
	service.BaseService
	Repository ObjectRepository
	EventSvc   event.Service
}

func NewService(env service.Env, repository ObjectRepository, eventSvc event.Service) *ObjectService {
	return &ObjectService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
		EventSvc:    eventSvc,
	}
}

func (svc ObjectService) Create(ctx context.Context, objectSpec CreateObjectSpec) (*ObjectSpec, error) {
	if objectSpec.ObjectId == "" {
		// generate an id for the object if one isn't supplied
		generatedUUID, err := uuid.NewRandom()
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

		newObjectId, err := svc.Repository.Create(txCtx, newObject)
		if err != nil {
			return err
		}

		createdObject, err = svc.Repository.GetById(txCtx, newObjectId)
		if err != nil {
			return err
		}

		createdObjectSpec, err := createdObject.ToObjectSpec()
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceCreated(txCtx, createdObject.GetObjectType(), createdObject.GetObjectId(), createdObjectSpec.Meta)
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
	object, err := svc.Repository.GetByObjectTypeAndId(ctx, objectType, objectId)
	if err != nil {
		return nil, err
	}

	return object.ToObjectSpec()
}

func (svc ObjectService) BatchGetByObjectTypeAndIds(ctx context.Context, objectType string, objectIds []string) ([]ObjectSpec, error) {
	objects, err := svc.Repository.BatchGetByObjectTypeAndIds(ctx, objectType, objectIds)
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

func (svc ObjectService) List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]ObjectSpec, error) {
	objectSpecs := make([]ObjectSpec, 0)
	objects, err := svc.Repository.List(ctx, filterOptions, listParams)
	if err != nil {
		return objectSpecs, err
	}

	for _, object := range objects {
		objectSpec, err := object.ToObjectSpec()
		if err != nil {
			return objectSpecs, err
		}

		objectSpecs = append(objectSpecs, *objectSpec)
	}

	return objectSpecs, nil
}

func (svc ObjectService) UpdateByObjectTypeAndId(ctx context.Context, objectType string, objectId string, updateSpec UpdateObjectSpec) (*ObjectSpec, error) {
	var updatedObjectSpec *ObjectSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		currentObject, err := svc.Repository.GetByObjectTypeAndId(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		err = currentObject.SetMeta(updateSpec.Meta)
		if err != nil {
			return err
		}

		err = svc.Repository.UpdateByObjectTypeAndId(txCtx, objectType, objectId, currentObject)
		if err != nil {
			return err
		}

		updatedObject, err := svc.Repository.GetByObjectTypeAndId(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceUpdated(txCtx, objectType, objectId, updateSpec.Meta)
		if err != nil {
			return err
		}

		updatedObjectSpec, err = updatedObject.ToObjectSpec()
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedObjectSpec, nil
}

func (svc ObjectService) DeleteByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (*wookie.Token, error) {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		objectSpec, err := svc.GetByObjectTypeAndId(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		err = svc.Repository.DeleteByObjectTypeAndId(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		// Delete existing warrants where this object is the object
		err = svc.Repository.DeleteWarrantsMatchingObject(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		// Delete existing warrants where this object is the subject
		err = svc.Repository.DeleteWarrantsMatchingSubject(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceDeleted(txCtx, objectType, objectId, objectSpec.Meta)
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
