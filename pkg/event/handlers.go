package event

import (
	"net/http"
	"strconv"
	"time"

	"github.com/warrant-dev/warrant/pkg/service"
)

const (
	DateFormatMessage = "Must be an integer specifying the epoch time in microseconds"
	SinceErrorMessage = "Must be a date occurring before the until date"
	LimitErrorMessage = "Must be an integer between 1 and 1000"
)

func (svc EventService) Routes() ([]service.Route, error) {
	return []service.Route{
		// get
		service.WarrantRoute{
			Pattern: "/v1/resource-events",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, ListResourceEvents),
		},

		service.WarrantRoute{
			Pattern: "/v1/access-events",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, ListAccessEvents),
		},
	}, nil
}

func ListResourceEvents(svc EventService, w http.ResponseWriter, r *http.Request) error {
	queryParams := r.URL.Query()
	listParams := ListResourceEventParams{
		Type:         queryParams.Get(QueryParamType),
		Source:       queryParams.Get(QueryParamSource),
		ResourceType: queryParams.Get(QueryParamResourceType),
		ResourceId:   queryParams.Get(QueryParamResourceId),
		LastId:       queryParams.Get(QueryParamLastId),
		Limit:        DefaultLimit,
	}

	var err error
	sinceString := queryParams.Get(QueryParamSince)
	if sinceString == "" {
		listParams.Since = time.UnixMicro(DefaultEpochMicroseconds).UTC()
	} else {
		sinceMicroseconds, err := strconv.ParseInt(sinceString, 10, 64)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamSince, DateFormatMessage)
		}

		listParams.Since = time.UnixMicro(sinceMicroseconds).UTC()
	}

	untilString := queryParams.Get(QueryParamUntil)
	if untilString == "" {
		listParams.Until = time.Now().UTC()
	} else {
		untilMicroseconds, err := strconv.ParseInt(untilString, 10, 64)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamUntil, DateFormatMessage)
		}

		listParams.Until = time.UnixMicro(untilMicroseconds).UTC()
	}

	if listParams.Since.After(listParams.Until) {
		return service.NewInvalidParameterError(QueryParamSince, SinceErrorMessage)
	}

	limitString := queryParams.Get(QueryParamLimit)
	if limitString == "" {
		listParams.Limit = DefaultLimit
	} else {
		listParams.Limit, err = strconv.ParseInt(limitString, 10, 64)
		if err != nil || listParams.Limit <= 0 || listParams.Limit > 1000 {
			return service.NewInvalidParameterError(QueryParamLimit, LimitErrorMessage)
		}
	}

	resourceEventSpecs, lastId, err := svc.ListResourceEvents(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, ListEventsSpec[ResourceEventSpec]{
		Events: resourceEventSpecs,
		LastId: lastId,
	})
	return nil
}

func ListAccessEvents(svc EventService, w http.ResponseWriter, r *http.Request) error {
	queryParams := r.URL.Query()
	listParams := ListAccessEventParams{
		Type:            queryParams.Get(QueryParamType),
		Source:          queryParams.Get(QueryParamSource),
		LastId:          queryParams.Get(QueryParamLastId),
		Limit:           DefaultLimit,
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
		listParams.Since = time.UnixMicro(DefaultEpochMicroseconds).UTC()
	} else {
		sinceMicroseconds, err := strconv.ParseInt(sinceString, 10, 64)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamSince, DateFormatMessage)
		}

		listParams.Since = time.UnixMicro(sinceMicroseconds).UTC()
	}

	untilString := queryParams.Get(QueryParamUntil)
	if untilString == "" {
		listParams.Until = time.Now().UTC()
	} else {
		untilMicroseconds, err := strconv.ParseInt(untilString, 10, 64)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamUntil, DateFormatMessage)
		}

		listParams.Until = time.UnixMicro(untilMicroseconds).UTC()
	}

	if listParams.Since.After(listParams.Until) {
		return service.NewInvalidParameterError(QueryParamSince, SinceErrorMessage)
	}

	limitString := queryParams.Get(QueryParamLimit)
	if limitString == "" {
		listParams.Limit = DefaultLimit
	} else {
		listParams.Limit, err = strconv.ParseInt(limitString, 10, 64)
		if err != nil || listParams.Limit <= 0 || listParams.Limit > 1000 {
			return service.NewInvalidParameterError(QueryParamLimit, LimitErrorMessage)
		}
	}

	accessEventSpecs, lastId, err := svc.ListAccessEvents(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, ListEventsSpec[AccessEventSpec]{
		Events: accessEventSpecs,
		LastId: lastId,
	})
	return nil
}
