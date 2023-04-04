package authz

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

func (svc ObjectService) Routes() []service.Route {
	return []service.Route{
		// create
		{
			Pattern: "/v1/objects",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, create),
		},

		// get
		{
			Pattern: "/v1/objects",
			Method:  "GET",
			Handler: middleware.ChainMiddleware(
				service.NewRouteHandler(svc, list),
				middleware.ListMiddleware[ObjectListParamParser],
			),
		},
		{
			Pattern: "/v1/objects/{objectType}/{objectId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, get),
		},

		// delete
		{
			Pattern: "/v1/objects/{objectType}/{objectId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, delete),
		},
	}
}

func create(svc ObjectService, w http.ResponseWriter, r *http.Request) error {
	var newObject ObjectSpec
	err := service.ParseJSONBody(r.Body, &newObject)
	if err != nil {
		return err
	}

	createdObject, err := svc.Create(r.Context(), newObject)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdObject)
	return nil
}

func list(svc ObjectService, w http.ResponseWriter, r *http.Request) error {
	listParams := middleware.GetListParamsFromContext(r.Context())
	queryParams := r.URL.Query()
	objectType, err := url.QueryUnescape(queryParams.Get("objectType"))
	if err != nil {
		return service.NewInvalidParameterError("objectType", "")
	}

	filterOptions := FilterOptions{ObjectType: objectType}
	objects, err := svc.List(r.Context(), &filterOptions, listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, objects)
	return nil
}

func get(svc ObjectService, w http.ResponseWriter, r *http.Request) error {
	objectType := mux.Vars(r)["objectType"]
	objectIdParam := mux.Vars(r)["objectId"]
	object, err := svc.GetByObjectId(r.Context(), objectType, objectIdParam)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, object)
	return nil
}

func delete(svc ObjectService, w http.ResponseWriter, r *http.Request) error {
	objectType := mux.Vars(r)["objectType"]
	objectId := mux.Vars(r)["objectId"]
	err := svc.DeleteByObjectTypeAndId(r.Context(), objectType, objectId)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
