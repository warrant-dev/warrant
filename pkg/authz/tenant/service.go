package tenant

import (
	"context"
	"regexp"

	"github.com/google/uuid"
	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

type TenantService struct {
	service.BaseService
}

func NewService(env service.Env) TenantService {
	return TenantService{
		BaseService: service.NewBaseService(env),
	}
}

func (svc TenantService) Create(ctx context.Context, tenantSpec TenantSpec) (*TenantSpec, error) {
	err := validateOrGenerateTenantIdInSpec(&tenantSpec)
	if err != nil {
		return nil, err
	}

	var newTenant *Tenant
	err = svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		tenantRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		createdObject, err := object.NewService(svc.Env()).Create(txCtx, *tenantSpec.ToObjectSpec())
		if err != nil {
			switch err.(type) {
			case *service.DuplicateRecordError:
				return service.NewDuplicateRecordError("Tenant", tenantSpec.TenantId, "A tenant with the given tenantId already exists")
			default:
				return err
			}
		}

		_, err = tenantRepository.GetByTenantId(txCtx, tenantSpec.TenantId)
		if err == nil {
			return service.NewDuplicateRecordError("Tenant", tenantSpec.TenantId, "A tenant with the given tenantId already exists")
		}

		newTenantId, err := tenantRepository.Create(txCtx, *tenantSpec.ToTenant(createdObject.ID))
		if err != nil {
			return err
		}

		newTenant, err = tenantRepository.GetById(txCtx, newTenantId)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newTenant.ToTenantSpec(), nil
}

func (svc TenantService) GetByTenantId(ctx context.Context, tenantId string) (*TenantSpec, error) {
	tenantRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	tenant, err := tenantRepository.GetByTenantId(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	return tenant.ToTenantSpec(), nil
}

func (svc TenantService) List(ctx context.Context, listParams middleware.ListParams) ([]TenantSpec, error) {
	tenantSpecs := make([]TenantSpec, 0)
	tenantRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return tenantSpecs, err
	}

	tenants, err := tenantRepository.List(ctx, listParams)
	if err != nil {
		return tenantSpecs, nil
	}

	for _, tenant := range tenants {
		tenantSpecs = append(tenantSpecs, *tenant.ToTenantSpec())
	}

	return tenantSpecs, nil
}

func (svc TenantService) UpdateByTenantId(ctx context.Context, tenantId string, tenantSpec UpdateTenantSpec) (*TenantSpec, error) {
	tenantRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	currentTenant, err := tenantRepository.GetByTenantId(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	currentTenant.Name = tenantSpec.Name
	err = tenantRepository.UpdateByTenantId(ctx, tenantId, *currentTenant)
	if err != nil {
		return nil, err
	}

	updatedTenant, err := tenantRepository.GetByTenantId(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	return updatedTenant.ToTenantSpec(), nil
}

func (svc TenantService) DeleteByTenantId(ctx context.Context, tenantId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		tenantRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		err = tenantRepository.DeleteByTenantId(txCtx, tenantId)
		if err != nil {
			return err
		}

		err = object.NewService(svc.Env()).DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypeTenant, tenantId)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func validateOrGenerateTenantIdInSpec(tenantSpec *TenantSpec) error {
	tenantIdRegExp := regexp.MustCompile(`^[a-zA-Z0-9_\-\.@\|]+$`)
	if tenantSpec.TenantId != "" {
		// Validate tenantId if provided
		if !tenantIdRegExp.Match([]byte(tenantSpec.TenantId)) {
			return service.NewInvalidParameterError("tenantId", "must be provided and can only contain alphanumeric characters and/or '-', '_', '@', and '|'")
		}
	} else {
		// Generate a TenantId for the tenant if one isn't supplied
		generatedUUID, err := uuid.NewRandom()
		if err != nil {
			return service.NewInternalError("unable to generate random UUID for tenant")
		}
		tenantSpec.TenantId = generatedUUID.String()
	}
	return nil
}
