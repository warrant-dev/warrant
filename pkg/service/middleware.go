package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

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

// Middleware defines the type of all middleware
type Middleware func(http.Handler) http.Handler

// ChainMiddleware a top-level middleware which applies the given middlewares in order from inner to outer (order of execution)
func ChainMiddleware(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

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
	Query       *string
	SortBy      string
	SortOrder   SortOrder
	AfterId     *string
	BeforeId    *string
	AfterValue  interface{}
	BeforeValue interface{}
}

func (lp ListParams) UseCursorPagination() bool {
	return lp.AfterId != nil || lp.BeforeId != nil || lp.AfterValue != nil || lp.BeforeValue != nil
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
		var err error
		listParamParser := ListParamParser(*new(T))
		urlQueryParams := r.URL.Query()
		ctx := r.Context()

		if urlQueryParams.Has(paramNameQuery) {
			ctx = context.WithValue(ctx, contextKeyQuery, urlQueryParams.Get(paramNameQuery))
		}

		if urlQueryParams.Has(paramNamePage) {
			page, err := ParsePage(urlQueryParams.Get(paramNamePage))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNamePage, err.Error()))
				return
			}
			ctx = context.WithValue(r.Context(), contextKeyPage, page)
		}

		if urlQueryParams.Has(paramNameLimit) {
			limit, err := ParseLimit(urlQueryParams.Get(paramNameLimit))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameLimit, err.Error()))
				return
			}
			ctx = context.WithValue(ctx, contextKeyLimit, limit)
		}

		if urlQueryParams.Has(paramNameSortOrder) {
			sortOrder, err := ParseSortOrder(urlQueryParams.Get(paramNameSortOrder))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameSortOrder, err.Error()))
				return
			}
			ctx = context.WithValue(ctx, contextKeySortOrder, sortOrder)
		}

		var sortBy string
		if urlQueryParams.Has(paramNameSortBy) {
			sortBy, err = ParseSortBy(urlQueryParams.Get(paramNameSortBy), listParamParser)
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameSortBy, err.Error()))
				return
			}
			ctx = context.WithValue(ctx, contextKeySortBy, sortBy)
		}

		if urlQueryParams.Has(paramNameAfterValue) && !urlQueryParams.Has(paramNameAfterId) {
			SendErrorResponse(w, NewMissingRequiredParameterError(paramNameAfterId))
			return
		}

		if urlQueryParams.Has(paramNameBeforeValue) && !urlQueryParams.Has(paramNameBeforeId) {
			SendErrorResponse(w, NewMissingRequiredParameterError(paramNameBeforeId))
			return
		}

		defaultSortBy := listParamParser.GetDefaultSortBy()
		if (urlQueryParams.Has(paramNameAfterValue) || urlQueryParams.Has(paramNameBeforeValue)) && sortBy == defaultSortBy {
			SendErrorResponse(w, NewInvalidRequestError(fmt.Sprintf("cannot pass %s or %s when sorting by %s", paramNameAfterValue, paramNameBeforeValue, defaultSortBy)))
			return
		}

		if urlQueryParams.Has(paramNameAfterId) {
			afterId, err := ParseId(urlQueryParams.Get(paramNameAfterId))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameAfterId, err.Error()))
				return
			}
			ctx = context.WithValue(ctx, contextKeyAfterId, afterId)
		}

		if urlQueryParams.Has(paramNameBeforeId) {
			beforeId, err := ParseId(urlQueryParams.Get(paramNameBeforeId))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameBeforeId, err.Error()))
				return
			}
			ctx = context.WithValue(ctx, contextKeyBeforeId, beforeId)
		}

		if urlQueryParams.Has(paramNameAfterValue) {
			afterValue, err := ParseValue(urlQueryParams.Get(paramNameAfterValue), sortBy, listParamParser)
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameAfterValue, err.Error()))
				return
			}
			ctx = context.WithValue(ctx, contextKeyAfterValue, afterValue)
		}

		if urlQueryParams.Has(paramNameBeforeValue) {
			beforeValue, err := ParseValue(urlQueryParams.Get(paramNameBeforeValue), sortBy, listParamParser)
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameBeforeValue, err.Error()))
				return
			}
			ctx = context.WithValue(ctx, contextKeyBeforeValue, beforeValue)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetListParamsFromContext[T ListParamParser](context context.Context) ListParams {
	var listParams ListParams
	listParamParser := ListParamParser(*new(T))

	contextPage := context.Value(contextKeyPage)
	if contextPage == nil {
		contextPage = defaultPage
	}
	listParams.Page = contextPage.(int)

	contextLimit := context.Value(contextKeyLimit)
	if contextLimit == nil {
		contextLimit = defaultLimit
	}
	listParams.Limit = contextLimit.(int)

	contextSortBy := context.Value(contextKeySortBy)
	if contextSortBy == nil {
		contextSortBy = listParamParser.GetDefaultSortBy()
	}
	listParams.SortBy = contextSortBy.(string)

	contextSortOrder := context.Value(contextKeySortOrder)
	if contextSortOrder == nil {
		contextSortOrder = SortOrderAsc
	}
	listParams.SortOrder = contextSortOrder.(SortOrder)

	contextQuery := context.Value(contextKeyQuery)
	if contextQuery != nil {
		query := contextQuery.(string)
		listParams.Query = &query
	}

	contextAfterId := context.Value(contextKeyAfterId)
	if contextAfterId != nil {
		afterId := contextAfterId.(string)
		listParams.AfterId = &afterId
	}

	contextBeforeId := context.Value(contextKeyBeforeId)
	if contextBeforeId != nil {
		beforeId := contextBeforeId.(string)
		listParams.BeforeId = &beforeId
	}

	contextAfterValue := context.Value(contextKeyAfterValue)
	if contextAfterValue != nil {
		listParams.AfterValue = contextAfterValue
	}

	contextBeforeValue := context.Value(contextKeyBeforeValue)
	if contextBeforeValue != nil {
		listParams.BeforeValue = contextBeforeValue
	}

	return listParams
}
