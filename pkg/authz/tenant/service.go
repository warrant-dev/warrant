package tenant

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypeTenant = "tenant"

type TenantService struct {
	service.BaseService
	Repository TenantRepository
	EventSvc   event.EventService
	ObjectSvc  object.ObjectService
}

func NewService(env service.Env, repository TenantRepository, eventSvc event.EventService, objectSvc object.ObjectService) TenantService {
	return TenantService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
		EventSvc:    eventSvc,
		ObjectSvc:   objectSvc,
	}
}

func (svc TenantService) Create(ctx context.Context, tenantSpec TenantSpec) (*TenantSpec, error) {
	if tenantSpec.TenantId == "" {
		// generate an id for the tenant if one isn't supplied
		generatedUUID, err := uuid.NewRandom()
		if err != nil {
			return nil, errors.New("unable to generate random UUID for tenant")
		}
		tenantSpec.TenantId = generatedUUID.String()
	}

	var newTenant Model
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		createdObject, err := svc.ObjectSvc.Create(txCtx, *tenantSpec.ToObjectSpec())
		if err != nil {
			switch err.(type) {
			case *service.DuplicateRecordError:
				return service.NewDuplicateRecordError("Tenant", tenantSpec.TenantId, "A tenant with the given tenantId already exists")
			default:
				return err
			}
		}

		_, err = svc.Repository.GetByTenantId(txCtx, tenantSpec.TenantId)
		if err == nil {
			return service.NewDuplicateRecordError("Tenant", tenantSpec.TenantId, "A tenant with the given tenantId already exists")
		}

		newTenantId, err := svc.Repository.Create(txCtx, tenantSpec.ToTenant(createdObject.ID))
		if err != nil {
			return err
		}

		newTenant, err = svc.Repository.GetById(txCtx, newTenantId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceCreated(txCtx, ResourceTypeTenant, newTenant.GetTenantId(), newTenant.ToTenantSpec())
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
	tenant, err := svc.Repository.GetByTenantId(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	return tenant.ToTenantSpec(), nil
}

func (svc TenantService) List(ctx context.Context, listParams service.ListParams) ([]TenantSpec, error) {
	tenantSpecs := make([]TenantSpec, 0)

	tenants, err := svc.Repository.List(ctx, listParams)
	if err != nil {
		return tenantSpecs, nil
	}

	for _, tenant := range tenants {
		tenantSpecs = append(tenantSpecs, *tenant.ToTenantSpec())
	}

	return tenantSpecs, nil
}

func (svc TenantService) UpdateByTenantId(ctx context.Context, tenantId string, tenantSpec UpdateTenantSpec) (*TenantSpec, error) {
	currentTenant, err := svc.Repository.GetByTenantId(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	currentTenant.SetName(tenantSpec.Name)
	err = svc.Repository.UpdateByTenantId(ctx, tenantId, currentTenant)
	if err != nil {
		return nil, err
	}

	updatedTenant, err := svc.Repository.GetByTenantId(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	updatedTenantSpec := updatedTenant.ToTenantSpec()
	err = svc.EventSvc.TrackResourceUpdated(ctx, ResourceTypeTenant, updatedTenant.GetTenantId(), updatedTenantSpec)
	if err != nil {
		return nil, err
	}

	return updatedTenantSpec, nil
}

func (svc TenantService) DeleteByTenantId(ctx context.Context, tenantId string) (*wookie.Token, error) {
	var newWookie *wookie.Token
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		err := svc.Repository.DeleteByTenantId(txCtx, tenantId)
		if err != nil {
			return err
		}

		newWookie, err = svc.ObjectSvc.DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypeTenant, tenantId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceDeleted(txCtx, ResourceTypeTenant, tenantId, nil)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newWookie, nil
}
