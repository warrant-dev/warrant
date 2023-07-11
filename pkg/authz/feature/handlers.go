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
	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc FeatureService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/features",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, CreateHandler),
		},

		// get
		service.WarrantRoute{
			Pattern: "/v1/features",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, ListHandler),
				service.ListMiddleware[FeatureListParamParser],
			),
		},
		service.WarrantRoute{
			Pattern: "/v1/features/{featureId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, GetHandler),
		},

		// update
		service.WarrantRoute{
			Pattern: "/v1/features/{featureId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/features/{featureId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/features/{featureId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, DeleteHandler),
		},
	}, nil
}

func CreateHandler(svc FeatureService, w http.ResponseWriter, r *http.Request) error {
	var newFeature FeatureSpec
	err := service.ParseJSONBody(r.Body, &newFeature)
	if err != nil {
		return err
	}

	createdFeature, err := svc.Create(r.Context(), newFeature)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdFeature)
	return nil
}

func GetHandler(svc FeatureService, w http.ResponseWriter, r *http.Request) error {
	featureId := mux.Vars(r)["featureId"]
	feature, err := svc.GetByFeatureId(r.Context(), featureId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, feature)
	return nil
}

func ListHandler(svc FeatureService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext[FeatureListParamParser](r.Context())
	features, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, features)
	return nil
}

func UpdateHandler(svc FeatureService, w http.ResponseWriter, r *http.Request) error {
	var updateFeature UpdateFeatureSpec
	err := service.ParseJSONBody(r.Body, &updateFeature)
	if err != nil {
		return err
	}

	featureId := mux.Vars(r)["featureId"]
	updatedFeature, err := svc.UpdateByFeatureId(r.Context(), featureId, updateFeature)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedFeature)
	return nil
}

func DeleteHandler(svc FeatureService, w http.ResponseWriter, r *http.Request) error {
	featureId := mux.Vars(r)["featureId"]
	if featureId == "" {
		return service.NewMissingRequiredParameterError("featureId")
	}

	newWookie, err := svc.DeleteByFeatureId(r.Context(), featureId)
	if err != nil {
		return err
	}
	wookie.AddAsResponseHeader(w, newWookie)

	return nil
}
