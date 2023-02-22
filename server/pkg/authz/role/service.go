package authz

import (
	"context"

	object "github.com/warrant-dev/warrant/server/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/server/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

type RoleService struct {
	service.BaseService
}

func NewService(env service.Env) RoleService {
	return RoleService{
		BaseService: service.NewBaseService(env),
	}
}

func (svc RoleService) Create(ctx context.Context, roleSpec RoleSpec) (*RoleSpec, error) {
	var newRole *Role
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		roleRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		createdObject, err := object.NewService(svc.Env()).Create(txCtx, *roleSpec.ToObjectSpec())
		if err != nil {
			return err
		}

		_, err = roleRepository.GetByRoleId(txCtx, roleSpec.RoleId)
		if err == nil {
			return service.NewDuplicateRecordError("Role", roleSpec.RoleId, "A role with the given roleId already exists")
		}

		newRoleId, err := roleRepository.Create(txCtx, *roleSpec.ToRole(createdObject.ID))
		if err != nil {
			return err
		}

		newRole, err = roleRepository.GetById(txCtx, newRoleId)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newRole.ToRoleSpec(), nil
}

func (svc RoleService) GetByRoleId(ctx context.Context, roleId string) (*RoleSpec, error) {
	roleRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	role, err := roleRepository.GetByRoleId(ctx, roleId)
	if err != nil {
		return nil, err
	}

	return role.ToRoleSpec(), nil
}

func (svc RoleService) List(ctx context.Context, listParams middleware.ListParams) ([]RoleSpec, error) {
	roleSpecs := make([]RoleSpec, 0)
	roleRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return roleSpecs, err
	}

	roles, err := roleRepository.List(ctx, listParams)
	if err != nil {
		return roleSpecs, nil
	}

	for _, role := range roles {
		roleSpecs = append(roleSpecs, *role.ToRoleSpec())
	}

	return roleSpecs, nil
}

func (svc RoleService) UpdateByRoleId(ctx context.Context, roleId string, roleSpec UpdateRoleSpec) (*RoleSpec, error) {
	roleRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	currentRole, err := roleRepository.GetByRoleId(ctx, roleId)
	if err != nil {
		return nil, err
	}

	currentRole.Name = roleSpec.Name
	currentRole.Description = roleSpec.Description
	err = roleRepository.UpdateByRoleId(ctx, roleId, *currentRole)
	if err != nil {
		return nil, err
	}

	updatedRole, err := roleRepository.GetByRoleId(ctx, roleId)
	if err != nil {
		return nil, err
	}

	return updatedRole.ToRoleSpec(), nil
}

func (svc RoleService) DeleteByRoleId(ctx context.Context, roleId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		roleRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		err = roleRepository.DeleteByRoleId(txCtx, roleId)
		if err != nil {
			return err
		}

		err = object.NewService(svc.Env()).DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypeRole, roleId)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
