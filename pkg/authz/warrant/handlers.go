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

package authz

import (
	"net/http"
	"strings"

	"github.com/warrant-dev/warrant/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc WarrantService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/warrants",
			Method:  "POST",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, CreateHandler),
			),
		},

		// list
		service.WarrantRoute{
			Pattern: "/v1/warrants",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, ListHandler),
				service.ListMiddleware[WarrantListParamParser],
			),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/warrants",
			Method:  "DELETE",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, DeleteHandler),
			),
		},
	}, nil
}

func CreateHandler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	var warrantSpec WarrantSpec
	err := service.ParseJSONBody(r.Body, &warrantSpec)
	if err != nil {
		return err
	}

	// TODO: move into a custom golang-validate function
	if warrantSpec.Policy != "" {
		err := warrantSpec.Policy.Validate()
		if err != nil {
			return service.NewInvalidParameterError("policy", err.Error())
		}
	}

	createdWarrant, err := svc.Create(r.Context(), warrantSpec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdWarrant)
	return nil
}

func ListHandler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	warrants, err := svc.List(
		r.Context(),
		buildFilterOptions(r),
		service.GetListParamsFromContext[WarrantListParamParser](r.Context()),
	)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, warrants)
	return nil
}

func DeleteHandler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	var warrantSpec WarrantSpec
	err := service.ParseJSONBody(r.Body, &warrantSpec)
	if err != nil {
		return err
	}

	err = svc.Delete(r.Context(), warrantSpec)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

func buildFilterOptions(r *http.Request) *FilterOptions {
	var filterOptions FilterOptions
	queryParams := r.URL.Query()

	if queryParams.Has("objectType") {
		filterOptions.ObjectType = strings.Split(queryParams.Get("objectType"), ",")
	}

	if queryParams.Has("objectId") {
		filterOptions.ObjectId = strings.Split(queryParams.Get("objectId"), ",")
	}

	if queryParams.Has("relation") {
		filterOptions.Relation = strings.Split(queryParams.Get("relation"), ",")
	}

	if queryParams.Has("subjectType") {
		filterOptions.SubjectType = strings.Split(queryParams.Get("subjectType"), ",")
	}

	if queryParams.Has("subjectId") {
		filterOptions.SubjectId = strings.Split(queryParams.Get("subjectId"), ",")
	}

	if queryParams.Has("subjectRelation") {
		filterOptions.SubjectRelation = strings.Split(queryParams.Get("subjectRelation"), ",")
	}

	if queryParams.Has("policy") {
		filterOptions.Policy = Policy(queryParams.Get("policy"))
	}

	return &filterOptions
}
