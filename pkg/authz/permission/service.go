package authz

import (
	"context"

	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypePermission = "permission"

type PermissionService struct {
	service.BaseService
	repo      PermissionRepository
	eventSvc  event.EventService
	objectSvc object.ObjectService
}

func NewService(env service.Env, repo PermissionRepository, eventSvc event.EventService, objectSvc object.ObjectService) PermissionService {
	return PermissionService{
		BaseService: service.NewBaseService(env),
		repo:        repo,
		eventSvc:    eventSvc,
		objectSvc:   objectSvc,
	}
}

func (svc PermissionService) Create(ctx context.Context, permissionSpec PermissionSpec) (*PermissionSpec, error) {
	var newPermission Model
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		createdObject, err := svc.objectSvc.Create(txCtx, *permissionSpec.ToObjectSpec())
		if err != nil {
			return err
		}

		_, err = svc.repo.GetByPermissionId(txCtx, permissionSpec.PermissionId)
		if err == nil {
			return service.NewDuplicateRecordError("Permission", permissionSpec.PermissionId, "A permission with the given permissionId already exists")
		}

		newPermissionId, err := svc.repo.Create(txCtx, permissionSpec.ToPermission(createdObject.ID))
		if err != nil {
			return err
		}

		newPermission, err = svc.repo.GetById(txCtx, newPermissionId)
		if err != nil {
			return err
		}

		svc.eventSvc.TrackResourceCreated(txCtx, ResourceTypePermission, newPermission.GetPermissionId(), newPermission.ToPermissionSpec())
		return nil
	})

	if err != nil {
		return nil, err
	}

	return newPermission.ToPermissionSpec(), nil
}

func (svc PermissionService) GetByPermissionId(ctx context.Context, permissionId string) (*PermissionSpec, error) {
	permissionRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	permission, err := permissionRepository.GetByPermissionId(ctx, permissionId)
	if err != nil {
		return nil, err
	}

	return permission.ToPermissionSpec(), nil
}

func (svc PermissionService) List(ctx context.Context, listParams middleware.ListParams) ([]PermissionSpec, error) {
	permissionSpecs := make([]PermissionSpec, 0)
	permissionRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return permissionSpecs, err
	}

	permissions, err := permissionRepository.List(ctx, listParams)
	if err != nil {
		return permissionSpecs, nil
	}

	for _, permission := range permissions {
		permissionSpecs = append(permissionSpecs, *permission.ToPermissionSpec())
	}

	return permissionSpecs, nil
}

func (svc PermissionService) UpdateByPermissionId(ctx context.Context, permissionId string, permissionSpec UpdatePermissionSpec) (*PermissionSpec, error) {
	permissionRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	currentPermission, err := permissionRepository.GetByPermissionId(ctx, permissionId)
	if err != nil {
		return nil, err
	}

	currentPermission.SetName(permissionSpec.Name)
	currentPermission.SetDescription(permissionSpec.Description)
	err = permissionRepository.UpdateByPermissionId(ctx, permissionId, currentPermission)
	if err != nil {
		return nil, err
	}

	updatedPermission, err := permissionRepository.GetByPermissionId(ctx, permissionId)
	if err != nil {
		return nil, err
	}

	updatedPermissionSpec := updatedPermission.ToPermissionSpec()
	svc.eventSvc.TrackResourceUpdated(ctx, ResourceTypePermission, updatedPermission.GetPermissionId(), updatedPermissionSpec)
	return updatedPermissionSpec, nil
}

func (svc PermissionService) DeleteByPermissionId(ctx context.Context, permissionId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		permissionRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		err = permissionRepository.DeleteByPermissionId(txCtx, permissionId)
		if err != nil {
			return err
		}

		err = svc.objectSvc.DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypePermission, permissionId)
		if err != nil {
			return err
		}

		svc.eventSvc.TrackResourceDeleted(txCtx, ResourceTypePermission, permissionId, nil)
		return nil
	})

	return err
}
