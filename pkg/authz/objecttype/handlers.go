package authz

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc ObjectTypeService) Routes() []service.Route {
	return []service.Route{
		// create
		{
			Pattern: "/v1/object-types",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, CreateHandler),
		},

		// get
		{
			Pattern: "/v1/object-types",
			Method:  "GET",
			Handler: middleware.Chain(
				service.NewRouteHandler(svc, ListHandler),
				middleware.ListMiddleware[ObjectTypeListParamParser],
			),
		},

		// list
		{
			Pattern: "/v1/object-types/{type}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, GetHandler),
		},

		// update
		{
			Pattern: "/v1/object-types/{type}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},
		{
			Pattern: "/v1/object-types/{type}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},

		// delete
		{
			Pattern: "/v1/object-types/{type}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, DeleteHandler),
		},
	}
}

func CreateHandler(svc ObjectTypeService, w http.ResponseWriter, r *http.Request) error {
	var objectTypeSpec ObjectTypeSpec
	err := service.ParseJSONBody(r.Body, &objectTypeSpec)
	if err != nil {
		return err
	}

	createdObjectTypeSpec, err := svc.Create(r.Context(), objectTypeSpec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdObjectTypeSpec)
	return nil
}

func ListHandler(svc ObjectTypeService, w http.ResponseWriter, r *http.Request) error {
	listParams := middleware.GetListParamsFromContext(r.Context())
	objectTypeSpecs, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, objectTypeSpecs)
	return nil
}

func GetHandler(svc ObjectTypeService, w http.ResponseWriter, r *http.Request) error {
	typeParam := mux.Vars(r)["type"]
	objectTypeSpec, err := svc.GetByTypeId(r.Context(), typeParam)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, objectTypeSpec)
	return nil
}

func UpdateHandler(svc ObjectTypeService, w http.ResponseWriter, r *http.Request) error {
	var objectTypeSpec ObjectTypeSpec
	err := service.ParseJSONBody(r.Body, &objectTypeSpec)
	if err != nil {
		return err
	}

	typeParam := mux.Vars(r)["type"]
	updatedObjectTypeSpec, err := svc.UpdateByTypeId(r.Context(), typeParam, objectTypeSpec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedObjectTypeSpec)
	return nil
}

func DeleteHandler(svc ObjectTypeService, w http.ResponseWriter, r *http.Request) error {
	typeId := mux.Vars(r)["type"]
	err := svc.DeleteByTypeId(r.Context(), typeId)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
