package tenant

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc TenantService) Routes() []service.Route {
	return []service.Route{
		// create
		{
			Pattern: "/v1/tenants",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, create),
		},

		// get
		{
			Pattern: "/v1/tenants/{tenantId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, get),
		},
		{
			Pattern: "/v1/tenants",
			Method:  "GET",
			Handler: middleware.ChainMiddleware(
				service.NewRouteHandler(svc, list),
				middleware.ListMiddleware[TenantListParamParser],
			),
		},

		// update
		{
			Pattern: "/v1/tenants/{tenantId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, update),
		},
		{
			Pattern: "/v1/tenants/{tenantId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, update),
		},

		// delete
		{
			Pattern: "/v1/tenants/{tenantId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, delete),
		},
	}
}

func create(svc TenantService, w http.ResponseWriter, r *http.Request) error {
	var newTenant TenantSpec
	err := service.ParseJSONBody(r.Body, &newTenant)
	if err != nil {
		return err
	}

	createdTenant, err := svc.Create(r.Context(), newTenant)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdTenant)
	return nil
}

func get(svc TenantService, w http.ResponseWriter, r *http.Request) error {
	tenantIdParam := mux.Vars(r)["tenantId"]
	tenantId, err := url.QueryUnescape(tenantIdParam)
	if err != nil {
		return service.NewInvalidParameterError("tenantId", "")
	}

	tenant, err := svc.GetByTenantId(r.Context(), tenantId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, tenant)
	return nil
}

func list(svc TenantService, w http.ResponseWriter, r *http.Request) error {
	listParams := middleware.GetListParamsFromContext(r.Context())
	tenants, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, tenants)
	return nil
}

func update(svc TenantService, w http.ResponseWriter, r *http.Request) error {
	var updateTenant UpdateTenantSpec
	err := service.ParseJSONBody(r.Body, &updateTenant)
	if err != nil {
		return err
	}

	tenantIdParam := mux.Vars(r)["tenantId"]
	tenantId, err := url.QueryUnescape(tenantIdParam)
	if err != nil {
		return service.NewInvalidParameterError("tenantId", "")
	}

	updatedTenant, err := svc.UpdateByTenantId(r.Context(), tenantId, updateTenant)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedTenant)
	return nil
}

func delete(svc TenantService, w http.ResponseWriter, r *http.Request) error {
	tenantId := mux.Vars(r)["tenantId"]
	err := svc.DeleteByTenantId(r.Context(), tenantId)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
