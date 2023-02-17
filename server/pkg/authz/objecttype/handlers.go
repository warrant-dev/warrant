package authz

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc ObjectTypeService) GetRoutes() []service.Route {
	return []service.Route{
		// create
		{
			Pattern: "/v1/object-types",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc.Env(), create),
		},

		// get
		{
			Pattern: "/v1/object-types",
			Method:  "GET",
			Handler: middleware.ChainMiddleware(
				service.NewRouteHandler(svc.Env(), list),
				middleware.ListMiddleware[ObjectTypeListParamParser],
			),
		},

		// list
		{
			Pattern: "/v1/object-types/{type}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc.Env(), getByTypeId),
		},

		// update
		{
			Pattern: "/v1/object-types/{type}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc.Env(), update),
		},
		{
			Pattern: "/v1/object-types/{type}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc.Env(), update),
		},

		// delete
		{
			Pattern: "/v1/object-types/{type}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc.Env(), delete),
		},
	}
}

func create(env service.Env, w http.ResponseWriter, r *http.Request) error {
	var objectTypeSpec ObjectTypeSpec
	err := service.ParseJSONBody(r.Body, &objectTypeSpec)
	if err != nil {
		return err
	}

	createdObjectTypeSpec, err := NewService(env).Create(objectTypeSpec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdObjectTypeSpec)
	return nil
}

func list(env service.Env, w http.ResponseWriter, r *http.Request) error {
	listParams := middleware.GetListParamsFromContext(r.Context())
	objectTypeSpecs, err := NewService(env).List(listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, objectTypeSpecs)
	return nil
}

func getByTypeId(env service.Env, w http.ResponseWriter, r *http.Request) error {
	typeParam := mux.Vars(r)["type"]
	objectTypeSpec, err := NewService(env).GetByTypeId(typeParam)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, objectTypeSpec)
	return nil
}

func update(env service.Env, w http.ResponseWriter, r *http.Request) error {
	var objectTypeSpec ObjectTypeSpec
	err := service.ParseJSONBody(r.Body, &objectTypeSpec)
	if err != nil {
		return err
	}

	typeParam := mux.Vars(r)["type"]
	updatedObjectTypeSpec, err := NewService(env).UpdateByTypeId(typeParam, objectTypeSpec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedObjectTypeSpec)
	return nil
}

func delete(env service.Env, w http.ResponseWriter, r *http.Request) error {
	typeId := mux.Vars(r)["type"]
	err := NewService(env).DeleteByTypeId(typeId)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
