package authz

import (
	"net/http"

	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc WarrantService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/warrants",
			Method:  "POST",
			Handler: middleware.Chain(
				service.NewRouteHandler(svc, CreateHandler),
			),
		},

		// get
		service.WarrantRoute{
			Pattern: "/v1/warrants",
			Method:  "GET",
			Handler: middleware.Chain(
				service.NewRouteHandler(svc, ListHandler),
				middleware.ListMiddleware[WarrantListParamParser],
			),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/warrants",
			Method:  "DELETE",
			Handler: middleware.Chain(
				service.NewRouteHandler(svc, DeleteHandler),
			),
		},
	}, nil
}

func CreateHandler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	var warrantSpec WarrantSpec
	err := service.ParseJSONBody(r.Body, &warrantSpec)
	if err != nil {
		return err
	}

	createdWarrant, err := svc.Create(r.Context(), warrantSpec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdWarrant)
	return nil
}

func ListHandler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	listParams := middleware.GetListParamsFromContext(r.Context())
	queryParams := r.URL.Query()
	filters := FilterOptions{
		ObjectType: queryParams.Get("objectType"),
		ObjectId:   queryParams.Get("objectId"),
		Relation:   queryParams.Get("relation"),
		Subject: &SubjectSpec{
			ObjectType: queryParams.Get("subjectType"),
			ObjectId:   queryParams.Get("subjectId"),
		},
	}
	subjectRelation := queryParams.Get("subjectRelation")
	if subjectRelation != "" {
		filters.Subject.Relation = &subjectRelation
	}

	warrants, err := svc.List(r.Context(), &filters, listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, warrants)
	return nil
}

func DeleteHandler(svc WarrantService, w http.ResponseWriter, r *http.Request) error {
	var warrantSpec WarrantSpec
	err := service.ParseJSONBody(r.Body, &warrantSpec)
	if err != nil {
		return err
	}

	err = svc.Delete(r.Context(), warrantSpec)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
