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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
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
	paramNameNextCursor      = "nextCursor"
	paramNamePrevCursor      = "prevCursor"
	defaultLimit             = 25
	defaultPage              = 1
	contextKeyListParams key = iota

	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)

type SortOrder string

func (so SortOrder) String() string {
	if so == SortOrderAsc || so == SortOrderDesc {
		return string(so)
	}

	return ""
}

type Cursor struct {
	id    string
	value interface{}
}

func (cursor Cursor) ID() string {
	return cursor.id
}

func (cursor Cursor) Value() interface{} {
	return cursor.value
}

func (cursor Cursor) String() string {
	if cursor.value == nil {
		return fmt.Sprintf("{id=%s}", cursor.id)
	}

	return fmt.Sprintf("{id=%s,value=%v}", cursor.id, cursor.value)
}

func (cursor Cursor) ToBase64String() (string, error) {
	var jsonStr string

	idJsonStr, err := json.Marshal(cursor.id)
	if err != nil {
		return "", errors.Wrapf(err, "error marshaling cursor %v", cursor)
	}
	jsonStr = fmt.Sprintf(`"id":%s`, idJsonStr)

	if cursor.value != nil {
		valJsonStr, err := json.Marshal(cursor.value)
		if err != nil {
			return "", errors.Wrapf(err, "error marshaling cursor %v", cursor)
		}
		jsonStr = fmt.Sprintf(`%s,"value":%s`, jsonStr, valJsonStr)
	}

	jsonStr = fmt.Sprintf(`{%s}`, jsonStr)
	return base64.StdEncoding.EncodeToString([]byte(jsonStr)), nil
}

func (cursor Cursor) MarshalJSON() ([]byte, error) {
	base64Str, err := cursor.ToBase64String()
	if err != nil {
		return nil, err
	}

	jsonStr := fmt.Sprintf(`"%s"`, base64Str)
	return []byte(jsonStr), nil
}

func NewCursor(id string, value interface{}) *Cursor {
	return &Cursor{
		id:    id,
		value: value,
	}
}

func NewCursorFromBase64String(base64CursorStr string, listParamParser ListParamParser, sortBy string) (*Cursor, error) {
	var cursor Cursor
	jsonStr, err := base64.StdEncoding.DecodeString(base64CursorStr)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid cursor %s", base64CursorStr)
	}

	var m map[string]string
	err = json.Unmarshal(jsonStr, &m)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("invalid cursor %s", base64CursorStr))
	}

	if id, exists := m["id"]; exists {
		cursor.id = id
	} else {
		return nil, errors.New(fmt.Sprintf("invalid cursor %s", base64CursorStr))
	}

	if rawValue, exists := m["value"]; exists {
		cursor.value, err = listParamParser.ParseValue(rawValue, sortBy)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("invalid cursor %s", base64CursorStr))
		}
	}

	return &cursor, nil
}

type ListParamParser interface {
	GetDefaultSortBy() string
	GetSupportedSortBys() []string
	ParseValue(val string, sortBy string) (interface{}, error)
}

type ListParams struct {
	Page          int       `json:"-"`
	Limit         int       `json:"limit,omitempty"`
	Query         *string   `json:"q,omitempty"`
	SortBy        string    `json:"sortBy,omitempty"`
	SortOrder     SortOrder `json:"sortOrder,omitempty"`
	PrevCursor    *Cursor   `json:"prevCursor,omitempty"`
	NextCursor    *Cursor   `json:"nextCursor,omitempty"`
	defaultSortBy string
}

func (lp *ListParams) WithPage(page int) {
	lp.Page = page
}

func (lp *ListParams) WithLimit(limit int) {
	lp.Limit = limit
}

func (lp *ListParams) WithQuery(query *string) {
	lp.Query = query
}

func (lp *ListParams) WithSortBy(sortBy string) {
	lp.SortBy = sortBy
}

func (lp *ListParams) WithSortOrder(sortOrder SortOrder) {
	lp.SortOrder = sortOrder
}

func (lp *ListParams) WithPrevCursor(prevCursor *Cursor) {
	lp.PrevCursor = prevCursor
}

func (lp *ListParams) WithNextCursor(nextCursor *Cursor) {
	lp.NextCursor = nextCursor
}

func (lp *ListParams) String() string {
	s := fmt.Sprintf("page=%d&limit=%d&sortBy=%s&sortOrder=%s&defaultSortBy=%s",
		lp.Page,
		lp.Limit,
		lp.SortBy,
		lp.SortOrder,
		lp.defaultSortBy,
	)

	if lp.Query != nil {
		s = fmt.Sprintf("%s&q=%s", s, *lp.Query)
	}
	if lp.NextCursor != nil {
		s = fmt.Sprintf("%s&nextCursor=%s", s, lp.NextCursor)
	}
	if lp.PrevCursor != nil {
		s = fmt.Sprintf("%s&bprevCursor=%s", s, lp.PrevCursor)
	}

	return s
}

func (lp *ListParams) DefaultSortBy() string {
	return lp.defaultSortBy
}

func (lp *ListParams) UseCursorPagination() bool {
	return lp.NextCursor != nil || lp.PrevCursor != nil
}

type GetDefaultSortByFunc func() string
type GetSupportedSortBys func() []string

func parsePage(val string) (int, error) {
	if val == "" {
		return defaultPage, nil
	}

	page, err := strconv.Atoi(val)
	if err != nil || page < 1 {
		return 0, fmt.Errorf("must be an integer greater than 0")
	}

	return page, nil
}

