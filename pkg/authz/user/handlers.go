package authz

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

func (svc UserService) Routes() []service.Route {
	return []service.Route{
		// create
		{
			Pattern: "/v1/users",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, create),
		},

		// get
		{
			Pattern: "/v1/users/{userId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, get),
		},
		{
			Pattern: "/v1/users",
			Method:  "GET",
			Handler: middleware.ChainMiddleware(
				service.NewRouteHandler(svc, list),
				middleware.ListMiddleware[UserListParamParser],
			),
		},

		// delete
		{
			Pattern: "/v1/users/{userId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, delete),
		},

		// update
		{
			Pattern: "/v1/users/{userId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, update),
		},
		{
			Pattern: "/v1/users/{userId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, update),
		},
	}
}

func create(svc UserService, w http.ResponseWriter, r *http.Request) error {
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

func get(svc UserService, w http.ResponseWriter, r *http.Request) error {
	userIdParam := mux.Vars(r)["userId"]
	userId, err := url.QueryUnescape(userIdParam)
	if err != nil {
		return service.NewInvalidParameterError("userId", "")
	}

	user, err := svc.GetByUserId(r.Context(), userId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, user)
	return nil
}

func list(svc UserService, w http.ResponseWriter, r *http.Request) error {
	listParams := middleware.GetListParamsFromContext(r.Context())
	users, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, users)
	return nil
}

func update(svc UserService, w http.ResponseWriter, r *http.Request) error {
	var updateUser UpdateUserSpec
	err := service.ParseJSONBody(r.Body, &updateUser)
	if err != nil {
		return err
	}

	userIdParam := mux.Vars(r)["userId"]
	userId, err := url.QueryUnescape(userIdParam)
	if err != nil {
		return service.NewInvalidParameterError("userId", "")
	}

	updatedUser, err := svc.UpdateByUserId(r.Context(), userId, updateUser)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedUser)
	return nil
}

func delete(svc UserService, w http.ResponseWriter, r *http.Request) error {
	userIdParam := mux.Vars(r)["userId"]
	userId, err := url.QueryUnescape(userIdParam)
	if err != nil {
		return service.NewInvalidParameterError("userId", "")
	}

	err = svc.DeleteByUserId(r.Context(), userId)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
