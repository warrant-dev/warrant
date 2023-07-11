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

package tenant

import (
	"net/http"

	"github.com/gorilla/mux"
	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc TenantService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/tenants",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, CreateHandler),
		},

		// get
		service.WarrantRoute{
			Pattern: "/v1/tenants/{tenantId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, GetHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/tenants",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, ListHandler),
				service.ListMiddleware[TenantListParamParser],
			),
		},

		// update
		service.WarrantRoute{
			Pattern: "/v1/tenants/{tenantId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/tenants/{tenantId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/tenants/{tenantId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, DeleteHandler),
		},
	}, nil
}

func CreateHandler(svc TenantService, w http.ResponseWriter, r *http.Request) error {
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

func GetHandler(svc TenantService, w http.ResponseWriter, r *http.Request) error {
	tenantId := mux.Vars(r)["tenantId"]
	tenant, err := svc.GetByTenantId(r.Context(), tenantId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, tenant)
	return nil
}

func ListHandler(svc TenantService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext[TenantListParamParser](r.Context())
	tenants, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, tenants)
	return nil
}

func UpdateHandler(svc TenantService, w http.ResponseWriter, r *http.Request) error {
	var updateTenant UpdateTenantSpec
	err := service.ParseJSONBody(r.Body, &updateTenant)
	if err != nil {
		return err
	}

	tenantId := mux.Vars(r)["tenantId"]
	updatedTenant, err := svc.UpdateByTenantId(r.Context(), tenantId, updateTenant)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedTenant)
	return nil
}

func DeleteHandler(svc TenantService, w http.ResponseWriter, r *http.Request) error {
	tenantId := mux.Vars(r)["tenantId"]
	newWookie, err := svc.DeleteByTenantId(r.Context(), tenantId)
	if err != nil {
		return err
	}
	wookie.AddAsResponseHeader(w, newWookie)

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
