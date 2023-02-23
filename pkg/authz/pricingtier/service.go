package authz

import (
	"context"

	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

type PricingTierService struct {
	service.BaseService
}

func NewService(env service.Env) PricingTierService {
	return PricingTierService{
		BaseService: service.NewBaseService(env),
	}
}

func (svc PricingTierService) Create(ctx context.Context, pricingTierSpec PricingTierSpec) (*PricingTierSpec, error) {
	var newPricingTier *PricingTier
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		pricingTierRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		createdObject, err := object.NewService(svc.Env()).Create(txCtx, *pricingTierSpec.ToObjectSpec())
		if err != nil {
			return err
		}

		_, err = pricingTierRepository.GetByPricingTierId(txCtx, pricingTierSpec.PricingTierId)
		if err == nil {
			return service.NewDuplicateRecordError("PricingTier", pricingTierSpec.PricingTierId, "A pricing tier with the given pricingTierId already exists")
		}

		newPricingTierId, err := pricingTierRepository.Create(txCtx, *pricingTierSpec.ToPricingTier(createdObject.ID))
		if err != nil {
			return err
		}

		newPricingTier, err = pricingTierRepository.GetById(txCtx, newPricingTierId)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newPricingTier.ToPricingTierSpec(), nil
}

func (svc PricingTierService) GetByPricingTierId(ctx context.Context, pricingTierId string) (*PricingTierSpec, error) {
	pricingTierRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	pricingTier, err := pricingTierRepository.GetByPricingTierId(ctx, pricingTierId)
	if err != nil {
		return nil, err
	}

	return pricingTier.ToPricingTierSpec(), nil
}

func (svc PricingTierService) List(ctx context.Context, listParams middleware.ListParams) ([]PricingTierSpec, error) {
	pricingTierSpecs := make([]PricingTierSpec, 0)
	pricingTierRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return pricingTierSpecs, err
	}

	pricingTiers, err := pricingTierRepository.List(ctx, listParams)
	if err != nil {
		return pricingTierSpecs, nil
	}

	for _, pricingTier := range pricingTiers {
		pricingTierSpecs = append(pricingTierSpecs, *pricingTier.ToPricingTierSpec())
	}

	return pricingTierSpecs, nil
}

func (svc PricingTierService) UpdateByPricingTierId(ctx context.Context, pricingTierId string, pricingTierSpec UpdatePricingTierSpec) (*PricingTierSpec, error) {
	pricingTierRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	currentPricingTier, err := pricingTierRepository.GetByPricingTierId(ctx, pricingTierId)
	if err != nil {
		return nil, err
	}

	currentPricingTier.Name = pricingTierSpec.Name
	currentPricingTier.Description = pricingTierSpec.Description
	err = pricingTierRepository.UpdateByPricingTierId(ctx, pricingTierId, *currentPricingTier)
	if err != nil {
		return nil, err
	}

	updatedPricingTier, err := pricingTierRepository.GetByPricingTierId(ctx, pricingTierId)
	if err != nil {
		return nil, err
	}

	return updatedPricingTier.ToPricingTierSpec(), nil
}

func (svc PricingTierService) DeleteByPricingTierId(ctx context.Context, pricingTierId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		pricingTierRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		err = pricingTierRepository.DeleteByPricingTierId(txCtx, pricingTierId)
		if err != nil {
			return err
		}

		err = object.NewService(svc.Env()).DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypePricingTier, pricingTierId)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
