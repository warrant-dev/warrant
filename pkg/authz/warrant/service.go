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
	"errors"

	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/object"
	"github.com/warrant-dev/warrant/pkg/service"
	"github.com/warrant-dev/warrant/pkg/wookie"
)

type Service interface {
	Create(ctx context.Context, spec CreateWarrantSpec) (*WarrantSpec, *wookie.Token, error)
	List(ctx context.Context, filterParams FilterParams, listParams service.ListParams) ([]WarrantSpec, *service.Cursor, *service.Cursor, error)
	Delete(ctx context.Context, spec DeleteWarrantSpec) (*wookie.Token, error)
}

type WarrantService struct {
	service.BaseService
	repository    WarrantRepository
	objectTypeSvc objecttype.Service
	objectSvc     object.Service
}

func NewService(env service.Env, repository WarrantRepository, objectTypeSvc objecttype.Service, objectSvc object.Service) *WarrantService {
	return &WarrantService{
		BaseService:   service.NewBaseService(env),
		repository:    repository,
		objectTypeSvc: objectTypeSvc,
		objectSvc:     objectSvc,
	}
}

func (svc WarrantService) Create(ctx context.Context, spec CreateWarrantSpec) (*WarrantSpec, *wookie.Token, error) {
	var createdWarrant Model
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		// Check that objectType exists
		objectTypeDef, err := svc.objectTypeSvc.GetByTypeId(txCtx, spec.ObjectType)
		if err != nil {
			var recordNotFoundError *service.RecordNotFoundError
			if errors.As(err, &recordNotFoundError) {
				return service.NewInvalidParameterError("objectType", "the object type does not exist.")
			}

			return err
		}

		// Check that relation is valid for objectType
		if _, exists := objectTypeDef.Relations[spec.Relation]; !exists {
			return service.NewInvalidParameterError("relation", "the relation does not exist on the specified object type.")
		}

		// Unless objectId is wildcard, create referenced object if it does not already exist
		if spec.ObjectId != Wildcard {
			objectSpec, err := svc.objectSvc.GetByObjectTypeAndId(txCtx, spec.ObjectType, spec.ObjectId)
			if err != nil {
				var recordNotFoundError *service.RecordNotFoundError
				if !errors.As(err, &recordNotFoundError) {
					return err
				}
			}

			if objectSpec == nil {
				_, err = svc.objectSvc.Create(txCtx, object.CreateObjectSpec{
					ObjectType: spec.ObjectType,
					ObjectId:   spec.ObjectId,
				})
				if err != nil {
					var duplicateRecordError *service.DuplicateRecordError
					if !errors.As(err, &duplicateRecordError) {
						return err
					}
				}
			}
		}

		// Unless subject objectId is wildcard, create referenced subject if it does not already exist
		if spec.Subject.ObjectId != Wildcard {
			objectSpec, err := svc.objectSvc.GetByObjectTypeAndId(txCtx, spec.Subject.ObjectType, spec.Subject.ObjectId)
			if err != nil {
				var recordNotFoundError *service.RecordNotFoundError
				if !errors.As(err, &recordNotFoundError) {
					return err
				}
			}

			if objectSpec == nil {
				_, err = svc.objectSvc.Create(txCtx, object.CreateObjectSpec{
					ObjectType: spec.Subject.ObjectType,
					ObjectId:   spec.Subject.ObjectId,
				})
				if err != nil {
					var duplicateRecordError *service.DuplicateRecordError
					if !errors.As(err, &duplicateRecordError) {
						return err
					}
				}
			}
		}

		warrant, err := spec.ToWarrant()
		if err != nil {
			return err
		}

		createdWarrantId, err := svc.repository.Create(txCtx, warrant)
		if err != nil {
			return err
		}

		createdWarrant, err = svc.repository.GetByID(txCtx, createdWarrantId)
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

func (svc WarrantService) List(ctx context.Context, filterParams FilterParams, listParams service.ListParams) ([]WarrantSpec, *service.Cursor, *service.Cursor, error) {
	warrantSpecs := make([]WarrantSpec, 0)
	warrants, prevCursor, nextCursor, err := svc.repository.List(ctx, filterParams, listParams)
	if err != nil {
		return warrantSpecs, prevCursor, nextCursor, err
	}

	for _, warrant := range warrants {
		warrantSpecs = append(warrantSpecs, *warrant.ToWarrantSpec())
	}

	return warrantSpecs, prevCursor, nextCursor, nil
}

func (svc WarrantService) Delete(ctx context.Context, spec DeleteWarrantSpec) (*wookie.Token, error) {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		warrantToDelete, err := spec.ToWarrant()
		if err != nil {
			return err
		}

		err = svc.repository.Delete(txCtx, warrantToDelete.GetObjectType(), warrantToDelete.GetObjectId(), warrantToDelete.GetRelation(), warrantToDelete.GetSubjectType(), warrantToDelete.GetSubjectId(), warrantToDelete.GetSubjectRelation(), warrantToDelete.GetPolicyHash())
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
