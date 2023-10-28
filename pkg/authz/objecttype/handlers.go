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

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/pkg/service"
)

func (svc ObjectTypeService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/object-types",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, createHandler),
		},
		service.WarrantRoute{
			Pattern: "/v2/object-types",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, createHandler),
		},

		// list
		service.WarrantRoute{
			Pattern: "/v1/object-types",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listHandlerV1),
				service.ListMiddleware[ObjectTypeListParamParser],
			),
		},
		service.WarrantRoute{
			Pattern: "/v2/object-types",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listHandlerV2),
				service.ListMiddleware[ObjectTypeListParamParser],
			),
		},

		// get
		service.WarrantRoute{
			Pattern: "/v1/object-types/{type}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, getHandler),
		},
		service.WarrantRoute{
			Pattern: "/v2/object-types/{type}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, getHandler),
		},

		// update
		service.WarrantRoute{
			Pattern: "/v1/object-types/{type}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/object-types/{type}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v2/object-types/{type}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v2/object-types/{type}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/object-types/{type}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, deleteHandler),
		},
		service.WarrantRoute{
			Pattern: "/v2/object-types/{type}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, deleteHandler),
		},
	}, nil
}

func createHandler(svc ObjectTypeService, w http.ResponseWriter, r *http.Request) error {
	var spec CreateObjectTypeSpec
	err := service.ParseJSONBody(r.Context(), r.Body, &spec)
	if err != nil {
		return err
	}

	createdObjectTypeSpec, _, err := svc.Create(r.Context(), spec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdObjectTypeSpec)
	return nil
}

func listHandlerV1(svc ObjectTypeService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext[ObjectTypeListParamParser](r.Context())
	objectTypeSpecs, _, _, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, objectTypeSpecs)
	return nil
}

func listHandlerV2(svc ObjectTypeService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext[ObjectTypeListParamParser](r.Context())
	objectTypeSpecs, prevCursor, nextCursor, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, ListObjectTypesSpecV2{
		Results:    objectTypeSpecs,
		PrevCursor: prevCursor,
		NextCursor: nextCursor,
	})
	return nil
}

func getHandler(svc ObjectTypeService, w http.ResponseWriter, r *http.Request) error {
	typeId := mux.Vars(r)["type"]
	objectTypeSpec, err := svc.GetByTypeId(r.Context(), typeId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, objectTypeSpec)
	return nil
}

func updateHandler(svc ObjectTypeService, w http.ResponseWriter, r *http.Request) error {
	var spec UpdateObjectTypeSpec
	err := service.ParseJSONBody(r.Context(), r.Body, &spec)
	if err != nil {
		return err
	}

	typeId := mux.Vars(r)["type"]
	updatedObjectTypeSpec, _, err := svc.UpdateByTypeId(r.Context(), typeId, spec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedObjectTypeSpec)
	return nil
}

func deleteHandler(svc ObjectTypeService, w http.ResponseWriter, r *http.Request) error {
	typeId := mux.Vars(r)["type"]
	_, err := svc.DeleteByTypeId(r.Context(), typeId)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
