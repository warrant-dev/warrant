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

	"github.com/google/uuid"
	"github.com/pkg/errors"
	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypeUser = "user"

type UserService struct {
	service.BaseService
	Repository UserRepository
	EventSvc   event.Service
	ObjectSvc  *object.ObjectService
}

func NewService(env service.Env, repository UserRepository, eventSvc event.Service, objectSvc *object.ObjectService) *UserService {
	return &UserService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
		EventSvc:    eventSvc,
		ObjectSvc:   objectSvc,
	}
}

func (svc UserService) Create(ctx context.Context, userSpec UserSpec) (*UserSpec, error) {
	if userSpec.UserId == "" {
		// generate an id for the user if one isn't provided
		generatedUUID, err := uuid.NewRandom()
		if err != nil {
			return nil, errors.New("unable to generate random UUID for user")
		}
		userSpec.UserId = generatedUUID.String()
	}

	var newUser Model
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		createdObject, err := svc.ObjectSvc.Create(txCtx, *userSpec.ToObjectSpec())
		if err != nil {
			switch err.(type) {
			case *service.DuplicateRecordError:
				return service.NewDuplicateRecordError("User", userSpec.UserId, "A user with the given userId already exists")
			default:
				return err
			}
		}

		_, err = svc.Repository.GetByUserId(txCtx, userSpec.UserId)
		if err == nil {
			return service.NewDuplicateRecordError("User", userSpec.UserId, "A user with the given userId already exists")
		}

		newUserId, err := svc.Repository.Create(txCtx, userSpec.ToUser(createdObject.ID))
		if err != nil {
			return err
		}

		newUser, err = svc.Repository.GetById(txCtx, newUserId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceCreated(txCtx, ResourceTypeUser, newUser.GetUserId(), newUser.ToUserSpec())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newUser.ToUserSpec(), nil
}

func (svc UserService) GetByUserId(ctx context.Context, userId string) (*UserSpec, error) {
	user, err := svc.Repository.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	return user.ToUserSpec(), nil
}

func (svc UserService) List(ctx context.Context, listParams service.ListParams) ([]UserSpec, error) {
	userSpecs := make([]UserSpec, 0)

	users, err := svc.Repository.List(ctx, listParams)
	if err != nil {
		return userSpecs, err
	}

	for _, user := range users {
		userSpecs = append(userSpecs, *user.ToUserSpec())
	}

	return userSpecs, nil
}

func (svc UserService) UpdateByUserId(ctx context.Context, userId string, userSpec UpdateUserSpec) (*UserSpec, error) {
	currentUser, err := svc.Repository.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	currentUser.SetEmail(userSpec.Email)
	err = svc.Repository.UpdateByUserId(ctx, userId, currentUser)
	if err != nil {
		return nil, err
	}

	updatedUser, err := svc.Repository.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	updatedUserSpec := updatedUser.ToUserSpec()
	err = svc.EventSvc.TrackResourceUpdated(ctx, ResourceTypeUser, updatedUser.GetUserId(), updatedUserSpec)
	if err != nil {
		return nil, err
	}

	return updatedUserSpec, nil
}

func (svc UserService) DeleteByUserId(ctx context.Context, userId string) (*wookie.Token, error) {
	var newWookie *wookie.Token
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		err := svc.Repository.DeleteByUserId(txCtx, userId)
		if err != nil {
			return err
		}

		newWookie, err = svc.ObjectSvc.DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypeUser, userId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceDeleted(txCtx, ResourceTypeUser, userId, nil)
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
