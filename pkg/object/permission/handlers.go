// Copyright 2023 Forerunner Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/pkg/service"
)

func (svc PermissionService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/permissions",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, createHandler),
		},

		// list
		service.WarrantRoute{
			Pattern: "/v1/permissions",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listHandler),
				service.ListMiddleware[PermissionListParamParser],
			),
		},

		// get
		service.WarrantRoute{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, getHandler),
		},

		// update
		service.WarrantRoute{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/permissions/{permissionId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, deleteHandler),
		},
	}, nil
}

func createHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
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

func getHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	permissionId := mux.Vars(r)["permissionId"]
	permission, err := svc.GetByPermissionId(r.Context(), permissionId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, permission)
	return nil
}

func listHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext[PermissionListParamParser](r.Context())
	permissions, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, permissions)
	return nil
}

func updateHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
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

func deleteHandler(svc PermissionService, w http.ResponseWriter, r *http.Request) error {
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
