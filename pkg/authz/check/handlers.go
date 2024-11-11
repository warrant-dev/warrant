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
	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/authz/adaptor"
	"net/http"
	"strings"

	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/pkg/service"
)

func (svc CheckService) Routes() ([]service.Route, error) {
	return []service.Route{
		service.WarrantRoute{
			Pattern:                    "/v2/authorize",
			Method:                     "POST",
			Handler:                    service.NewRouteHandler(svc, authorizeHandler),
			OverrideAuthMiddlewareFunc: service.ApiKeyAndSessionAuthMiddleware,
		},
		service.WarrantRoute{
			Pattern:                    "/v2/check",
			Method:                     "POST",
			Handler:                    service.NewRouteHandler(svc, authorizeHandler),
			OverrideAuthMiddlewareFunc: service.ApiKeyAndSessionAuthMiddleware,
		},
		service.WarrantRoute{
			Pattern:                    "/v2/checkUser",
			Method:                     "POST",
			Handler:                    service.NewRouteHandler(svc, authorize4UserHandler),
			OverrideAuthMiddlewareFunc: service.ApiKeyAndSessionAuthMiddleware,
		},
	}, nil
}

func authorizeHandler(svc CheckService, w http.ResponseWriter, r *http.Request) error {
	authInfo, err := service.GetAuthInfoFromRequestContext(r.Context())
	if err != nil {
		return err
	}

	if authInfo != nil && authInfo.UserId != "" {
		var sessionCheckManySpec SessionCheckManySpec
		err := service.ParseJSONBody(r.Context(), r.Body, &sessionCheckManySpec)
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

		checkResult, err := svc.CheckMany(r.Context(), authInfo, &checkManySpec)
		if err != nil {
			return err
		}

		service.SendJSONResponse(w, checkResult)
		return nil
	}

	var checkManySpec CheckManySpec
	err = service.ParseJSONBody(r.Context(), r.Body, &checkManySpec)
	if err != nil {
		return err
	}

	checkResult, err := svc.CheckMany(r.Context(), authInfo, &checkManySpec)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, checkResult)
	return nil
}

func authorize4UserHandler(svc CheckService, w http.ResponseWriter, r *http.Request) error {
	var checkUserSpec CheckUserSpec
	err := service.ParseJSONBody(r.Context(), r.Body, &checkUserSpec)
	if err != nil {
		return err
	}

	if checkUserSpec.BizType == BizTypeWorkspaceAppAccess {
		orgId, imGroupIds, err := adaptor.GetUserIds(checkUserSpec.UserId, true, true)
		if err != nil {
			return err
		}

		checkManySpec := CheckManySpec{
			Op:       objecttype.InheritIfAnyOf,
			Warrants: buildWarrantSpecs(checkUserSpec, orgId, imGroupIds),
			Context: warrant.PolicyContext{
				"orgId": orgId,
			},
			Debug: checkUserSpec.Debug,
		}

		checkResult, err := svc.CheckMany(r.Context(), nil, &checkManySpec)
		if err != nil {
			return err
		}

		service.SendJSONResponse(w, checkResult)
		return nil
	}

	return errors.New("unsupported bizType:" + string(checkUserSpec.BizType))
}

func buildWarrantSpecs(checkUserSpec CheckUserSpec, orgId string, imGroupIds []string) []CheckWarrantSpec {
	var objectId string
	if checkUserSpec.Resource.Id != "" {
		objectId = checkUserSpec.Resource.Id
	} else {
		//todo
		objectId = "xxxx"
	}

	warrantSpecs := make([]CheckWarrantSpec, 0)

	warrantSpecs = append(warrantSpecs, CheckWarrantSpec{
		ObjectType: checkUserSpec.Resource.ResType,
		ObjectId:   objectId,
		Relation:   checkUserSpec.Operate,
		Subject: &warrant.SubjectSpec{
			ObjectType: objecttype.ObjectTypeUser,
			ObjectId:   checkUserSpec.UserId,
		},
	})

	if len(strings.TrimSpace(orgId)) > 0 {
		warrantSpecs = append(warrantSpecs, CheckWarrantSpec{
			ObjectType: checkUserSpec.Resource.ResType,
			ObjectId:   objectId,
			Relation:   checkUserSpec.Operate,
			Subject: &warrant.SubjectSpec{
				ObjectType: objecttype.ObjectTypeOrg,
				ObjectId:   orgId,
			},
		})
	}

	if len(imGroupIds) > 0 {
		for _, imGroupId := range imGroupIds {
			warrantSpecs = append(warrantSpecs, CheckWarrantSpec{
				ObjectType: checkUserSpec.Resource.ResType,
				ObjectId:   objectId,
				Relation:   checkUserSpec.Operate,
				Subject: &warrant.SubjectSpec{
					ObjectType: objecttype.ObjectTypeImGroup,
					ObjectId:   imGroupId,
				},
			})
		}
	}
	return warrantSpecs
}
