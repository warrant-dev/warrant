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

	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	object "github.com/warrant-dev/warrant/pkg/object"
	"github.com/warrant-dev/warrant/pkg/service"
)

type FeatureService struct {
	service.BaseService
	objectSvc object.Service
}

func NewService(env service.Env, objectSvc object.Service) *FeatureService {
	return &FeatureService{
		BaseService: service.NewBaseService(env),
		objectSvc:   objectSvc,
	}
}

func (svc FeatureService) Create(ctx context.Context, featureSpec FeatureSpec) (*FeatureSpec, error) {
	var createdFeatureSpec *FeatureSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		objectSpec, err := featureSpec.ToCreateObjectSpec()
		if err != nil {
			return err
		}

		createdObjectSpec, err := svc.objectSvc.Create(txCtx, *objectSpec)
		if err != nil {
			return err
		}

		createdFeatureSpec, err = NewFeatureSpecFromObjectSpec(createdObjectSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdFeatureSpec, nil
}

func (svc FeatureService) GetByFeatureId(ctx context.Context, featureId string) (*FeatureSpec, error) {
	objectSpec, err := svc.objectSvc.GetByObjectTypeAndId(ctx, objecttype.ObjectTypeFeature, featureId)
	if err != nil {
		return nil, err
	}

	return NewFeatureSpecFromObjectSpec(objectSpec)
}

func (svc FeatureService) List(ctx context.Context, listParams service.ListParams) ([]FeatureSpec, error) {
	featureSpecs := make([]FeatureSpec, 0)
	objectSpecs, _, _, err := svc.objectSvc.List(ctx, &object.FilterOptions{ObjectType: objecttype.ObjectTypeFeature}, listParams)
	if err != nil {
		return featureSpecs, err
	}

	for i := range objectSpecs {
		featureSpec, err := NewFeatureSpecFromObjectSpec(&objectSpecs[i])
		if err != nil {
			return featureSpecs, err
		}

		featureSpecs = append(featureSpecs, *featureSpec)
	}

	return featureSpecs, nil
}

func (svc FeatureService) UpdateByFeatureId(ctx context.Context, featureId string, featureSpec UpdateFeatureSpec) (*FeatureSpec, error) {
	var updatedFeatureSpec *FeatureSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		updatedObjectSpec, err := svc.objectSvc.UpdateByObjectTypeAndId(txCtx, objecttype.ObjectTypeFeature, featureId, *featureSpec.ToUpdateObjectSpec())
		if err != nil {
			return err
		}

		updatedFeatureSpec, err = NewFeatureSpecFromObjectSpec(updatedObjectSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return updatedFeatureSpec, nil
}

func (svc FeatureService) DeleteByFeatureId(ctx context.Context, featureId string) error {
	_, err := svc.objectSvc.DeleteByObjectTypeAndId(ctx, objecttype.ObjectTypeFeature, featureId)
	if err != nil {
		return err
	}

	return nil
}
