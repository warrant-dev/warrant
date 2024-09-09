// Copyright 2024 WorkOS, Inc.
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

func (svc RoleService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/roles",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, createHandler),
		},

		// list
		service.WarrantRoute{
			Pattern: "/v1/roles",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listHandler),
				service.ListMiddleware[RoleListParamParser],
			),
		},

		// get
		service.WarrantRoute{
			Pattern: "/v1/roles/{roleId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, getHandler),
		},

		// update
		service.WarrantRoute{
			Pattern: "/v1/roles/{roleId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/roles/{roleId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/roles/{roleId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, deleteHandler),
		},
	}, nil
}

func createHandler(svc RoleService, w http.ResponseWriter, r *http.Request) error {
	var newRole RoleSpec
	err := service.ParseJSONBody(r.Context(), r.Body, &newRole)
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

func getHandler(svc RoleService, w http.ResponseWriter, r *http.Request) error {
	roleId := mux.Vars(r)["roleId"]
	role, err := svc.GetByRoleId(r.Context(), roleId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, role)
	return nil
}

func listHandler(svc RoleService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext[RoleListParamParser](r.Context())
	roles, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, roles)
	return nil
}

func updateHandler(svc RoleService, w http.ResponseWriter, r *http.Request) error {
	var updateRole UpdateRoleSpec
	err := service.ParseJSONBody(r.Context(), r.Body, &updateRole)
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

func deleteHandler(svc RoleService, w http.ResponseWriter, r *http.Request) error {
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
