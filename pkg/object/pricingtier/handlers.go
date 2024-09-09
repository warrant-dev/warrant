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

func (svc PricingTierService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/pricing-tiers",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, createHandler),
		},

		// list
		service.WarrantRoute{
			Pattern: "/v1/pricing-tiers",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listHandler),
				service.ListMiddleware[PricingTierListParamParser],
			),
		},

		// get
		service.WarrantRoute{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, getHandler),
		},

		// update
		service.WarrantRoute{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, updateHandler),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, deleteHandler),
		},
	}, nil
}

func createHandler(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
	var newPricingTier PricingTierSpec
	err := service.ParseJSONBody(r.Context(), r.Body, &newPricingTier)
	if err != nil {
		return err
	}

	createdPricingTier, err := svc.Create(r.Context(), newPricingTier)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdPricingTier)
	return nil
}

func getHandler(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
	pricingTierId := mux.Vars(r)["pricingTierId"]
	pricingTier, err := svc.GetByPricingTierId(r.Context(), pricingTierId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, pricingTier)
	return nil
}

func listHandler(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext[PricingTierListParamParser](r.Context())
	pricingTiers, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, pricingTiers)
	return nil
}

func updateHandler(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
	var updatePricingTier UpdatePricingTierSpec
	err := service.ParseJSONBody(r.Context(), r.Body, &updatePricingTier)
	if err != nil {
		return err
	}

	pricingTierId := mux.Vars(r)["pricingTierId"]
	updatedPricingTier, err := svc.UpdateByPricingTierId(r.Context(), pricingTierId, updatePricingTier)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedPricingTier)
	return nil
}

func deleteHandler(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
	pricingTierId := mux.Vars(r)["pricingTierId"]
	if pricingTierId == "" {
		return service.NewMissingRequiredParameterError("pricingTierId")
	}

	err := svc.DeleteByPricingTierId(r.Context(), pricingTierId)
	if err != nil {
		return err
	}

	return nil
}
