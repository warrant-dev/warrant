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

type PricingTierService struct {
	service.BaseService
	objectSvc object.Service
}

func NewService(env service.Env, objectSvc object.Service) *PricingTierService {
	return &PricingTierService{
		BaseService: service.NewBaseService(env),
		objectSvc:   objectSvc,
	}
}

func (svc PricingTierService) Create(ctx context.Context, pricingTierSpec PricingTierSpec) (*PricingTierSpec, error) {
	var createdPricingTierSpec *PricingTierSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		objectSpec, err := pricingTierSpec.ToCreateObjectSpec()
		if err != nil {
			return err
		}

		createdObjectSpec, err := svc.objectSvc.Create(txCtx, *objectSpec)
		if err != nil {
			return err
		}

		createdPricingTierSpec, err = NewPricingTierSpecFromObjectSpec(createdObjectSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdPricingTierSpec, nil
}

func (svc PricingTierService) GetByPricingTierId(ctx context.Context, pricingTierId string) (*PricingTierSpec, error) {
	objectSpec, err := svc.objectSvc.GetByObjectTypeAndId(ctx, objecttype.ObjectTypePricingTier, pricingTierId)
	if err != nil {
		return nil, err
	}

	return NewPricingTierSpecFromObjectSpec(objectSpec)
}

func (svc PricingTierService) List(ctx context.Context, listParams service.ListParams) ([]PricingTierSpec, error) {
	pricingTierSpecs := make([]PricingTierSpec, 0)
	objectSpecs, _, _, err := svc.objectSvc.List(ctx, &object.FilterOptions{ObjectType: objecttype.ObjectTypePricingTier}, listParams)
	if err != nil {
		return pricingTierSpecs, err
	}

	for i := range objectSpecs {
		pricingTierSpec, err := NewPricingTierSpecFromObjectSpec(&objectSpecs[i])
		if err != nil {
			return pricingTierSpecs, err
		}

		pricingTierSpecs = append(pricingTierSpecs, *pricingTierSpec)
	}

	return pricingTierSpecs, nil
}

func (svc PricingTierService) UpdateByPricingTierId(ctx context.Context, pricingTierId string, pricingTierSpec UpdatePricingTierSpec) (*PricingTierSpec, error) {
	var updatedPricingTierSpec *PricingTierSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		updatedObjectSpec, err := svc.objectSvc.UpdateByObjectTypeAndId(txCtx, objecttype.ObjectTypePricingTier, pricingTierId, *pricingTierSpec.ToUpdateObjectSpec())
		if err != nil {
			return err
		}

		updatedPricingTierSpec, err = NewPricingTierSpecFromObjectSpec(updatedObjectSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return updatedPricingTierSpec, nil
}

func (svc PricingTierService) DeleteByPricingTierId(ctx context.Context, pricingTierId string) error {
	_, err := svc.objectSvc.DeleteByObjectTypeAndId(ctx, objecttype.ObjectTypePricingTier, pricingTierId)
	if err != nil {
		return err
	}

	return nil
}
