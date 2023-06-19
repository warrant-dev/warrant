package authz

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/pkg/service"
)

func (svc UserService) Routes() ([]service.Route, error) {
	return []service.Route{
		// create
		service.WarrantRoute{
			Pattern: "/v1/users",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, CreateHandler),
		},

		// get
		service.WarrantRoute{
			Pattern: "/v1/users/{userId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, GetHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/users",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, ListHandler),
				service.ListMiddleware[UserListParamParser],
			),
		},

		// delete
		service.WarrantRoute{
			Pattern: "/v1/users/{userId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, DeleteHandler),
		},

		// update
		service.WarrantRoute{
			Pattern: "/v1/users/{userId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},
		service.WarrantRoute{
			Pattern: "/v1/users/{userId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},
	}, nil
}

func CreateHandler(svc UserService, w http.ResponseWriter, r *http.Request) error {
	var userSpec UserSpec
	err := service.ParseJSONBody(r.Body, &userSpec)
	if err != nil {
		return err
	}

	createdUser, err := svc.Create(r.Context(), userSpec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdUser)
	return nil
}

func GetHandler(svc UserService, w http.ResponseWriter, r *http.Request) error {
	userId := mux.Vars(r)["userId"]
	user, err := svc.GetByUserId(r.Context(), userId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, user)
	return nil
}

func ListHandler(svc UserService, w http.ResponseWriter, r *http.Request) error {
	listParams := service.GetListParamsFromContext[UserListParamParser](r.Context())
	users, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, users)
	return nil
}

func UpdateHandler(svc UserService, w http.ResponseWriter, r *http.Request) error {
	var updateUser UpdateUserSpec
	err := service.ParseJSONBody(r.Body, &updateUser)
	if err != nil {
		return err
	}

	userId := mux.Vars(r)["userId"]
	updatedUser, err := svc.UpdateByUserId(r.Context(), userId, updateUser)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedUser)
	return nil
}

func DeleteHandler(svc UserService, w http.ResponseWriter, r *http.Request) error {
	userId := mux.Vars(r)["userId"]
	newWookie, err := svc.DeleteByUserId(r.Context(), userId)
	if err != nil {
		return err
	}
	if newWookie != nil {
		w.Header().Set("Warrant-Token", newWookie.AsString())
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
