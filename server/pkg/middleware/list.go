package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

type key uint8

const (
	paramNameLimit            = "limit"
	paramNamePage             = "page"
	paramNameQuery            = "q"
	paramNameSortBy           = "sortBy"
	paramNameSortOrder        = "sortOrder"
	paramNameAfterId          = "afterId"
	paramNameBeforeId         = "beforeId"
	paramNameAfterValue       = "afterValue"
	paramNameBeforeValue      = "beforeValue"
	defaultLimit              = 25
	defaultPage               = 1
	contextKeyLimit       key = iota
	contextKeyPage        key = iota
	contextKeyQuery       key = iota
	contextKeySortBy      key = iota
	contextKeySortOrder   key = iota
	contextKeyAfterId     key = iota
	contextKeyBeforeId    key = iota
	contextKeyAfterValue  key = iota
	contextKeyBeforeValue key = iota

	SortOrderAsc  SortOrder = iota
	SortOrderDesc SortOrder = iota
)

type SortOrder int

func (so SortOrder) String() string {
	if so == SortOrderAsc {
		return "ASC"
	}

	if so == SortOrderDesc {
		return "DESC"
	}

	return ""
}

type ListParamParser interface {
	GetDefaultSortBy() string
	GetSupportedSortBys() []string
	ParseValue(val string, sortBy string) (interface{}, error)
}

type ListParams struct {
	Page        int
	Limit       int
	Query       string
	SortBy      string
	SortOrder   SortOrder
	AfterId     string
	BeforeId    string
	AfterValue  interface{}
	BeforeValue interface{}
}

func (lp ListParams) UseCursorPagination() bool {
	return lp.AfterId != "" || lp.BeforeId != "" || lp.AfterValue != nil || lp.BeforeValue != nil
}

type GetDefaultSortByFunc func() string
type GetSupportedSortBys func() []string

func ParsePage(val string) (int, error) {
	if val == "" {
		return defaultPage, nil
	}

	page, err := strconv.Atoi(val)
	if err != nil || page < 1 {
		return 0, fmt.Errorf("must be an integer greater than 0")
	}

	return page, nil
}

func ParseLimit(val string) (int, error) {
	if val == "" {
		return defaultLimit, nil
	}

	limit, err := strconv.Atoi(val)
	if err != nil || limit < 1 || limit > 10000 {
		return 0, fmt.Errorf("must be an integer greater than 0 and less than or equal to 10000")
	}

	return limit, nil
}

func ParseSortBy(val string, listParamParser ListParamParser) (string, error) {
	sortBy := val
	if sortBy == "" {
		sortBy = listParamParser.GetDefaultSortBy()
	}

	for _, supportedSortBy := range listParamParser.GetSupportedSortBys() {
		if sortBy == supportedSortBy {
			return sortBy, nil
		}
	}

	return "", fmt.Errorf("unsupported sortBy")
}

func ParseSortOrder(val string) (SortOrder, error) {
	switch val {
	case "ASC":
		return SortOrderAsc, nil
	case "DESC":
		return SortOrderDesc, nil
	case "":
		return SortOrderAsc, nil
	default:
		return SortOrderAsc, fmt.Errorf("must be ASC or DESC")
	}
}

func ParseId(val string) (string, error) {
	return val, nil
}

