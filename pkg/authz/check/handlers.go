// Copyright 2023 Forerunner Labs, Inc.
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
	"net/http"

	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/service"
)

func (svc CheckService) Routes() ([]service.Route, error) {
	return []service.Route{
		// Standard Authorization
		service.WarrantRoute{
			Pattern: "/v2/authorize",
			Method:  "POST",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, AuthorizeHandler),
				wookie.ClientTokenMiddleware,
			),
			OverrideAuthMiddlewareFunc: service.ApiKeyAndSessionAuthMiddleware,
		},
	}, nil
}

func AuthorizeHandler(svc CheckService, w http.ResponseWriter, r *http.Request) error {
	authInfo := service.GetAuthInfoFromRequestContext(r.Context())
	if authInfo != nil && authInfo.UserId != "" {
		var sessionCheckManySpec SessionCheckManySpec
		err := service.ParseJSONBody(r.Body, &sessionCheckManySpec)
		if err != nil {
			return err
		}

		warrantSpecs := make([]CheckWarrantSpec, 0)
		for _, warrantSpec := range sessionCheckManySpec.Warrants {
			warrantSpecs = append(warrantSpecs, CheckWarrantSpec{
				ObjectType: warrantSpec.ObjectType,
				ObjectId:   warrantSpec.ObjectId,
				Relation:   warrantSpec.Relation,
				Subject: &warrant.SubjectSpec{
					ObjectType: objecttype.ObjectTypeUser,
					ObjectId:   authInfo.UserId,
				},
			})
		}

		checkManySpec := CheckManySpec{
			Op:       sessionCheckManySpec.Op,
			Warrants: warrantSpecs,
			Context:  sessionCheckManySpec.Context,
			Debug:    sessionCheckManySpec.Debug,
		}

		checkResult, updatedWookie, err := svc.CheckMany(r.Context(), authInfo, &checkManySpec)
		if err != nil {
			return err
		}
		wookie.AddAsResponseHeader(w, updatedWookie)

		service.SendJSONResponse(w, checkResult)
		return nil
	}

	var checkManySpec CheckManySpec
	err := service.ParseJSONBody(r.Body, &checkManySpec)
	if err != nil {
		return err
	}

	checkResult, updatedWookie, err := svc.CheckMany(r.Context(), authInfo, &checkManySpec)
	if err != nil {
		return err
	}
	wookie.AddAsResponseHeader(w, updatedWookie)

	service.SendJSONResponse(w, checkResult)
	return nil
}
