package tenant

import (
	"context"
	"regexp"

	"github.com/google/uuid"
	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypeTenant = "tenant"

type TenantService struct {
	service.BaseService
	repo      TenantRepository
	eventSvc  event.EventService
	objectSvc object.ObjectService
}

func NewService(env service.Env, repo TenantRepository, eventSvc event.EventService, objectSvc object.ObjectService) TenantService {
	return TenantService{
		BaseService: service.NewBaseService(env),
		repo:        repo,
		eventSvc:    eventSvc,
		objectSvc:   objectSvc,
	}
}

func (svc TenantService) Create(ctx context.Context, tenantSpec TenantSpec) (*TenantSpec, error) {
	err := validateOrGenerateTenantIdInSpec(&tenantSpec)
	if err != nil {
		return nil, err
	}

	var newTenant Model
	err = svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		createdObject, err := svc.objectSvc.Create(txCtx, *tenantSpec.ToObjectSpec())
		if err != nil {
			switch err.(type) {
			case *service.DuplicateRecordError:
				return service.NewDuplicateRecordError("Tenant", tenantSpec.TenantId, "A tenant with the given tenantId already exists")
			default:
				return err
			}
		}

		_, err = svc.repo.GetByTenantId(txCtx, tenantSpec.TenantId)
		if err == nil {
			return service.NewDuplicateRecordError("Tenant", tenantSpec.TenantId, "A tenant with the given tenantId already exists")
		}

		newTenantId, err := svc.repo.Create(txCtx, tenantSpec.ToTenant(createdObject.ID))
		if err != nil {
			return err
		}

		newTenant, err = svc.repo.GetById(txCtx, newTenantId)
		if err != nil {
			return err
		}

		svc.eventSvc.TrackResourceCreated(txCtx, ResourceTypeTenant, newTenant.GetTenantId(), newTenant.ToTenantSpec())
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

	currentTenant.SetName(tenantSpec.Name)
	err = tenantRepository.UpdateByTenantId(ctx, tenantId, currentTenant)
	if err != nil {
		return nil, err
	}

	updatedTenant, err := tenantRepository.GetByTenantId(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	updatedTenantSpec := updatedTenant.ToTenantSpec()
	svc.eventSvc.TrackResourceUpdated(ctx, ResourceTypeTenant, updatedTenant.GetTenantId(), updatedTenantSpec)
	return updatedTenantSpec, nil
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

		err = svc.objectSvc.DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypeTenant, tenantId)
		if err != nil {
			return err
		}

		svc.eventSvc.TrackResourceDeleted(ctx, ResourceTypeTenant, tenantId, nil)
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
