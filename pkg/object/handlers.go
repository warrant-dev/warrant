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
	"net/url"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/pkg/service"
)

func (svc ObjectService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/objects",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, createHandler),
		},
		service.WarrantRoute{
			Pattern: "/v2/objects",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, createHandler),
		},

		// list
		service.WarrantRoute{
			Pattern: "/v1/objects",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listHandlerV1),
				service.ListMiddleware[ObjectListParamParser],
			),
		},
		service.WarrantRoute{
			Pattern: "/v2/objects",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listHandlerV2),
				service.ListMiddleware[ObjectListParamParser],
			),
		},

		// get
		service.WarrantRoute{
			Pattern: "/v1/objects/{objectType}/{objectId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, getHandler),
		},
		service.WarrantRoute{
			Pattern: "/v2/objects/{objectType}/{objectId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, getHandler),
		},

		// update
		service.WarrantRoute{
			Pattern: "/v1/objects/{objectType}/{objectId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/objects/{objectType}/{objectId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v2/objects/{objectType}/{objectId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v2/objects/{objectType}/{objectId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/objects/{objectType}/{objectId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, deleteHandler),
		},
		service.WarrantRoute{
			Pattern: "/v2/objects/{objectType}/{objectId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, deleteHandler),
		},
	}, nil
}

func createHandler(svc ObjectService, w http.ResponseWriter, r *http.Request) error {
	var objectSpec CreateObjectSpec
	err := service.ParseJSONBody(r.Context(), r.Body, &objectSpec)
	if err != nil {
		return err
	}

	createdObject, err := svc.Create(r.Context(), objectSpec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdObject)
	return nil
}

func listHandlerV1(svc ObjectService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext[ObjectListParamParser](r.Context())
	queryParams := r.URL.Query()
	objectType, err := url.QueryUnescape(queryParams.Get("objectType"))
	if err != nil {
		return service.NewInvalidParameterError("objectType", "")
	}

	filterOptions := FilterOptions{ObjectType: objectType}
	objects, _, _, err := svc.List(r.Context(), &filterOptions, listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, objects)
	return nil
}

func listHandlerV2(svc ObjectService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext[ObjectListParamParser](r.Context())
	queryParams := r.URL.Query()
	objectType, err := url.QueryUnescape(queryParams.Get("objectType"))
	if err != nil {
		return service.NewInvalidParameterError("objectType", "")
	}

	filterOptions := FilterOptions{ObjectType: objectType}
	objects, prevCursor, nextCursor, err := svc.List(r.Context(), &filterOptions, listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, ListObjectsSpecV2{
		Results:    objects,
		PrevCursor: prevCursor,
		NextCursor: nextCursor,
	})
	return nil
}

func getHandler(svc ObjectService, w http.ResponseWriter, r *http.Request) error {
	objectType := mux.Vars(r)["objectType"]
	objectId := mux.Vars(r)["objectId"]
	object, err := svc.GetByObjectTypeAndId(r.Context(), objectType, objectId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, object)
	return nil
}

func updateHandler(svc ObjectService, w http.ResponseWriter, r *http.Request) error {
	var updateObject UpdateObjectSpec
	err := service.ParseJSONBody(r.Context(), r.Body, &updateObject)
	if err != nil {
		return err
	}

	objectType := mux.Vars(r)["objectType"]
	objectId := mux.Vars(r)["objectId"]
	updatedObject, err := svc.UpdateByObjectTypeAndId(r.Context(), objectType, objectId, updateObject)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedObject)
	return nil
}

func deleteHandler(svc ObjectService, w http.ResponseWriter, r *http.Request) error {
	objectType := mux.Vars(r)["objectType"]
	objectId := mux.Vars(r)["objectId"]
	_, err := svc.DeleteByObjectTypeAndId(r.Context(), objectType, objectId)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
