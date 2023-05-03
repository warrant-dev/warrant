package event

import (
	"net/http"
	"strconv"
	"time"

	"github.com/warrant-dev/warrant/pkg/service"
)

const (
	DateFormatMessage = "Must be in the format YYYY-MM-DD"
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
		since, err := time.Parse(DateFormat, DateFormat)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamSince, DateFormatMessage)
		}

		listParams.Since = since.UTC()
	} else {
		since, err := time.Parse(DateFormat, sinceString)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamSince, DateFormatMessage)
		}

		listParams.Since = since.UTC()
	}

	untilString := queryParams.Get(QueryParamUntil)
	if untilString == "" {
		year, month, day := time.Now().Date()
		listParams.Until = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	} else {
		until, err := time.Parse(DateFormat, untilString)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamUntil, DateFormatMessage)
		}

		listParams.Until = until.Add(24 * time.Hour).Add(-1 * time.Microsecond).UTC()
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
		since, err := time.Parse(DateFormat, DateFormat)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamSince, DateFormatMessage)
		}

		listParams.Since = since.UTC()
	} else {
		since, err := time.Parse(DateFormat, sinceString)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamSince, DateFormatMessage)
		}

		listParams.Since = since.UTC()
	}

	untilString := queryParams.Get(QueryParamUntil)
	if untilString == "" {
		year, month, day := time.Now().Date()
		listParams.Until = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	} else {
		until, err := time.Parse(DateFormat, untilString)
		if err != nil {
			return service.NewInvalidParameterError(QueryParamUntil, DateFormatMessage)
		}

		listParams.Until = until.Add(24 * time.Hour).Add(-1 * time.Microsecond).UTC()
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
