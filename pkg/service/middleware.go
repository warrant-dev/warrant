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

package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
)

const (
	paramNameLimit           = "limit"
	paramNamePage            = "page"
	paramNameQuery           = "q"
	paramNameSortBy          = "sortBy"
	paramNameSortOrder       = "sortOrder"
	paramNameAfterId         = "afterId"
	paramNameBeforeId        = "beforeId"
	paramNameAfterValue      = "afterValue"
	paramNameBeforeValue     = "beforeValue"
	defaultLimit             = 25
	defaultPage              = 1
	contextKeyListParams key = iota

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
	Page          int
	Limit         int
	Query         *string
	SortBy        string
	SortOrder     SortOrder
	AfterId       *string
	BeforeId      *string
	AfterValue    interface{}
	BeforeValue   interface{}
	defaultSortBy string
}

func (lp ListParams) String() string {
	s := fmt.Sprintf("page=%d&limit=%d&sortBy=%s&sortOrder=%d&defaultSortBy=%s", lp.Page, lp.Limit, lp.SortBy, lp.SortOrder, lp.defaultSortBy)
	if lp.Query != nil {
		s = s + "&q=" + *lp.Query
	}
	if lp.AfterId != nil {
		s = s + "&afterId=" + *lp.AfterId
	}
	if lp.BeforeId != nil {
		s = s + "&beforeId=" + *lp.BeforeId
	}
	if lp.AfterValue != nil {
		s = fmt.Sprintf("%s&afterValue=%v", s, lp.AfterValue)
	}
	if lp.BeforeValue != nil {
		s = fmt.Sprintf("%s&beforeValue=%v", s, lp.BeforeValue)
	}

	return s
}

func (lp ListParams) DefaultSortBy() string {
	return lp.defaultSortBy
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
		listParams := DefaultListParams(listParamParser)
		urlQueryParams := r.URL.Query()

		if urlQueryParams.Has(paramNameQuery) {
			query := urlQueryParams.Get(paramNameQuery)
			listParams.Query = &query
		}

		if urlQueryParams.Has(paramNamePage) {
			page, err := ParsePage(urlQueryParams.Get(paramNamePage))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNamePage, err.Error()))
				return
			}
			listParams.Page = page
		}

		if urlQueryParams.Has(paramNameLimit) {
			limit, err := ParseLimit(urlQueryParams.Get(paramNameLimit))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameLimit, err.Error()))
				return
			}
			listParams.Limit = limit
		}

		if urlQueryParams.Has(paramNameSortOrder) {
			sortOrder, err := ParseSortOrder(urlQueryParams.Get(paramNameSortOrder))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameSortOrder, err.Error()))
				return
			}
			listParams.SortOrder = sortOrder
		}

		var sortBy string
		if urlQueryParams.Has(paramNameSortBy) {
			sortBy, err = ParseSortBy(urlQueryParams.Get(paramNameSortBy), listParamParser)
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameSortBy, err.Error()))
				return
			}
			listParams.SortBy = sortBy
		}

		if urlQueryParams.Has(paramNameAfterValue) && !urlQueryParams.Has(paramNameAfterId) {
			SendErrorResponse(w, NewMissingRequiredParameterError(paramNameAfterId))
			return
		}

		if urlQueryParams.Has(paramNameBeforeValue) && !urlQueryParams.Has(paramNameBeforeId) {
			SendErrorResponse(w, NewMissingRequiredParameterError(paramNameBeforeId))
			return
		}

		if (urlQueryParams.Has(paramNameBeforeId) || urlQueryParams.Has(paramNameBeforeValue)) && (urlQueryParams.Has(paramNameAfterId) || urlQueryParams.Has(paramNameAfterValue)) {
			SendErrorResponse(w, NewInvalidRequestError(fmt.Sprintf("cannot pass %s and/or %s with %s and/or %s", paramNameBeforeId, paramNameBeforeValue, paramNameAfterId, paramNameAfterValue)))
			return
		}

		if (urlQueryParams.Has(paramNameAfterValue) || urlQueryParams.Has(paramNameBeforeValue)) && sortBy == listParams.DefaultSortBy() {
			SendErrorResponse(w, NewInvalidRequestError(fmt.Sprintf("cannot pass %s or %s when sorting by %s", paramNameAfterValue, paramNameBeforeValue, listParams.DefaultSortBy())))
			return
		}

		if urlQueryParams.Has(paramNameAfterId) {
			afterId, err := ParseId(urlQueryParams.Get(paramNameAfterId))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameAfterId, err.Error()))
				return
			}
			listParams.AfterId = &afterId
		}

		if urlQueryParams.Has(paramNameBeforeId) {
			beforeId, err := ParseId(urlQueryParams.Get(paramNameBeforeId))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameBeforeId, err.Error()))
				return
			}
			listParams.BeforeId = &beforeId
		}

		if urlQueryParams.Has(paramNameAfterValue) {
			afterValue, err := ParseValue(urlQueryParams.Get(paramNameAfterValue), sortBy, listParamParser)
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameAfterValue, err.Error()))
				return
			}
			listParams.AfterValue = afterValue
		}

		if urlQueryParams.Has(paramNameBeforeValue) {
			beforeValue, err := ParseValue(urlQueryParams.Get(paramNameBeforeValue), sortBy, listParamParser)
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameBeforeValue, err.Error()))
				return
			}
			listParams.BeforeValue = beforeValue
		}

		ctx := context.WithValue(r.Context(), contextKeyListParams, &listParams)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetListParamsFromContext[T ListParamParser](ctx context.Context) ListParams {
	ctxListParams := ctx.Value(contextKeyListParams)
	if ctxListParams != nil {
		listParams, ok := ctxListParams.(*ListParams)
		if !ok {
			log.Ctx(ctx).Error().Msg("service: unsuccessful type cast of listParams context value to *ListParams type")
			return DefaultListParams(ListParamParser(*new(T)))
		}

		return *listParams
	}

	return DefaultListParams(ListParamParser(*new(T)))
}

func DefaultListParams(listParamParser ListParamParser) ListParams {
	return ListParams{
		Page:          defaultPage,
		Limit:         defaultLimit,
		SortBy:        listParamParser.GetDefaultSortBy(),
		SortOrder:     SortOrderAsc,
		defaultSortBy: listParamParser.GetDefaultSortBy(),
	}
}
