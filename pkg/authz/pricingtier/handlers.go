package authz

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc PricingTierService) Routes() []service.Route {
	return []service.Route{
		// create
		{
			Pattern: "/v1/pricing-tiers",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, create),
		},

		// get
		{
			Pattern: "/v1/pricing-tiers",
			Method:  "GET",
			Handler: middleware.ChainMiddleware(
				service.NewRouteHandler(svc, list),
				middleware.ListMiddleware[PricingTierListParamParser],
			),
		},
		{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, get),
		},

		// update
		{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, update),
		},
		{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, update),
		},

		// delete
		{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, delete),
		},
	}
}

func create(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
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

func get(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
	pricingTierIdParam := mux.Vars(r)["pricingTierId"]
	pricingTierId, err := url.QueryUnescape(pricingTierIdParam)
	if err != nil {
		return service.NewInvalidParameterError("pricingTierId", "")
	}

	pricingTier, err := svc.GetByPricingTierId(r.Context(), pricingTierId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, pricingTier)
	return nil
}

func list(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
	listParams := middleware.GetListParamsFromContext(r.Context())
	pricingTiers, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, pricingTiers)
	return nil
}

func update(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
	var updatePricingTier UpdatePricingTierSpec
	err := service.ParseJSONBody(r.Body, &updatePricingTier)
	if err != nil {
		return err
	}

	pricingTierIdParam := mux.Vars(r)["pricingTierId"]
	pricingTierId, err := url.QueryUnescape(pricingTierIdParam)
	if err != nil {
		return service.NewInvalidParameterError("pricingTierId", "")
	}

	updatedPricingTier, err := svc.UpdateByPricingTierId(r.Context(), pricingTierId, updatePricingTier)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedPricingTier)
	return nil
}

func delete(svc PricingTierService, w http.ResponseWriter, r *http.Request) error {
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
