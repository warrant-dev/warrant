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

func (svc PricingTierService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/pricing-tiers",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, CreateHandler),
		},

		// list
		service.WarrantRoute{
			Pattern: "/v1/pricing-tiers",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, ListHandler),
				service.ListMiddleware[PricingTierListParamParser],
			),
		},

		// get
		service.WarrantRoute{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, GetHandler),
		},

		// update
		service.WarrantRoute{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, DeleteHandler),
		},
	}, nil
}

func CreateHandler(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
	var newPricingTier PricingTierSpec
	err := service.ParseJSONBody(r.Body, &newPricingTier)
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

func GetHandler(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
	pricingTierId := mux.Vars(r)["pricingTierId"]
	pricingTier, err := svc.GetByPricingTierId(r.Context(), pricingTierId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, pricingTier)
	return nil
}

func ListHandler(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext[PricingTierListParamParser](r.Context())
	pricingTiers, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, pricingTiers)
	return nil
}

func UpdateHandler(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
	var updatePricingTier UpdatePricingTierSpec
	err := service.ParseJSONBody(r.Body, &updatePricingTier)
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

func DeleteHandler(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
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
