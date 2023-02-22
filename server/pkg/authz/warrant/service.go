package authz

import (
	"context"

	wntContext "github.com/warrant-dev/warrant/server/pkg/authz/context"
	objecttype "github.com/warrant-dev/warrant/server/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

type WarrantService struct {
	service.BaseService
	objectTypeMap map[string]*objecttype.ObjectTypeSpec
}

func NewService(env service.Env) WarrantService {
	return WarrantService{
		BaseService:   service.NewBaseService(env),
		objectTypeMap: make(map[string]*objecttype.ObjectTypeSpec),
	}
}

func (svc WarrantService) Create(ctx context.Context, warrantSpec WarrantSpec) (*WarrantSpec, error) {
	// Check that objectType is valid
	objectTypeDef, err := objecttype.NewService(svc.Env()).GetByTypeId(ctx, warrantSpec.ObjectType)
	if err != nil {
		return nil, service.NewInvalidParameterError("objectType", "The given object type does not exist.")
	}

	// Check that relation is valid for objectType
	_, exists := objectTypeDef.Relations[warrantSpec.Relation]
	if !exists {
		return nil, service.NewInvalidParameterError("relation", "An object type with the given relation does not exist.")
	}

	warrantRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	// Check that warrant does not already exist
	_, err = warrantRepository.Get(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Context.String())
	if err == nil {
		return nil, service.NewDuplicateRecordError("Warrant", warrantSpec, "A warrant with the given objectType, objectId, relation, subject, and context already exists")
	}

	var createdWarrant *Warrant
	err = svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		createdWarrantId, err := warrantRepository.Create(txCtx, *warrantSpec.ToWarrant())
		if err != nil {
			return err
		}

		createdWarrant, err = warrantRepository.GetByID(txCtx, createdWarrantId)
		if err != nil {
			return err
		}

		contexts := warrantSpec.Context.ToSlice(createdWarrantId)
		for _, contextObject := range contexts {
			if !contextObject.IsValid() {
				return service.NewInvalidParameterError("context", "The context name and value must only contain alphanumeric characters, '-', and/or '_'")
			}
		}
		if len(contexts) > 0 {
			contextRepository, err := wntContext.NewRepository(svc.Env().DB())
			if err != nil {
				return err
			}

			createdWarrant.Context, err = contextRepository.CreateAll(txCtx, contexts)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdWarrant.ToWarrantSpec(), nil
}

func (svc WarrantService) Get(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, wntCtx wntContext.ContextSetSpec) (*WarrantSpec, error) {
	warrantRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	warrant, err := warrantRepository.Get(ctx, objectType, objectId, relation, subjectType, subjectId, subjectRelation, wntCtx.ToHash())
	if err != nil {
		return nil, err
	}

	contextRepository, err := wntContext.NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	warrant.Context, err = contextRepository.ListByWarrantId(ctx, []int64{warrant.ID})
	if err != nil {
		return nil, err
	}

	return warrant.ToWarrantSpec(), nil
}

func (svc WarrantService) List(ctx context.Context, filterOptions *FilterOptions, listParams middleware.ListParams) ([]*WarrantSpec, error) {
	warrantSpecs := make([]*WarrantSpec, 0)
	warrantRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	warrants, err := warrantRepository.List(ctx, filterOptions, listParams)
	if err != nil {
		return nil, err
	}

	warrantMap := make(map[int64]*Warrant)
	warrantIds := make([]int64, 0)
	for i, warrant := range warrants {
		warrantIds = append(warrantIds, warrant.ID)
		warrantMap[warrant.ID] = &warrants[i]
	}

	contextRepository, err := wntContext.NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	contexts, err := contextRepository.ListByWarrantId(ctx, warrantIds)
	if err != nil {
		return nil, err
	}

	for _, context := range contexts {
		warrantMap[context.WarrantId].Context = append(warrantMap[context.WarrantId].Context, context)
	}

	for _, warrant := range warrants {
		warrantSpecs = append(warrantSpecs, warrant.ToWarrantSpec())
	}

	return warrantSpecs, nil
}

func (svc WarrantService) Delete(ctx context.Context, warrantSpec WarrantSpec) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		warrantRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		warrant, err := warrantRepository.Get(txCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Context.ToHash())
		if err != nil {
			return err
		}

		contextRepository, err := wntContext.NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		err = contextRepository.DeleteAllByWarrantId(txCtx, warrant.ID)
		if err != nil {
			return err
		}

		err = warrantRepository.DeleteById(txCtx, warrant.ID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (svc WarrantService) DeleteRelatedWarrants(ctx context.Context, objectType string, objectId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		warrantRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		err = warrantRepository.DeleteAllByObject(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		err = warrantRepository.DeleteAllBySubject(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
