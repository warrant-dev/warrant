package authz

import (
	"net/http"

	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc WarrantService) GetRoutes() []service.Route {
	return []service.Route{
		// create
		{
			Pattern: "/v1/warrants",
			Method:  "POST",
			Handler: middleware.ChainMiddleware(
				service.NewRouteHandler(svc.Env(), create),
			),
		},

		// get
		{
			Pattern: "/v1/warrants",
			Method:  "GET",
			Handler: middleware.ChainMiddleware(
				service.NewRouteHandler(svc.Env(), list),
				middleware.ListMiddleware[WarrantListParamParser],
			),
		},

		// delete
		{
			Pattern: "/v1/warrants",
			Method:  "DELETE",
			Handler: middleware.ChainMiddleware(
				service.NewRouteHandler(svc.Env(), delete),
			),
		},
	}
}

func create(env service.Env, w http.ResponseWriter, r *http.Request) error {
	var warrantSpec WarrantSpec
	err := service.ParseJSONBody(r.Body, &warrantSpec)
	if err != nil {
		return err
	}

	createdWarrant, err := NewService(env).Create(r.Context(), warrantSpec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdWarrant)
	return nil
}

func list(env service.Env, w http.ResponseWriter, r *http.Request) error {
	listParams := middleware.GetListParamsFromContext(r.Context())
	queryParams := r.URL.Query()
	filters := FilterOptions{
		ObjectType: queryParams.Get("objectType"),
		ObjectId:   queryParams.Get("objectId"),
		Relation:   queryParams.Get("relation"),
		Subject: &SubjectSpec{
			ObjectType: queryParams.Get("subjectType"),
			ObjectId:   queryParams.Get("subjectId"),
			Relation:   queryParams.Get("subjectRelation"),
		},
	}

	warrants, err := NewService(env).List(r.Context(), &filters, listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, warrants)
	return nil
}

func delete(env service.Env, w http.ResponseWriter, r *http.Request) error {
	var warrantSpec WarrantSpec
	err := service.ParseJSONBody(r.Body, &warrantSpec)
	if err != nil {
		return err
	}

	err = NewService(env).Delete(r.Context(), warrantSpec)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
