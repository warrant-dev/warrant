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

	"github.com/google/uuid"
	"github.com/pkg/errors"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

type ObjectService struct {
	service.BaseService
	Repository ObjectRepository
	EventSvc   event.Service
	WarrantSvc *warrant.WarrantService
}

func NewService(env service.Env, repository ObjectRepository, eventSvc event.Service, warrantSvc *warrant.WarrantService) *ObjectService {
	return &ObjectService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
		EventSvc:    eventSvc,
		WarrantSvc:  warrantSvc,
	}
}

func (svc ObjectService) Create(ctx context.Context, objectSpec CreateObjectSpec) (*ObjectSpec, error) {
	if objectSpec.ObjectId == "" {
		// generate an id for the tenant if one isn't supplied
		generatedUUID, err := uuid.NewRandom()
		if err != nil {
			return nil, errors.New("unable to generate random UUID for object")
		}
		objectSpec.ObjectId = generatedUUID.String()
	}

	var newObject Model
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		_, err := svc.Repository.GetByObjectTypeAndId(txCtx, objectSpec.ObjectType, objectSpec.ObjectId)
		if err == nil {
			return service.NewDuplicateRecordError("Object", fmt.Sprintf("%s:%s", objectSpec.ObjectType, objectSpec.ObjectId), "An object with the given objectType and objectId already exists")
		}

		newObjectId, err := svc.Repository.Create(txCtx, *objectSpec.ToObject())
		if err != nil {
			return err
		}

		newObject, err = svc.Repository.GetById(txCtx, newObjectId)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newObject.ToObjectSpec(), nil
}

func (svc ObjectService) DeleteByObjectTypeAndId(ctx context.Context, objectType string, objectId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		err := svc.Repository.DeleteByObjectTypeAndId(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		err = svc.WarrantSvc.DeleteRelatedWarrants(txCtx, objectType, objectId)
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

func (svc ObjectService) GetByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (*ObjectSpec, error) {
	object, err := svc.Repository.GetByObjectTypeAndId(ctx, objectType, objectId)
	if err != nil {
		return nil, err
	}

	return object.ToObjectSpec(), nil
}

func (svc ObjectService) List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]ObjectSpec, error) {
	objectSpecs := make([]ObjectSpec, 0)
	objects, err := svc.Repository.List(ctx, filterOptions, listParams)
	if err != nil {
		return objectSpecs, err
	}

	for _, object := range objects {
		objectSpecs = append(objectSpecs, *object.ToObjectSpec())
	}

	return objectSpecs, nil
}