func ParseValue(val string, sortBy string, listParamParser ListParamParser) (interface{}, error) {
	value, err := listParamParser.ParseValue(val, sortBy)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func ListMiddleware[T ListParamParser](next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		listParamParser := ListParamParser(*new(T))
		urlQueryParams := r.URL.Query()
		pageParam := urlQueryParams.Get(paramNamePage)
		limitParam := urlQueryParams.Get(paramNameLimit)
		query := urlQueryParams.Get(paramNameQuery)
		sortBy := urlQueryParams.Get(paramNameSortBy)
		sortOrderParam := urlQueryParams.Get(paramNameSortOrder)
		afterIdParam := urlQueryParams.Get(paramNameAfterId)
		beforeIdParam := urlQueryParams.Get(paramNameBeforeId)
		afterValueParam := urlQueryParams.Get(paramNameAfterValue)
		beforeValueParam := urlQueryParams.Get(paramNameBeforeValue)

		page, err := ParsePage(pageParam)
		if err != nil {
			service.SendErrorResponse(w, service.NewInvalidParameterError(paramNamePage, err.Error()))
			return
		}

		limit, err := ParseLimit(limitParam)
		if err != nil {
			service.SendErrorResponse(w, service.NewInvalidParameterError(paramNameLimit, err.Error()))
			return
		}

		sortOrder, err := ParseSortOrder(sortOrderParam)
		if err != nil {
			service.SendErrorResponse(w, service.NewInvalidParameterError(paramNameSortOrder, err.Error()))
			return
		}

		sortBy, err = ParseSortBy(sortBy, listParamParser)
		if err != nil {
			service.SendErrorResponse(w, service.NewInvalidParameterError(paramNameSortBy, err.Error()))
			return
		}

		afterId, err := ParseId(afterIdParam)
		if err != nil {
			service.SendErrorResponse(w, service.NewInvalidParameterError(paramNameAfterId, err.Error()))
			return
		}

		beforeId, err := ParseId(beforeIdParam)
		if err != nil {
			service.SendErrorResponse(w, service.NewInvalidParameterError(paramNameBeforeId, err.Error()))
			return
		}

		if urlQueryParams.Has(paramNameAfterValue) && !urlQueryParams.Has(paramNameAfterId) {
			service.SendErrorResponse(w, service.NewMissingRequiredParameterError(paramNameAfterId))
			return
		}

		if urlQueryParams.Has(paramNameBeforeValue) && !urlQueryParams.Has(paramNameBeforeId) {
			service.SendErrorResponse(w, service.NewMissingRequiredParameterError(paramNameBeforeId))
			return
		}

		defaultSortBy := listParamParser.GetDefaultSortBy()
		if !urlQueryParams.Has(paramNameAfterValue) && sortBy != defaultSortBy && urlQueryParams.Has(paramNameAfterId) {
			service.SendErrorResponse(w, service.NewMissingRequiredParameterError(paramNameAfterValue))
			return
		}

		if !urlQueryParams.Has(paramNameBeforeValue) && sortBy != defaultSortBy && urlQueryParams.Has(paramNameBeforeId) {
			service.SendErrorResponse(w, service.NewMissingRequiredParameterError(paramNameBeforeValue))
			return
		}

		if (urlQueryParams.Has(paramNameAfterValue) || urlQueryParams.Has(paramNameBeforeValue)) && sortBy == defaultSortBy {
			service.SendErrorResponse(w, service.NewInvalidRequestError(fmt.Sprintf("cannot pass %s or %s when sorting by %s", paramNameAfterValue, paramNameBeforeValue, defaultSortBy)))
			return
		}

		var afterValue interface{} = nil
		if urlQueryParams.Has(paramNameAfterValue) {
			afterValue, err = ParseValue(afterValueParam, sortBy, listParamParser)
			if err != nil {
				service.SendErrorResponse(w, service.NewInvalidParameterError(paramNameAfterValue, err.Error()))
				return
			}
		}

		var beforeValue interface{} = nil
		if urlQueryParams.Has(paramNameBeforeValue) {
			beforeValue, err = ParseValue(beforeValueParam, sortBy, listParamParser)
			if err != nil {
				service.SendErrorResponse(w, service.NewInvalidParameterError(paramNameBeforeValue, err.Error()))
				return
			}
		}

		ctx := context.WithValue(r.Context(), contextKeyPage, page)
		ctx = context.WithValue(ctx, contextKeyLimit, limit)
		ctx = context.WithValue(ctx, contextKeyQuery, query)
		ctx = context.WithValue(ctx, contextKeySortBy, sortBy)
		ctx = context.WithValue(ctx, contextKeySortOrder, sortOrder)
		ctx = context.WithValue(ctx, contextKeyAfterId, afterId)
		ctx = context.WithValue(ctx, contextKeyBeforeId, beforeId)
		ctx = context.WithValue(ctx, contextKeyAfterValue, afterValue)
		ctx = context.WithValue(ctx, contextKeyBeforeValue, beforeValue)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetListParamsFromContext(context context.Context) ListParams {
	contextPage := context.Value(contextKeyPage)
	if contextPage == nil {
		log.Fatal().Msg("List context not available. Did you forget to add ListMiddleware to a handler?")
	}

	contextLimit := context.Value(contextKeyLimit)
	if contextLimit == nil {
		log.Fatal().Msg("List context not available. Did you forget to add ListMiddleware to a handler?")
	}

	contextSortBy := context.Value(contextKeySortBy)
	if contextSortBy == nil {
		log.Fatal().Msg("List context not available. Did you forget to add ListMiddleware to a handler?")
	}

	contextSortOrder := context.Value(contextKeySortOrder)
	if contextSortOrder == nil {
		log.Fatal().Msg("List context not available. Did you forget to add ListMiddleware to a handler?")
	}

	contextQuery := context.Value(contextKeyQuery)
	contextAfterId := context.Value(contextKeyAfterId)
	contextBeforeId := context.Value(contextKeyBeforeId)
	contextAfterValue := context.Value(contextKeyAfterValue)
	contextBeforeValue := context.Value(contextKeyBeforeValue)

	return ListParams{
		Page:        contextPage.(int),
		Limit:       contextLimit.(int),
		Query:       contextQuery.(string),
		SortBy:      contextSortBy.(string),
		SortOrder:   contextSortOrder.(SortOrder),
		AfterId:     contextAfterId.(string),
		BeforeId:    contextBeforeId.(string),
		AfterValue:  contextAfterValue,
		BeforeValue: contextBeforeValue,
	}
}
