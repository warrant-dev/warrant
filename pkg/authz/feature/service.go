package authz

import (
	"context"

	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

type FeatureService struct {
	service.BaseService
}

func NewService(env service.Env) FeatureService {
	return FeatureService{
		BaseService: service.NewBaseService(env),
	}
}

func (svc FeatureService) Create(ctx context.Context, featureSpec FeatureSpec) (*FeatureSpec, error) {
	var newFeature *Feature
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		featureRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		createdObject, err := object.NewService(svc.Env()).Create(txCtx, *featureSpec.ToObjectSpec())
		if err != nil {
			return err
		}

		_, err = featureRepository.GetByFeatureId(txCtx, featureSpec.FeatureId)
		if err == nil {
			return service.NewDuplicateRecordError("Feature", featureSpec.FeatureId, "A feature with the given featureId already exists")
		}

		newFeatureId, err := featureRepository.Create(txCtx, *featureSpec.ToFeature(createdObject.ID))
		if err != nil {
			return err
		}

		newFeature, err = featureRepository.GetById(txCtx, newFeatureId)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newFeature.ToFeatureSpec(), nil
}

func (svc FeatureService) GetByFeatureId(ctx context.Context, featureId string) (*FeatureSpec, error) {
	featureRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	feature, err := featureRepository.GetByFeatureId(ctx, featureId)
	if err != nil {
		return nil, err
	}

	return feature.ToFeatureSpec(), nil
}

func (svc FeatureService) List(ctx context.Context, listParams middleware.ListParams) ([]FeatureSpec, error) {
	featureSpecs := make([]FeatureSpec, 0)
	featureRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return featureSpecs, err
	}

	features, err := featureRepository.List(ctx, listParams)
	if err != nil {
		return featureSpecs, nil
	}

	for _, feature := range features {
		featureSpecs = append(featureSpecs, *feature.ToFeatureSpec())
	}

	return featureSpecs, nil
}

func (svc FeatureService) UpdateByFeatureId(ctx context.Context, featureId string, featureSpec UpdateFeatureSpec) (*FeatureSpec, error) {
	featureRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	currentFeature, err := featureRepository.GetByFeatureId(ctx, featureId)
	if err != nil {
		return nil, err
	}

	currentFeature.Name = featureSpec.Name
	currentFeature.Description = featureSpec.Description
	err = featureRepository.UpdateByFeatureId(ctx, featureId, *currentFeature)
	if err != nil {
		return nil, err
	}

	updatedFeature, err := featureRepository.GetByFeatureId(ctx, featureId)
	if err != nil {
		return nil, err
	}

	return updatedFeature.ToFeatureSpec(), nil
}

func (svc FeatureService) DeleteByFeatureId(ctx context.Context, featureId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		featureRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		err = featureRepository.DeleteByFeatureId(txCtx, featureId)
		if err != nil {
			return err
		}

		err = object.NewService(svc.Env()).DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypeFeature, featureId)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
