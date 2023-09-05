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
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypeUser = "user"

type UserService struct {
	service.BaseService
	EventSvc  event.Service
	ObjectSvc *object.ObjectService
}

func NewService(env service.Env, eventSvc event.Service, objectSvc *object.ObjectService) *UserService {
	return &UserService{
		BaseService: service.NewBaseService(env),
		EventSvc:    eventSvc,
		ObjectSvc:   objectSvc,
	}
}

func (svc UserService) Create(ctx context.Context, userSpec UserSpec) (*UserSpec, error) {
	var createdUserSpec *UserSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		objectSpec, err := userSpec.ToCreateObjectSpec()
		if err != nil {
			return err
		}

		createdObjectSpec, err := svc.ObjectSvc.Create(txCtx, *objectSpec)
		if err != nil {
			return err
		}

		createdUserSpec, err = NewUserSpecFromObjectSpec(createdObjectSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdUserSpec, nil
}

func (svc UserService) GetByUserId(ctx context.Context, userId string) (*UserSpec, error) {
	objectSpec, err := svc.ObjectSvc.GetByObjectTypeAndId(ctx, objecttype.ObjectTypeUser, userId)
	if err != nil {
		return nil, err
	}

	return NewUserSpecFromObjectSpec(objectSpec)
}

func (svc UserService) List(ctx context.Context, listParams service.ListParams) ([]UserSpec, error) {
	userSpecs := make([]UserSpec, 0)
	objectSpecs, err := svc.ObjectSvc.List(ctx, &object.FilterOptions{ObjectType: objecttype.ObjectTypeUser}, listParams)
	if err != nil {
		return userSpecs, err
	}

	for i := range objectSpecs {
		userSpec, err := NewUserSpecFromObjectSpec(&objectSpecs[i])
		if err != nil {
			return userSpecs, err
		}

		userSpecs = append(userSpecs, *userSpec)
	}

	return userSpecs, nil
}

func (svc UserService) UpdateByUserId(ctx context.Context, userId string, userSpec UpdateUserSpec) (*UserSpec, error) {
	var updatedUserSpec *UserSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		updatedObjectSpec, err := svc.ObjectSvc.UpdateByObjectTypeAndId(txCtx, objecttype.ObjectTypeUser, userId, *userSpec.ToUpdateObjectSpec())
		if err != nil {
			return err
		}

		updatedUserSpec, err = NewUserSpecFromObjectSpec(updatedObjectSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return updatedUserSpec, nil
}

func (svc UserService) DeleteByUserId(ctx context.Context, userId string) error {
	err := svc.ObjectSvc.DeleteByObjectTypeAndId(ctx, objecttype.ObjectTypeUser, userId)
	if err != nil {
		return err
	}

	return nil
}
