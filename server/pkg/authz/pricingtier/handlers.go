package authz

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc PricingTierService) GetRoutes() []service.Route {
	return []service.Route{
		// create
		{
			Pattern: "/v1/pricing-tiers",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc.Env(), create),
		},

		// get
		{
			Pattern: "/v1/pricing-tiers",
			Method:  "GET",
			Handler: middleware.ChainMiddleware(
				service.NewRouteHandler(svc.Env(), list),
				middleware.ListMiddleware[PricingTierListParamParser],
			),
		},
		{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc.Env(), get),
		},

		// update
		{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc.Env(), update),
		},
		{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc.Env(), update),
		},

		// delete
		{
			Pattern: "/v1/pricing-tiers/{pricingTierId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc.Env(), delete),
		},
	}
}

func create(env service.Env, w http.ResponseWriter, r *http.Request) error {
	var newPricingTier PricingTierSpec
	err := service.ParseJSONBody(r.Body, &newPricingTier)
	if err != nil {
		return err
	}

	createdPricingTier, err := NewService(env).Create(r.Context(), newPricingTier)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdPricingTier)
	return nil
}

func get(env service.Env, w http.ResponseWriter, r *http.Request) error {
	pricingTierIdParam := mux.Vars(r)["pricingTierId"]
	pricingTierId, err := url.QueryUnescape(pricingTierIdParam)
	if err != nil {
		return service.NewInvalidParameterError("pricingTierId", "")
	}

	pricingTier, err := NewService(env).GetByPricingTierId(r.Context(), pricingTierId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, pricingTier)
	return nil
}

func list(env service.Env, w http.ResponseWriter, r *http.Request) error {
	listParams := middleware.GetListParamsFromContext(r.Context())
	pricingTiers, err := NewService(env).List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, pricingTiers)
	return nil
}

func update(env service.Env, w http.ResponseWriter, r *http.Request) error {
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

	updatedPricingTier, err := NewService(env).UpdateByPricingTierId(r.Context(), pricingTierId, updatePricingTier)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedPricingTier)
	return nil
}

func delete(env service.Env, w http.ResponseWriter, r *http.Request) error {
	pricingTierId := mux.Vars(r)["pricingTierId"]
	if pricingTierId == "" {
		return service.NewMissingRequiredParameterError("pricingTierId")
	}

	err := NewService(env).DeleteByPricingTierId(r.Context(), pricingTierId)
	if err != nil {
		return err
	}

	return nil
}
