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

	"github.com/warrant-dev/warrant/pkg/service"
)

func (svc WarrantService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/warrants",
			Method:  "POST",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, createHandler),
			),
		},
		service.WarrantRoute{
			Pattern: "/v2/warrants",
			Method:  "POST",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, createHandler),
			),
		},

		// list
		service.WarrantRoute{
			Pattern: "/v1/warrants",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listV1Handler),
				service.ListMiddleware[WarrantListParamParser],
			),
		},
		service.WarrantRoute{
			Pattern: "/v2/warrants",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listV2Handler),
				service.ListMiddleware[WarrantListParamParser],
			),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/warrants",
			Method:  "DELETE",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, deleteHandler),
			),
		},
		service.WarrantRoute{
			Pattern: "/v2/warrants",
			Method:  "DELETE",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, deleteHandler),
			),
		},
	}, nil
}

func createHandler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	var spec CreateWarrantSpec
	err := service.ParseJSONBody(r.Context(), r.Body, &spec)
	if err != nil {
		return err
	}

	// TODO: move into a custom golang-validate function
	if spec.Policy != "" {
		err := spec.Policy.Validate()
		if err != nil {
			return service.NewInvalidParameterError("policy", err.Error())
		}
	}

	createdWarrant, _, err := svc.Create(r.Context(), spec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdWarrant)
	return nil
}

func listV1Handler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	warrants, _, _, err := svc.List(
		r.Context(),
		*buildFilterOptions(r),
		service.GetListParamsFromContext[WarrantListParamParser](r.Context()),
	)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, warrants)
	return nil
}

func listV2Handler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	warrants, prevCursor, nextCursor, err := svc.List(
		r.Context(),
		*buildFilterOptions(r),
		service.GetListParamsFromContext[WarrantListParamParser](r.Context()),
	)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, ListWarrantsSpecV2{
		Results:    warrants,
		PrevCursor: prevCursor,
		NextCursor: nextCursor,
	})
	return nil
}

func deleteHandler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	var spec DeleteWarrantSpec
	err := service.ParseJSONBody(r.Context(), r.Body, &spec)
	if err != nil {
		return err
	}

	_, err = svc.Delete(r.Context(), spec)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

func buildFilterOptions(r *http.Request) *FilterParams {
	var filterOptions FilterParams
	queryParams := r.URL.Query()

	if queryParams.Has("objectType") {
		filterOptions.ObjectType = queryParams.Get("objectType")
	}

	if queryParams.Has("objectId") {
		filterOptions.ObjectId = queryParams.Get("objectId")
	}

	if queryParams.Has("relation") {
		filterOptions.Relation = queryParams.Get("relation")
	}

	if queryParams.Has("subjectType") {
		filterOptions.SubjectType = queryParams.Get("subjectType")
	}

	if queryParams.Has("subjectId") {
		filterOptions.SubjectId = queryParams.Get("subjectId")
	}

	if queryParams.Has("subjectRelation") {
		filterOptions.SubjectRelation = queryParams.Get("subjectRelation")
	}

	if queryParams.Has("policy") {
		filterOptions.Policy = Policy(queryParams.Get("policy"))
	}

	return &filterOptions
}
