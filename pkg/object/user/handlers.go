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

func (svc UserService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/users",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, createHandler),
		},

		// list
		service.WarrantRoute{
			Pattern: "/v1/users",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listHandler),
				service.ListMiddleware[UserListParamParser],
			),
		},

		// get
		service.WarrantRoute{
			Pattern: "/v1/users/{userId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, getHandler),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/users/{userId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, deleteHandler),
		},

		// update
		service.WarrantRoute{
			Pattern: "/v1/users/{userId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/users/{userId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},
	}, nil
}

func createHandler(svc UserService, w http.ResponseWriter, r *http.Request) error {
	var userSpec UserSpec
	err := service.ParseJSONBody(r.Body, &userSpec)
	if err != nil {
		return err
	}

	createdUser, err := svc.Create(r.Context(), userSpec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdUser)
	return nil
}

func getHandler(svc UserService, w http.ResponseWriter, r *http.Request) error {
	userId := mux.Vars(r)["userId"]
	user, err := svc.GetByUserId(r.Context(), userId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, user)
	return nil
}

func listHandler(svc UserService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext[UserListParamParser](r.Context())
	users, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, users)
	return nil
}

func updateHandler(svc UserService, w http.ResponseWriter, r *http.Request) error {
	var updateUser UpdateUserSpec
	err := service.ParseJSONBody(r.Body, &updateUser)
	if err != nil {
		return err
	}

	userId := mux.Vars(r)["userId"]
	updatedUser, err := svc.UpdateByUserId(r.Context(), userId, updateUser)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedUser)
	return nil
}

func deleteHandler(svc UserService, w http.ResponseWriter, r *http.Request) error {
	userId := mux.Vars(r)["userId"]
	err := svc.DeleteByUserId(r.Context(), userId)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
