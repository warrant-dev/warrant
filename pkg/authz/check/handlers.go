package authz

import (
	"net/http"

	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

func (svc CheckService) GetRoutes() []service.Route {
	return []service.Route{
		// Standard Authorization
		{
			Pattern: "/v2/authorize",
			Method:  "POST",
			Handler: middleware.ChainMiddleware(
				service.NewRouteHandler(svc.Env(), authorize),
			),
		},
	}
}

func authorize(env service.Env, w http.ResponseWriter, r *http.Request) error {
	var checkManySpec CheckManySpec
	err := service.ParseJSONBody(r.Body, &checkManySpec)
	if err != nil {
		return err
	}

	checkResult, err := NewService(env).CheckMany(r.Context(), &checkManySpec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, checkResult)
	return nil
}
