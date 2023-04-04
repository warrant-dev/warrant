package authz

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc RoleService) Routes() []service.Route {
	return []service.Route{
		// create
		{
			Pattern: "/v1/roles",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, create),
		},

		// get
		{
			Pattern: "/v1/roles",
			Method:  "GET",
			Handler: middleware.ChainMiddleware(
				service.NewRouteHandler(svc, list),
				middleware.ListMiddleware[RoleListParamParser],
			),
		},
		{
			Pattern: "/v1/roles/{roleId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, get),
		},

		// update
		{
			Pattern: "/v1/roles/{roleId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, update),
		},
		{
			Pattern: "/v1/roles/{roleId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, update),
		},

		// delete
		{
			Pattern: "/v1/roles/{roleId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, delete),
		},
	}
}

func create(svc RoleService, w http.ResponseWriter, r *http.Request) error {
	var newRole RoleSpec
	err := service.ParseJSONBody(r.Body, &newRole)
	if err != nil {
		return err
	}

	createdRole, err := svc.Create(r.Context(), newRole)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdRole)
	return nil
}

func get(svc RoleService, w http.ResponseWriter, r *http.Request) error {
	roleIdParam := mux.Vars(r)["roleId"]
	roleId, err := url.QueryUnescape(roleIdParam)
	if err != nil {
		return service.NewInvalidParameterError("roleId", "")
	}

	role, err := svc.GetByRoleId(r.Context(), roleId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, role)
	return nil
}

func list(svc RoleService, w http.ResponseWriter, r *http.Request) error {
	listParams := middleware.GetListParamsFromContext(r.Context())
	roles, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, roles)
	return nil
}

func update(svc RoleService, w http.ResponseWriter, r *http.Request) error {
	var updateRole UpdateRoleSpec
	err := service.ParseJSONBody(r.Body, &updateRole)
	if err != nil {
		return err
	}

	roleIdParam := mux.Vars(r)["roleId"]
	roleId, err := url.QueryUnescape(roleIdParam)
	if err != nil {
		return service.NewInvalidParameterError("roleId", "")
	}

	updatedRole, err := svc.UpdateByRoleId(r.Context(), roleId, updateRole)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedRole)
	return nil
}

func delete(svc RoleService, w http.ResponseWriter, r *http.Request) error {
	roleId := mux.Vars(r)["roleId"]
	if roleId == "" {
		return service.NewMissingRequiredParameterError("roleId")
	}

	err := svc.DeleteByRoleId(r.Context(), roleId)
	if err != nil {
		return err
	}

	return nil
}
