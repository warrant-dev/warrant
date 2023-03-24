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
}

func NewService(env service.Env) PermissionService {
	return PermissionService{
		BaseService: service.NewBaseService(env),
	}
}

func (svc PermissionService) Create(ctx context.Context, permissionSpec PermissionSpec) (*PermissionSpec, error) {
	var newPermission *Permission
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		permissionRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		createdObject, err := object.NewService(svc.Env()).Create(txCtx, *permissionSpec.ToObjectSpec())
		if err != nil {
			return err
		}

		_, err = permissionRepository.GetByPermissionId(txCtx, permissionSpec.PermissionId)
		if err == nil {
			return service.NewDuplicateRecordError("Permission", permissionSpec.PermissionId, "A permission with the given permissionId already exists")
		}

		newPermissionId, err := permissionRepository.Create(txCtx, *permissionSpec.ToPermission(createdObject.ID))
		if err != nil {
			return err
		}

		newPermission, err = permissionRepository.GetById(txCtx, newPermissionId)
		if err != nil {
			return err
		}

		event.NewService(svc.Env()).TrackResourceCreated(txCtx, ResourceTypePermission, newPermission.PermissionId, newPermission.ToPermissionSpec())
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

	currentPermission.Name = permissionSpec.Name
	currentPermission.Description = permissionSpec.Description
	err = permissionRepository.UpdateByPermissionId(ctx, permissionId, *currentPermission)
	if err != nil {
		return nil, err
	}

	updatedPermission, err := permissionRepository.GetByPermissionId(ctx, permissionId)
	if err != nil {
		return nil, err
	}

	updatedPermissionSpec := updatedPermission.ToPermissionSpec()
	event.NewService(svc.Env()).TrackResourceUpdated(ctx, ResourceTypePermission, updatedPermission.PermissionId, updatedPermissionSpec)
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

		err = object.NewService(svc.Env()).DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypePermission, permissionId)
		if err != nil {
			return err
		}

		event.NewService(svc.Env()).TrackResourceDeleted(txCtx, ResourceTypePermission, permissionId, nil)
		return nil
	})

	return err
}
