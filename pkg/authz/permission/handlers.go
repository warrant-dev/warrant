package authz

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc PermissionService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/permissions",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, CreateHandler),
		},

		// get
		service.WarrantRoute{
			Pattern: "/v1/permissions",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, ListHandler),
				service.ListMiddleware[PermissionListParamParser],
			),
		},
		service.WarrantRoute{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, GetHandler),
		},

		// update
		service.WarrantRoute{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, DeleteHandler),
		},
	}, nil
}

func CreateHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	var newPermission PermissionSpec
	err := service.ParseJSONBody(r.Body, &newPermission)
	if err != nil {
		return err
	}

	createdPermission, err := svc.Create(r.Context(), newPermission)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdPermission)
	return nil
}

func GetHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	permissionId := mux.Vars(r)["permissionId"]
	permission, err := svc.GetByPermissionId(r.Context(), permissionId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, permission)
	return nil
}

func ListHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext(r.Context())
	permissions, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, permissions)
	return nil
}

func UpdateHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	var updatePermission UpdatePermissionSpec
	err := service.ParseJSONBody(r.Body, &updatePermission)
	if err != nil {
		return err
	}

	permissionId := mux.Vars(r)["permissionId"]
	updatedPermission, err := svc.UpdateByPermissionId(r.Context(), permissionId, updatePermission)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedPermission)
	return nil
}

func DeleteHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	permissionId := mux.Vars(r)["permissionId"]
	if permissionId == "" {
		return service.NewMissingRequiredParameterError("permissionId")
	}

	err := svc.DeleteByPermissionId(r.Context(), permissionId)
	if err != nil {
		return err
	}

	return nil
}
