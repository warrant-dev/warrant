package authz

import (
	"net/http"

	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	authz "github.com/warrant-dev/warrant/pkg/authz/warrant"
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
			EnableSessionAuth: true,
		},
	}
}

func authorize(env service.Env, w http.ResponseWriter, r *http.Request) error {
	authInfo := service.GetAuthInfoFromRequestContext(r.Context())

	if authInfo.UserId != "" {
		var sessionCheckManySpec SessionCheckManySpec
		err := service.ParseJSONBody(r.Body, &sessionCheckManySpec)
		if err != nil {
			return err
		}

		warrantSpecs := make([]authz.WarrantSpec, 0)
		for _, warrantSpec := range sessionCheckManySpec.Warrants {
			warrantSpecs = append(warrantSpecs, authz.WarrantSpec{
				ObjectType: warrantSpec.ObjectType,
				ObjectId:   warrantSpec.ObjectId,
				Relation:   warrantSpec.Relation,
				Subject: &authz.SubjectSpec{
					ObjectType: objecttype.ObjectTypeUser,
					ObjectId:   authInfo.UserId,
				},
			})
		}

		checkManySpec := CheckManySpec{
			Op:             sessionCheckManySpec.Op,
			Warrants:       warrantSpecs,
			Context:        sessionCheckManySpec.Context,
			ConsistentRead: sessionCheckManySpec.ConsistentRead,
			Debug:          sessionCheckManySpec.Debug,
		}

		checkResult, err := NewService(env).CheckMany(r.Context(), &checkManySpec)
		if err != nil {
			return err
		}

		service.SendJSONResponse(w, checkResult)
		return nil
	}

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
