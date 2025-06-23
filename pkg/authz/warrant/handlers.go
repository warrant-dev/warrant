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

package authz

import (
	"context"
	"github.com/warrant-dev/warrant/pkg/wookie"
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
		// mgmt create
		service.WarrantRoute{
			Pattern: "/mgmt/warrants",
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
		// mgmt list
		service.WarrantRoute{
			Pattern: "/mgmt/warrants",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, list4MgmtHandler),
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

		service.WarrantRoute{
			Pattern: "/v2/warrants/batch",
			Method:  "DELETE",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, batchDeleteHandler),
			),
		},

		service.WarrantRoute{
			Pattern: "/mgmt/warrant/list/org/apps",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listAppsHandler),
			),
		},
	}, nil
}
func createHandler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	var specs BatchWarrantSpec
	ctx := r.Context()
	err := service.ParseJSONBody(ctx, r.Body, &specs)
	if err != nil {
		return err
	}
	var createdWarrants []*WarrantSpec
	for _, spec := range specs.Warrants {
		if spec.Policy != "" {
			err := spec.Policy.Validate()
			if err != nil {
				return service.NewInvalidParameterError("policy", err.Error())
			}
		}

		orgIdInCtx := ctx.Value(wookie.OrgIdKey)
		if orgIdInCtx == nil || orgIdInCtx == "" {
			ctx = context.WithValue(ctx, wookie.OrgIdKey, spec.OrgId)
		}

		createdWarrant, _, err := svc.Create(ctx, spec)
		if err != nil {
			return err
		}
		createdWarrants = append(createdWarrants, createdWarrant)
	}

	service.SendJSONResponse(w, createdWarrants)
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

func list4MgmtHandler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	return listV2Handler(svc, w, r)
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

func batchDeleteHandler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	var specs BatchDeleteWarrantSpec
	err := service.ParseJSONBody(r.Context(), r.Body, &specs)
	if err != nil {
		return err
	}

	if len(specs.Warrants) == 0 || len(specs.Warrants) >= MaxDeleteCountLimit {
		return service.NewInvalidParameterError("deleteSize", "batch delete size must be less than 500 and greater than 1")
	}

	for _, spec := range specs.Warrants {
		err = deleteOneWarrant(svc, r.Context(), spec)
		if err != nil {
			return err
		}
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

func deleteHandler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	var spec DeleteWarrantSpec
	err := service.ParseJSONBody(r.Context(), r.Body, &spec)
	if err != nil {
		return err
	}

	err = deleteOneWarrant(svc, r.Context(), spec)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

func deleteOneWarrant(svc WarrantService, context context.Context, spec DeleteWarrantSpec) error {
	if !spec.HasAnyValue() {
		return service.NewInvalidParameterError("objectType", "must specify at least one of objectType or objectId or Subject")
	}

	_, err := svc.Delete(context, spec)
	if err != nil {
		return err
	}
	return nil
}

func listAppsHandler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	apps, err := svc.ListWarrantApps(r.Context())
	if err != nil {
		return err
	}
	service.SendJSONResponse(w, apps)
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

	return &filterOptions
}
