// Copyright 2024 WorkOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authz

import (
	"context"
	"github.com/rs/zerolog/log"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	authz "github.com/warrant-dev/warrant/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/pkg/wookie"
	"net/http"

	"github.com/warrant-dev/warrant/pkg/service"
)

func (svc QueryService) Routes() ([]service.Route, error) {
	return []service.Route{
		service.WarrantRoute{
			Pattern: "/v1/query",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, queryV1),
				service.ListMiddleware[QueryListParamParser],
			),
		},
		service.WarrantRoute{
			Pattern: "/v2/query",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, queryV2),
				service.ListMiddleware[QueryListParamParser],
			),
		},
	}, nil
}

func queryV1(svc QueryService, w http.ResponseWriter, r *http.Request) error {
	queryParams := r.URL.Query()
	queryString := queryParams.Get("q")
	query, err := NewQueryFromString(queryString)
	if err != nil {
		return err
	}

	if queryParams.Has("context") {
		err = query.WithContext(queryParams.Get("context"))
		if err != nil {
			return service.NewInvalidParameterError("context", "invalid")
		}
	}

	listParams := service.GetListParamsFromContext[QueryListParamParser](r.Context())
	// create next cursor from lastId or afterId param
	if r.URL.Query().Has("lastId") {
		lastIdCursor, err := service.NewCursorFromBase64String(r.URL.Query().Get("lastId"), QueryListParamParser{}, listParams.SortBy)
		if err != nil {
			return service.NewInvalidParameterError("lastId", "invalid lastId")
		}

		listParams.WithNextCursor(lastIdCursor)
	} else if r.URL.Query().Has("afterId") {
		afterIdCursor, err := service.NewCursorFromBase64String(r.URL.Query().Get("afterId"), QueryListParamParser{}, listParams.SortBy)
		if err != nil {
			return service.NewInvalidParameterError("afterId", "invalid afterId")
		}

		listParams.WithNextCursor(afterIdCursor)
	}

	results, _, nextCursor, err := svc.Query(r.Context(), query, listParams)
	if err != nil {
		return err
	}

	var newLastId string
	if nextCursor != nil {
		base64EncodedNextCursor, err := nextCursor.ToBase64String()
		if err != nil {
			return err
		}
		newLastId = base64EncodedNextCursor
	}

	service.SendJSONResponse(w, QueryResponseV1{
		Results: results,
		LastId:  newLastId,
	})
	return nil
}

func queryV2(svc QueryService, w http.ResponseWriter, r *http.Request) error {
	queryParams := r.URL.Query()
	queryString := queryParams.Get("q")
	query, err := NewQueryFromString(queryString)
	if err != nil {
		return err
	}

	syncUserRelationsOnQuery(r.Context(), svc, query)

	if queryParams.Has("context") {
		err = query.WithContext(queryParams.Get("context"))
		if err != nil {
			return service.NewInvalidParameterError("context", "invalid")
		}
	}

	listParams := service.GetListParamsFromContext[QueryListParamParser](r.Context())
	results, prevCursor, nextCursor, err := svc.Query(r.Context(), query, listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, QueryResponseV2{
		Results:    results,
		PrevCursor: prevCursor,
		NextCursor: nextCursor,
	})
	return nil
}

func syncUserRelationsOnQuery(context context.Context, svc QueryService, query Query) {
	if query.SelectObjects == nil ||
		query.SelectObjects.WhereSubject == nil ||
		query.SelectObjects.WhereSubject.Type != objecttype.ObjectTypeUser ||
		query.SelectObjects.WhereSubject.Id == "" {
		return
	}
	userId := query.SelectObjects.WhereSubject.Id
	//orgId, imGroupIds, err := adaptor.GetUserIds(userId, true, true)
	//if err != nil {
	//	log.Error().Err(err).Msgf("syncUserRelationsOnQuery: cannot get user ids for user %s", userId)
	//	return
	//}
	orgId := context.Value(wookie.OrgIdKey).(string)
	log.Info().Msgf("syncUserRelationsOnQuery,uid:%s, orgId:%s", userId, orgId)

	if orgId != "" {
		_, _, err := svc.warrantSvc.Create(context, authz.CreateWarrantSpec{
			ObjectType: objecttype.ObjectTypeOrg,
			ObjectId:   orgId,
			Relation:   "member",
			Subject: &authz.SubjectSpec{
				ObjectType: objecttype.ObjectTypeUser,
				ObjectId:   userId,
			},
		})
		if err != nil {
			log.Error().Err(err).Msgf("syncUserRelationsOnQuery: cannot create warrant  for user %s and  orgId %s", userId, orgId)
		}
	}
	//if imGroupIds != nil && len(imGroupIds) > 0 {
	//	for _, groupId := range imGroupIds {
	//		_, _, err = svc.warrantSvc.Create(context, authz.CreateWarrantSpec{
	//			ObjectType: objecttype.ObjectTypeImGroup,
	//			ObjectId:   groupId,
	//			Relation:   "member",
	//			Subject: &authz.SubjectSpec{
	//				ObjectType: objecttype.ObjectTypeUser,
	//				ObjectId:   userId,
	//			},
	//		})
	//		if err != nil {
	//			log.Error().Err(err).Msgf("syncUserRelationsOnQuery: cannot create warrant  for user %s and  imGroupId %s", userId, groupId)
	//		}
	//	}
	//}
}
