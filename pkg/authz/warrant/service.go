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
	"errors"

	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/event"
	object "github.com/warrant-dev/warrant/pkg/object"
	"github.com/warrant-dev/warrant/pkg/service"
	"github.com/warrant-dev/warrant/pkg/wookie"
)

type Service interface {
	Create(ctx context.Context, warrantSpec WarrantSpec) (*WarrantSpec, *wookie.Token, error)
	List(ctx context.Context, filterParams *FilterParams, listParams service.ListParams) ([]WarrantSpec, error)
	Delete(ctx context.Context, warrantSpec WarrantSpec) (*wookie.Token, error)
}

type WarrantService struct {
	service.BaseService
	Repository    WarrantRepository
	EventSvc      event.Service
	ObjectTypeSvc objecttype.Service
	ObjectSvc     object.Service
}

func NewService(env service.Env, repository WarrantRepository, eventSvc event.Service, objectTypeSvc objecttype.Service, objectSvc object.Service) *WarrantService {
	return &WarrantService{
		BaseService:   service.NewBaseService(env),
		Repository:    repository,
		EventSvc:      eventSvc,
		ObjectTypeSvc: objectTypeSvc,
		ObjectSvc:     objectSvc,
	}
}

func (svc WarrantService) Create(ctx context.Context, warrantSpec WarrantSpec) (*WarrantSpec, *wookie.Token, error) {
	var createdWarrant Model
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		// Check that objectType is valid
		objectTypeDef, err := svc.ObjectTypeSvc.GetByTypeId(txCtx, warrantSpec.ObjectType)
		if err != nil {
			return service.NewInvalidParameterError("objectType", "The given object type does not exist.")
		}

		// Check that relation is valid for objectType
		if _, exists := objectTypeDef.Relations[warrantSpec.Relation]; !exists {
			return service.NewInvalidParameterError("relation", "An object type with the given relation does not exist.")
		}

		// Check that warrant does not already exist
		_, err = svc.Repository.Get(txCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Policy.Hash())
		if err == nil {
			return service.NewDuplicateRecordError("Warrant", warrantSpec, "A warrant with the given objectType, objectId, relation, subject, and policy already exists")
		}

		// Unless objectId is wildcard, create referenced object if it does not already exist
		if warrantSpec.ObjectId != "*" {
			objectSpec, err := svc.ObjectSvc.GetByObjectTypeAndId(txCtx, warrantSpec.ObjectType, warrantSpec.ObjectId)
			if err != nil {
				var recordNotFoundError *service.RecordNotFoundError
				if !errors.As(err, &recordNotFoundError) {
					return err
				}
			}

			if objectSpec == nil {
				_, err = svc.ObjectSvc.Create(txCtx, object.CreateObjectSpec{
					ObjectType: warrantSpec.ObjectType,
					ObjectId:   warrantSpec.ObjectId,
				})
				if err != nil {
					var duplicateRecordError *service.DuplicateRecordError
					if !errors.As(err, &duplicateRecordError) {
						return err
					}
				}
			}
		}

		// Create referenced subject if it does not already exist
		objectSpec, err := svc.ObjectSvc.GetByObjectTypeAndId(txCtx, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId)
		if err != nil {
			var recordNotFoundError *service.RecordNotFoundError
			if !errors.As(err, &recordNotFoundError) {
				return err
			}
		}

		if objectSpec == nil {
			_, err = svc.ObjectSvc.Create(txCtx, object.CreateObjectSpec{
				ObjectType: warrantSpec.Subject.ObjectType,
				ObjectId:   warrantSpec.Subject.ObjectId,
			})
			if err != nil {
				var duplicateRecordError *service.DuplicateRecordError
				if !errors.As(err, &duplicateRecordError) {
					return err
				}
			}
		}

		warrant, err := warrantSpec.ToWarrant()
		if err != nil {
			return err
		}

		createdWarrantId, err := svc.Repository.Create(txCtx, warrant)
		if err != nil {
			return err
		}

		createdWarrant, err = svc.Repository.GetByID(txCtx, createdWarrantId)
		if err != nil {
			return err
		}

		var eventMeta map[string]interface{}
		if createdWarrant.GetPolicy() != "" {
			eventMeta = make(map[string]interface{})
			eventMeta["policy"] = createdWarrant.GetPolicy()
		}

		err = svc.EventSvc.TrackAccessGrantedEvent(txCtx, createdWarrant.GetObjectType(), createdWarrant.GetObjectId(), createdWarrant.GetRelation(), createdWarrant.GetSubjectType(), createdWarrant.GetSubjectId(), createdWarrant.GetSubjectRelation(), eventMeta)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return createdWarrant.ToWarrantSpec(), nil, nil
}

func (svc WarrantService) List(ctx context.Context, filterParams *FilterParams, listParams service.ListParams) ([]WarrantSpec, error) {
	warrantSpecs := make([]WarrantSpec, 0)
	warrants, err := svc.Repository.List(ctx, filterParams, listParams)
	if err != nil {
		return warrantSpecs, err
	}

	for _, warrant := range warrants {
		warrantSpecs = append(warrantSpecs, *warrant.ToWarrantSpec())
	}

	return warrantSpecs, nil
}

func (svc WarrantService) Delete(ctx context.Context, warrantSpec WarrantSpec) (*wookie.Token, error) {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		warrantToDelete, err := warrantSpec.ToWarrant()
		if err != nil {
			return err
		}

		_, err = svc.Repository.Get(txCtx, warrantToDelete.GetObjectType(), warrantToDelete.GetObjectId(), warrantToDelete.GetRelation(), warrantToDelete.GetSubjectType(), warrantToDelete.GetSubjectId(), warrantToDelete.GetSubjectRelation(), warrantToDelete.GetPolicyHash())
		if err != nil {
			return err
		}

		err = svc.Repository.Delete(txCtx, warrantToDelete.GetObjectType(), warrantToDelete.GetObjectId(), warrantToDelete.GetRelation(), warrantToDelete.GetSubjectType(), warrantToDelete.GetSubjectId(), warrantToDelete.GetSubjectRelation(), warrantToDelete.GetPolicyHash())
		if err != nil {
			return err
		}

		var eventMeta map[string]interface{}
		if warrantSpec.Policy != "" {
			eventMeta = make(map[string]interface{})
			eventMeta["policy"] = warrantSpec.Policy
		}

		err = svc.EventSvc.TrackAccessRevokedEvent(txCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
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
