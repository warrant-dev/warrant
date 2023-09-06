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

const ResourceTypePricingTier = "pricing-tier"

type PricingTierService struct {
	service.BaseService
	EventSvc  event.Service
	ObjectSvc *object.ObjectService
}

func NewService(env service.Env, eventSvc event.Service, objectSvc *object.ObjectService) *PricingTierService {
	return &PricingTierService{
		BaseService: service.NewBaseService(env),
		EventSvc:    eventSvc,
		ObjectSvc:   objectSvc,
	}
}

func (svc PricingTierService) Create(ctx context.Context, pricingTierSpec PricingTierSpec) (*PricingTierSpec, error) {
	var createdPricingTierSpec *PricingTierSpec
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		objectSpec, err := pricingTierSpec.ToCreateObjectSpec()
		if err != nil {
			return err
		}

		createdObjectSpec, err := svc.ObjectSvc.Create(txCtx, *objectSpec)
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
	objectSpec, err := svc.ObjectSvc.GetByObjectTypeAndId(ctx, objecttype.ObjectTypePricingTier, pricingTierId)
	if err != nil {
		return nil, err
	}

	return NewPricingTierSpecFromObjectSpec(objectSpec)
}

func (svc PricingTierService) List(ctx context.Context, listParams service.ListParams) ([]PricingTierSpec, error) {
	pricingTierSpecs := make([]PricingTierSpec, 0)
	objectSpecs, err := svc.ObjectSvc.List(ctx, &object.FilterOptions{ObjectType: objecttype.ObjectTypePricingTier}, listParams)
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
		updatedObjectSpec, err := svc.ObjectSvc.UpdateByObjectTypeAndId(txCtx, objecttype.ObjectTypePricingTier, pricingTierId, *pricingTierSpec.ToUpdateObjectSpec())
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
	err := svc.ObjectSvc.DeleteByObjectTypeAndId(ctx, objecttype.ObjectTypePricingTier, pricingTierId)
	if err != nil {
		return err
	}

	return nil
}
