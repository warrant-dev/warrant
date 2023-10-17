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

package event

import (
	"net/http"
	"strconv"
	"time"

	"github.com/warrant-dev/warrant/pkg/service"
)

const (
	DateFormatMessage = "Must be an integer specifying a unix timestamp in microseconds"
	SinceErrorMessage = "Must be a date occurring before the until date"
)

func (svc EventService) Routes() ([]service.Route, error) {
	return []service.Route{
		// list resource events
		service.WarrantRoute{
			Pattern: "/v1/resource-events",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listResourceEventsV1),
				service.ListMiddleware[ResourceEventListParamParser],
			),
		},
		service.WarrantRoute{
			Pattern: "/v2/resource-events",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listResourceEventsV2),
				service.ListMiddleware[ResourceEventListParamParser],
			),
		},

		// list access events
		service.WarrantRoute{
			Pattern: "/v1/access-events",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listAccessEventsV1),
				service.ListMiddleware[AccessEventListParamParser],
			),
		},
		service.WarrantRoute{
			Pattern: "/v2/access-events",
			Method:  "GET",
			Handler: service.ChainMiddleware(
				service.NewRouteHandler(svc, listAccessEventsV2),
				service.ListMiddleware[AccessEventListParamParser],
			),
		},
	}, nil
}

func listResourceEventsV1(svc EventService, w http.ResponseWriter, r *http.Request) error {
	queryParams := r.URL.Query()
	listParams := service.GetListParamsFromContext[ResourceEventListParamParser](r.Context())
	filterParams := ResourceEventFilterParams{
		Type:         queryParams.Get(QueryParamType),
		Source:       queryParams.Get(QueryParamSource),
		ResourceType: queryParams.Get(QueryParamResourceType),
		ResourceId:   queryParams.Get(QueryParamResourceId),
	}

	var err error
	sinceString := queryParams.Get(QueryParamSince)
	if sinceString == "" {
		filterParams.Since = time.UnixMicro(DefaultEpochMicroseconds).UTC()
	} else {
		sinceMicroseconds, err := strconv.ParseInt(sinceString, 10, 64)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamSince, DateFormatMessage)
		}

		filterParams.Since = time.UnixMicro(sinceMicroseconds).UTC()
	}

	untilString := queryParams.Get(QueryParamUntil)
	if untilString == "" {
		filterParams.Until = time.Now().UTC()
	} else {
		untilMicroseconds, err := strconv.ParseInt(untilString, 10, 64)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamUntil, DateFormatMessage)
		}

		filterParams.Until = time.UnixMicro(untilMicroseconds).UTC()
	}

	if filterParams.Since.After(filterParams.Until) {
		return service.NewInvalidParameterError(QueryParamSince, SinceErrorMessage)
	}

	// create next cursor from lastId param
	if r.URL.Query().Has("lastId") {
		lastIdCursor, err := service.NewCursorFromBase64String(r.URL.Query().Get("lastId"))
		if err != nil {
			return service.NewInvalidParameterError("lastId", "invalid lastId")
		}

		listParams.NextCursor = lastIdCursor
	}

	resourceEventSpecs, _, nextCursor, err := svc.ListResourceEvents(r.Context(), filterParams, listParams)
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

	service.SendJSONResponse(w, ListEventsSpecV1[ResourceEventSpec]{
		Events: resourceEventSpecs,
		LastId: newLastId,
	})
	return nil
}

func listResourceEventsV2(svc EventService, w http.ResponseWriter, r *http.Request) error {
	queryParams := r.URL.Query()
	listParams := service.GetListParamsFromContext[ResourceEventListParamParser](r.Context())
	filterParams := ResourceEventFilterParams{
		Type:         queryParams.Get(QueryParamType),
		Source:       queryParams.Get(QueryParamSource),
		ResourceType: queryParams.Get(QueryParamResourceType),
		ResourceId:   queryParams.Get(QueryParamResourceId),
	}

	var err error
	sinceString := queryParams.Get(QueryParamSince)
	if sinceString == "" {
		filterParams.Since = time.UnixMicro(DefaultEpochMicroseconds).UTC()
	} else {
		sinceMicroseconds, err := strconv.ParseInt(sinceString, 10, 64)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamSince, DateFormatMessage)
		}

		filterParams.Since = time.UnixMicro(sinceMicroseconds).UTC()
	}

	untilString := queryParams.Get(QueryParamUntil)
	if untilString == "" {
		filterParams.Until = time.Now().UTC()
	} else {
		untilMicroseconds, err := strconv.ParseInt(untilString, 10, 64)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamUntil, DateFormatMessage)
		}

		filterParams.Until = time.UnixMicro(untilMicroseconds).UTC()
	}

	if filterParams.Since.After(filterParams.Until) {
		return service.NewInvalidParameterError(QueryParamSince, SinceErrorMessage)
	}

	// create next cursor from lastId param
	if r.URL.Query().Has("lastId") {
		lastIdCursor, err := service.NewCursorFromBase64String(r.URL.Query().Get("lastId"))
		if err != nil {
			return service.NewInvalidParameterError("lastId", "invalid lastId")
		}

		listParams.NextCursor = lastIdCursor
	}

	resourceEventSpecs, prevCursor, nextCursor, err := svc.ListResourceEvents(r.Context(), filterParams, listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, ListEventsSpecV2[ResourceEventSpec]{
		Results:    resourceEventSpecs,
		PrevCursor: prevCursor,
		NextCursor: nextCursor,
	})
	return nil
}