func parseLimit(val string) (int, error) {
	if val == "" {
		return defaultLimit, nil
	}

	limit, err := strconv.Atoi(val)
	if err != nil || limit < 1 || limit > 10000 {
		return 0, fmt.Errorf("must be an integer greater than 0 and less than or equal to 10000")
	}

	return limit, nil
}

func parseSortBy(val string, listParamParser ListParamParser) (string, error) {
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

func parseSortOrder(val string) (SortOrder, error) {
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

func parseId(val string) (string, error) {
	return val, nil
}

func parseValue(val string, sortBy string, listParamParser ListParamParser) (interface{}, error) {
	value, err := listParamParser.ParseValue(val, sortBy)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func parseCursor(base64CursorStr string, sortBy string, listParamParser ListParamParser) (*Cursor, error) {
	var cursor Cursor
	jsonStr, err := base64.StdEncoding.DecodeString(base64CursorStr)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid cursor %s", base64CursorStr)
	}

	var m map[string]string
	err = json.Unmarshal(jsonStr, &m)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("invalid cursor %s", base64CursorStr))
	}

	if id, exists := m["id"]; exists {
		cursor.id = id
	} else {
		return nil, errors.New(fmt.Sprintf("invalid cursor %s", base64CursorStr))
	}

	if rawValue, exists := m["value"]; exists {
		value, err := parseValue(rawValue, sortBy, listParamParser)
		if err != nil {
			return nil, err
		}

		cursor.value = value
	}

	return &cursor, nil
}

func ListMiddleware[T ListParamParser](next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		listParamParser := ListParamParser(*new(T))
		listParams := DefaultListParams(listParamParser)
		urlQueryParams := r.URL.Query()

		if urlQueryParams.Has(paramNameQuery) {
			query := urlQueryParams.Get(paramNameQuery)
			listParams.WithQuery(&query)
		}

		if urlQueryParams.Has(paramNamePage) {
			page, err := parsePage(urlQueryParams.Get(paramNamePage))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNamePage, err.Error()))
				return
			}
			listParams.WithPage(page)
		}

		if urlQueryParams.Has(paramNameLimit) {
			limit, err := parseLimit(urlQueryParams.Get(paramNameLimit))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameLimit, err.Error()))
				return
			}
			listParams.WithLimit(limit)
		}

		if urlQueryParams.Has(paramNameSortOrder) {
			sortOrder, err := parseSortOrder(urlQueryParams.Get(paramNameSortOrder))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameSortOrder, err.Error()))
				return
			}
			listParams.WithSortOrder(sortOrder)
		}

		var sortBy string
		if urlQueryParams.Has(paramNameSortBy) {
			sortBy, err = parseSortBy(urlQueryParams.Get(paramNameSortBy), listParamParser)
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameSortBy, err.Error()))
				return
			}
			listParams.WithSortBy(sortBy)
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
			var nextCursor *Cursor
			afterId, err := parseId(urlQueryParams.Get(paramNameAfterId))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameAfterId, err.Error()))
				return
			}

			if urlQueryParams.Has(paramNameAfterValue) {
				afterValue, err := parseValue(urlQueryParams.Get(paramNameAfterValue), listParams.SortBy, listParamParser)
				if err != nil {
					SendErrorResponse(w, NewInvalidParameterError(paramNameAfterValue, err.Error()))
					return
				}
				nextCursor = NewCursor(afterId, afterValue)
			} else {
				nextCursor = NewCursor(afterId, nil)
			}

			listParams.WithNextCursor(nextCursor)
		}

		if urlQueryParams.Has(paramNameBeforeId) {
			var prevCursor *Cursor
			beforeId, err := parseId(urlQueryParams.Get(paramNameBeforeId))
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameBeforeId, err.Error()))
				return
			}

			if urlQueryParams.Has(paramNameBeforeValue) {
				beforeValue, err := parseValue(urlQueryParams.Get(paramNameBeforeValue), listParams.SortBy, listParamParser)
				if err != nil {
					SendErrorResponse(w, NewInvalidParameterError(paramNameBeforeValue, err.Error()))
					return
				}
				prevCursor = NewCursor(beforeId, beforeValue)
			} else {
				prevCursor = NewCursor(beforeId, nil)
			}

			listParams.WithPrevCursor(prevCursor)
		}

		if urlQueryParams.Has(paramNameNextCursor) && urlQueryParams.Has(paramNamePrevCursor) {
			SendErrorResponse(w, NewInvalidRequestError(fmt.Sprintf("cannot pass both %s and %s together", paramNameNextCursor, paramNamePrevCursor)))
			return
		}

		if urlQueryParams.Has(paramNameNextCursor) {
			nextCursor, err := parseCursor(urlQueryParams.Get(paramNameNextCursor), listParams.SortBy, listParamParser)
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNameNextCursor, err.Error()))
				return
			}
			listParams.WithNextCursor(nextCursor)
		}

		if urlQueryParams.Has(paramNamePrevCursor) {
			prevCursor, err := parseCursor(urlQueryParams.Get(paramNamePrevCursor), listParams.SortBy, listParamParser)
			if err != nil {
				SendErrorResponse(w, NewInvalidParameterError(paramNamePrevCursor, err.Error()))
				return
			}
			listParams.WithPrevCursor(prevCursor)
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
		NextCursor:    nil,
		PrevCursor:    nil,
		defaultSortBy: listParamParser.GetDefaultSortBy(),
	}
}
