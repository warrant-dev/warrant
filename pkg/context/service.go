package context

import (
	"context"

	"github.com/warrant-dev/warrant/pkg/service"
)

type ContextService struct {
	service.BaseService
	Repository ContextRepository
}

func NewService(env service.Env, repository ContextRepository) ContextService {
	return ContextService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
	}
}

func (svc ContextService) CreateAll(ctx context.Context, warrantId int64, spec ContextSetSpec) (ContextSetSpec, error) {
	contexts, err := svc.Repository.CreateAll(ctx, spec.ToSlice(warrantId))
	if err != nil {
		return nil, err
	}

	return NewContextSetSpecFromSlice(contexts), nil
}

func (svc ContextService) ListByWarrantId(ctx context.Context, warrantIds []int64) (map[int64]ContextSetSpec, error) {
	contexts, err := svc.Repository.ListByWarrantId(ctx, warrantIds)
	if err != nil {
		return nil, err
	}

	contextMap := make(map[int64][]Model)
	for _, c := range contexts {
		contextMap[c.GetWarrantId()] = append(contextMap[c.GetWarrantId()], c)
	}

	contextSpecMap := make(map[int64]ContextSetSpec)
	for warrantId, cs := range contextMap {
		contextSpecMap[warrantId] = NewContextSetSpecFromSlice(cs)
	}

	return contextSpecMap, nil
}

func (svc ContextService) DeleteAllByWarrantId(ctx context.Context, warrantId int64) error {
	return svc.Repository.DeleteAllByWarrantId(ctx, warrantId)
}