func listAccessEventsV1(svc EventService, w http.ResponseWriter, r *http.Request) error {
	queryParams := r.URL.Query()
	listParams := service.GetListParamsFromContext[AccessEventListParamParser](r.Context())
	filterParams := AccessEventFilterParams{
		Type:            queryParams.Get(QueryParamType),
		Source:          queryParams.Get(QueryParamSource),
		ObjectType:      queryParams.Get(QueryParamObjectType),
		ObjectId:        queryParams.Get(QueryParamObjectId),
		Relation:        queryParams.Get(QueryParamRelation),
		SubjectType:     queryParams.Get(QueryParamSubjectType),
		SubjectId:       queryParams.Get(QueryParamSubjectId),
		SubjectRelation: queryParams.Get(QueryParamSubjectRelation),
	}

	var err error
	sinceString := queryParams.Get(QueryParamSince)
	if sinceString == "" {
		filterParams.Since = time.UnixMicro(DefaultEpochMicroseconds).UTC()
	} else {
		sinceMicroseconds, err := strconv.ParseInt(sinceString, 10, 64)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamSince, DateFormatMessage)
		}

		filterParams.Since = time.UnixMicro(sinceMicroseconds).UTC()
	}

	untilString := queryParams.Get(QueryParamUntil)
	if untilString == "" {
		filterParams.Until = time.Now().UTC()
	} else {
		untilMicroseconds, err := strconv.ParseInt(untilString, 10, 64)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamUntil, DateFormatMessage)
		}

		filterParams.Until = time.UnixMicro(untilMicroseconds).UTC()
	}

	if filterParams.Since.After(filterParams.Until) {
		return service.NewInvalidParameterError(QueryParamSince, SinceErrorMessage)
	}

	// create next cursor from lastId param
	if r.URL.Query().Has("lastId") {
		lastIdCursor, err := service.NewCursorFromBase64String(r.URL.Query().Get("lastId"))
		if err != nil {
			return service.NewInvalidParameterError("lastId", "invalid lastId")
		}

		listParams.NextCursor = lastIdCursor
	}

	accessEventSpecs, _, nextCursor, err := svc.ListAccessEvents(r.Context(), filterParams, listParams)
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

	service.SendJSONResponse(w, ListEventsSpecV1[AccessEventSpec]{
		Events: accessEventSpecs,
		LastId: newLastId,
	})
	return nil
}

func listAccessEventsV2(svc EventService, w http.ResponseWriter, r *http.Request) error {
	queryParams := r.URL.Query()
	listParams := service.GetListParamsFromContext[AccessEventListParamParser](r.Context())
	filterParams := AccessEventFilterParams{
		Type:            queryParams.Get(QueryParamType),
		Source:          queryParams.Get(QueryParamSource),
		ObjectType:      queryParams.Get(QueryParamObjectType),
		ObjectId:        queryParams.Get(QueryParamObjectId),
		Relation:        queryParams.Get(QueryParamRelation),
		SubjectType:     queryParams.Get(QueryParamSubjectType),
		SubjectId:       queryParams.Get(QueryParamSubjectId),
		SubjectRelation: queryParams.Get(QueryParamSubjectRelation),
	}

	var err error
	sinceString := queryParams.Get(QueryParamSince)
	if sinceString == "" {
		filterParams.Since = time.UnixMicro(DefaultEpochMicroseconds).UTC()
	} else {
		sinceMicroseconds, err := strconv.ParseInt(sinceString, 10, 64)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamSince, DateFormatMessage)
		}

		filterParams.Since = time.UnixMicro(sinceMicroseconds).UTC()
	}

	untilString := queryParams.Get(QueryParamUntil)
	if untilString == "" {
		filterParams.Until = time.Now().UTC()
	} else {
		untilMicroseconds, err := strconv.ParseInt(untilString, 10, 64)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamUntil, DateFormatMessage)
		}

		filterParams.Until = time.UnixMicro(untilMicroseconds).UTC()
	}

	if filterParams.Since.After(filterParams.Until) {
		return service.NewInvalidParameterError(QueryParamSince, SinceErrorMessage)
	}

	// create next cursor from lastId param
	if r.URL.Query().Has("lastId") {
		lastIdCursor, err := service.NewCursorFromBase64String(r.URL.Query().Get("lastId"))
		if err != nil {
			return service.NewInvalidParameterError("lastId", "invalid lastId")
		}

		listParams.NextCursor = lastIdCursor
	}

	accessEventSpecs, prevCursor, nextCursor, err := svc.ListAccessEvents(r.Context(), filterParams, listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, ListEventsSpecV2[AccessEventSpec]{
		Results:    accessEventSpecs,
		PrevCursor: prevCursor,
		NextCursor: nextCursor,
	})
	return nil
}
