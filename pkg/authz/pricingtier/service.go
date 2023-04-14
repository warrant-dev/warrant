package authz

import (
	"context"

	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypePricingTier = "pricing-tier"

type PricingTierService struct {
	service.BaseService
	repo      PricingTierRepository
	eventSvc  event.EventService
	objectSvc object.ObjectService
}

func NewService(env service.Env, repo PricingTierRepository, eventSvc event.EventService, objectSvc object.ObjectService) PricingTierService {
	return PricingTierService{
		BaseService: service.NewBaseService(env),
		repo:        repo,
		eventSvc:    eventSvc,
		objectSvc:   objectSvc,
	}
}

func (svc PricingTierService) Create(ctx context.Context, pricingTierSpec PricingTierSpec) (*PricingTierSpec, error) {
	var newPricingTier Model
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		createdObject, err := svc.objectSvc.Create(txCtx, *pricingTierSpec.ToObjectSpec())
		if err != nil {
			return err
		}

		_, err = svc.repo.GetByPricingTierId(txCtx, pricingTierSpec.PricingTierId)
		if err == nil {
			return service.NewDuplicateRecordError("PricingTier", pricingTierSpec.PricingTierId, "A pricing tier with the given pricingTierId already exists")
		}

		newPricingTierId, err := svc.repo.Create(txCtx, pricingTierSpec.ToPricingTier(createdObject.ID))
		if err != nil {
			return err
		}

		newPricingTier, err = svc.repo.GetById(txCtx, newPricingTierId)
		if err != nil {
			return err
		}

		err = svc.eventSvc.TrackResourceCreated(ctx, ResourceTypePricingTier, newPricingTier.GetPricingTierId(), newPricingTier.ToPricingTierSpec())
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
	pricingTier, err := svc.repo.GetByPricingTierId(ctx, pricingTierId)
	if err != nil {
		return nil, err
	}

	return pricingTier.ToPricingTierSpec(), nil
}

func (svc PricingTierService) List(ctx context.Context, listParams middleware.ListParams) ([]PricingTierSpec, error) {
	pricingTierSpecs := make([]PricingTierSpec, 0)

	pricingTiers, err := svc.repo.List(ctx, listParams)
	if err != nil {
		return pricingTierSpecs, nil
	}

	for _, pricingTier := range pricingTiers {
		pricingTierSpecs = append(pricingTierSpecs, *pricingTier.ToPricingTierSpec())
	}

	return pricingTierSpecs, nil
}

func (svc PricingTierService) UpdateByPricingTierId(ctx context.Context, pricingTierId string, pricingTierSpec UpdatePricingTierSpec) (*PricingTierSpec, error) {
	currentPricingTier, err := svc.repo.GetByPricingTierId(ctx, pricingTierId)
	if err != nil {
		return nil, err
	}

	currentPricingTier.SetName(pricingTierSpec.Name)
	currentPricingTier.SetDescription(pricingTierSpec.Description)
	err = svc.repo.UpdateByPricingTierId(ctx, pricingTierId, currentPricingTier)
	if err != nil {
		return nil, err
	}

	updatedPricingTier, err := svc.repo.GetByPricingTierId(ctx, pricingTierId)
	if err != nil {
		return nil, err
	}

	updatedPricingTierSpec := updatedPricingTier.ToPricingTierSpec()
	err = svc.eventSvc.TrackResourceUpdated(ctx, ResourceTypePricingTier, updatedPricingTier.GetPricingTierId(), updatedPricingTierSpec)
	if err != nil {
		return nil, err
	}

	return updatedPricingTierSpec, nil
}

func (svc PricingTierService) DeleteByPricingTierId(ctx context.Context, pricingTierId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		err := svc.repo.DeleteByPricingTierId(txCtx, pricingTierId)
		if err != nil {
			return err
		}

		err = svc.objectSvc.DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypePricingTier, pricingTierId)
		if err != nil {
			return err
		}

		err = svc.eventSvc.TrackResourceDeleted(ctx, ResourceTypePricingTier, pricingTierId, nil)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
