package authz

import (
	"net/http"

	"github.com/gorilla/mux"
	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc RoleService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/roles",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, CreateHandler),
		},

		// get
		service.WarrantRoute{
			Pattern: "/v1/roles",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, ListHandler),
				service.ListMiddleware[RoleListParamParser],
			),
		},
		service.WarrantRoute{
			Pattern: "/v1/roles/{roleId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, GetHandler),
		},

		// update
		service.WarrantRoute{
			Pattern: "/v1/roles/{roleId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/roles/{roleId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/roles/{roleId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, DeleteHandler),
		},
	}, nil
}

func CreateHandler(svc RoleService, w http.ResponseWriter, r *http.Request) error {
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

func GetHandler(svc RoleService, w http.ResponseWriter, r *http.Request) error {
	roleId := mux.Vars(r)["roleId"]
	role, err := svc.GetByRoleId(r.Context(), roleId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, role)
	return nil
}

func ListHandler(svc RoleService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext[RoleListParamParser](r.Context())
	roles, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, roles)
	return nil
}

func UpdateHandler(svc RoleService, w http.ResponseWriter, r *http.Request) error {
	var updateRole UpdateRoleSpec
	err := service.ParseJSONBody(r.Body, &updateRole)
	if err != nil {
		return err
	}

	roleId := mux.Vars(r)["roleId"]
	updatedRole, err := svc.UpdateByRoleId(r.Context(), roleId, updateRole)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedRole)
	return nil
}

func DeleteHandler(svc RoleService, w http.ResponseWriter, r *http.Request) error {
	roleId := mux.Vars(r)["roleId"]
	if roleId == "" {
		return service.NewMissingRequiredParameterError("roleId")
	}

	newWookie, err := svc.DeleteByRoleId(r.Context(), roleId)
	if err != nil {
		return err
	}
	wookie.AddAsResponseHeader(w, newWookie)

	return nil
}
