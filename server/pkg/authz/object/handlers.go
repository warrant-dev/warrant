package authz

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

func (svc ObjectService) GetRoutes() []service.Route {
	return []service.Route{
		// create
		{
			Pattern: "/v1/objects",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc.Env(), create),
		},

		// get
		{
			Pattern: "/v1/objects",
			Method:  "GET",
			Handler: middleware.ChainMiddleware(
				service.NewRouteHandler(svc.Env(), list),
				middleware.ListMiddleware[ObjectListParamParser],
			),
		},
		{
			Pattern: "/v1/objects/{objectType}/{objectId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc.Env(), get),
		},

		// delete
		{
			Pattern: "/v1/objects/{objectType}/{objectId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc.Env(), delete),
		},
	}
}

func create(env service.Env, w http.ResponseWriter, r *http.Request) error {
	var newObject ObjectSpec
	err := service.ParseJSONBody(r.Body, &newObject)
	if err != nil {
		return err
	}

	createdObject, err := NewService(env).Create(r.Context(), newObject)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdObject)
	return nil
}

func list(env service.Env, w http.ResponseWriter, r *http.Request) error {
	listParams := middleware.GetListParamsFromContext(r.Context())
	queryParams := r.URL.Query()
	objectType, err := url.QueryUnescape(queryParams.Get("objectType"))
	if err != nil {
		return service.NewInvalidParameterError("objectType", "")
	}

	filterOptions := FilterOptions{ObjectType: objectType}
	objects, err := NewService(env).List(r.Context(), &filterOptions, listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, objects)
	return nil
}

func get(env service.Env, w http.ResponseWriter, r *http.Request) error {
	objectType := mux.Vars(r)["objectType"]
	objectIdParam := mux.Vars(r)["objectId"]
	object, err := NewService(env).GetByObjectId(r.Context(), objectType, objectIdParam)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, object)
	return nil
}

func delete(env service.Env, w http.ResponseWriter, r *http.Request) error {
	objectType := mux.Vars(r)["objectType"]
	objectId := mux.Vars(r)["objectId"]
	err := NewService(env).DeleteByObjectTypeAndId(r.Context(), objectType, objectId)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
