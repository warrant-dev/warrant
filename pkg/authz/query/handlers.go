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

	"github.com/warrant-dev/warrant/pkg/service"
)

func (svc QueryService) Routes() ([]service.Route, error) {
	return []service.Route{
		service.WarrantRoute{
			Pattern: "/v1/query",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, QueryHandler),
				service.ListMiddleware[QueryListParamParser],
			),
		},
	}, nil
}

func QueryHandler(svc QueryService, w http.ResponseWriter, r *http.Request) error {
	queryString := r.URL.Query().Get("q")
	query, err := NewQueryFromString(queryString)
	if err != nil {
		return err
	}

	listParams := service.GetListParamsFromContext[QueryListParamParser](r.Context())
	result, err := svc.Query(r.Context(), query, listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, result)
	return nil
}
