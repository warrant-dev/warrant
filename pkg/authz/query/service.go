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
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	object "github.com/warrant-dev/warrant/pkg/object"
	"github.com/warrant-dev/warrant/pkg/service"
)

const (
	MaxObjectTypes = 1000
	MaxEdges       = 5000
)

var ErrInvalidQuery = errors.New("invalid query")

type QueryService struct {
	service.BaseService
	objectTypeSvc objecttype.Service
	warrantSvc    warrant.Service
	objectSvc     object.Service
}

func NewService(env service.Env, objectTypeSvc objecttype.Service, warrantSvc warrant.Service, objectSvc object.Service) QueryService {
	return QueryService{
		BaseService:   service.NewBaseService(env),
		objectTypeSvc: objectTypeSvc,
		warrantSvc:    warrantSvc,
		objectSvc:     objectSvc,
	}
}

func (svc QueryService) Query(ctx context.Context, query *Query, listParams service.ListParams) ([]QueryResult, *service.Cursor, *service.Cursor, error) {
	queryResults := make([]QueryResult, 0)
	resultMap := make(map[string]int)
	objects := make(map[string][]string)
	selectedObjectTypes := make(map[string]bool)

	if (query.SelectObjects == nil && query.SelectSubjects == nil) || (query.SelectObjects != nil && query.SelectSubjects != nil) {
		return nil, nil, nil, ErrInvalidQuery
	}

	resultSet, err := svc.query(ctx, query)
	if err != nil {
		return nil, nil, nil, err
	}

	for res := resultSet.List(); res != nil; res = res.Next() {
		var isImplicit bool
		//nolint:gocritic
		if query.SelectObjects != nil {
			isImplicit = !matches(query.SelectObjects.ObjectTypes, res.Warrant.ObjectType) || !matches(query.SelectObjects.Relations, res.Warrant.Relation)
		} else if query.SelectSubjects != nil {
			isImplicit = !matches(query.SelectSubjects.SubjectTypes, res.Warrant.Subject.ObjectType) || !matches(query.SelectSubjects.Relations, res.Warrant.Relation)
		} else {
			return nil, nil, nil, ErrInvalidQuery
		}

		queryResults = append(queryResults, QueryResult{
			ObjectType: res.ObjectType,
			ObjectId:   res.ObjectId,
			Warrant:    res.Warrant,
			IsImplicit: isImplicit,
		})
	}

	// handle sorting and pagination
	switch listParams.SortBy {
	case PrimarySortKey:
		switch listParams.SortOrder {
		case service.SortOrderAsc:
			sort.Sort(ByObjectTypeAndObjectIdAsc(queryResults))
		case service.SortOrderDesc:
			sort.Sort(ByObjectTypeAndObjectIdDesc(queryResults))
		}
	case "createdAt":
		switch listParams.SortOrder {
		case service.SortOrderAsc:
			sort.Sort(ByCreatedAtAsc(queryResults))
		case service.SortOrderDesc:
			sort.Sort(ByCreatedAtDesc(queryResults))
		}
	default:
		return nil, nil, nil, ErrInvalidQuery
	}

	var (
		prevCursor *service.Cursor
		nextCursor *service.Cursor
		start      int
		end        int
	)
	paginatedQueryResults := make([]QueryResult, 0)
	//nolint:gocritic
	if listParams.NextCursor != nil { // seek forward if NextCursor passed in
		lastObjectType, lastObjectId, err := objectTypeAndObjectIdFromCursor(listParams.NextCursor)
		if err != nil {
			return nil, nil, nil, service.NewInvalidParameterError("nextCursor", "invalid cursor")
		}

		start = 0
		for start < len(queryResults) && (queryResults[start].ObjectType != lastObjectType || queryResults[start].ObjectId != lastObjectId) {
			start++
		}

		end = start + listParams.Limit
	} else if listParams.PrevCursor != nil { // seek backward if PrevCursor passed in
		lastObjectType, lastObjectId, err := objectTypeAndObjectIdFromCursor(listParams.PrevCursor)
		if err != nil {
			return nil, nil, nil, service.NewInvalidParameterError("prevCursor", "invalid cursor")
		}

		end = len(queryResults) - 1
		for end > 0 && (queryResults[end].ObjectType != lastObjectType || queryResults[end].ObjectId != lastObjectId) {
			end--
		}

		start = end - listParams.Limit
	} else {
		start = 0
		end = start + listParams.Limit
	}

	// if there are more results backward
	if start > 0 {
		var value interface{} = nil
		switch listParams.SortBy {
		case PrimarySortKey:
			// do nothing
		case "createdAt":
			value = queryResults[start].Warrant.CreatedAt
		default:
			value = queryResults[start].Meta[listParams.SortBy]
		}

		prevCursor = service.NewCursor(objectKey(queryResults[start].ObjectType, queryResults[start].ObjectId), value)
	}

	// if there are more results forward
	if end < len(queryResults) {
		var value interface{} = nil
		switch listParams.SortBy {
		case PrimarySortKey:
			// do nothing
		case "createdAt":
			value = queryResults[end].Warrant.CreatedAt
		default:
			value = queryResults[end].Meta[listParams.SortBy]
		}

		nextCursor = service.NewCursor(objectKey(queryResults[end].ObjectType, queryResults[end].ObjectId), value)
	}

	for start < end && start < len(queryResults) {
		paginatedQueryResult := queryResults[start]
		paginatedQueryResults = append(paginatedQueryResults, paginatedQueryResult)
		selectedObjectTypes[paginatedQueryResult.ObjectType] = true
		objects[paginatedQueryResult.ObjectType] = append(objects[paginatedQueryResult.ObjectType], paginatedQueryResult.ObjectId)
		resultMap[objectKey(paginatedQueryResult.ObjectType, paginatedQueryResult.ObjectId)] = len(paginatedQueryResults) - 1
		start++
	}

	for selectedObjectType := range selectedObjectTypes {
		if len(objects[selectedObjectType]) > 0 {
			objectSpecs, err := svc.objectSvc.BatchGetByObjectTypeAndIds(ctx, selectedObjectType, objects[selectedObjectType])
			if err != nil {
				return nil, nil, nil, err
			}

			for _, objectSpec := range objectSpecs {
				paginatedQueryResults[resultMap[objectKey(selectedObjectType, objectSpec.ObjectId)]].Meta = objectSpec.Meta
			}
		}
	}

	return paginatedQueryResults, prevCursor, nextCursor, nil
}

func (svc QueryService) query(ctx context.Context, query *Query) (*ResultSet, error) {
	var objectTypes []string
	var selectSubjects bool
	objectTypesListParams := service.DefaultListParams(objecttype.ObjectTypeListParamParser{})
	objectTypesListParams.Limit = MaxObjectTypes
	typesList, _, _, err := svc.objectTypeSvc.List(ctx, objectTypesListParams)
	if err != nil {
		return nil, err
	}
	//nolint:gocritic
	if query.SelectObjects != nil {
		selectSubjects = false
		if query.SelectObjects.ObjectTypes[0] == warrant.Wildcard {
			for _, objectType := range typesList {
				objectTypes = append(objectTypes, objectType.Type)
			}
		} else {
			objectTypes = append(objectTypes, query.SelectObjects.ObjectTypes...)
		}
	} else if query.SelectSubjects != nil {
		selectSubjects = true
		if query.SelectSubjects.ForObject != nil {
			objectTypes = append(objectTypes, query.SelectSubjects.ForObject.Type)
		} else {
			for _, objectType := range typesList {
				objectTypes = append(objectTypes, objectType.Type)
			}
		}
	} else {
		return nil, ErrInvalidQuery
	}

	resultSet := NewResultSet()
	for _, objectType := range objectTypes {
		var relations []string
		objectTypeDef, err := svc.objectTypeSvc.GetByTypeId(ctx, objectType)
		if err != nil {
			return nil, err
		}

		if query.SelectObjects != nil {
			if query.SelectObjects.Relations[0] == warrant.Wildcard {
				for relation := range objectTypeDef.Relations {
					relations = append(relations, relation)
				}
			} else {
				relations = append(relations, query.SelectObjects.Relations...)
			}
		} else {
			if query.SelectSubjects.Relations[0] == warrant.Wildcard {
				for relation := range objectTypeDef.Relations {
					relations = append(relations, relation)
				}
			} else {
				relations = append(relations, query.SelectSubjects.Relations...)
			}
		}

		for _, relation := range relations {
			var matchFilters warrant.FilterParams
			if query.SelectObjects != nil && query.SelectObjects.WhereSubject != nil {
				matchFilters.SubjectType = []string{query.SelectObjects.WhereSubject.Type}

				if query.SelectObjects.WhereSubject.Id != warrant.Wildcard {
					matchFilters.SubjectId = []string{query.SelectObjects.WhereSubject.Id}
				}
			} else if query.SelectSubjects != nil && query.SelectSubjects.ForObject != nil {
				matchFilters.ObjectType = []string{query.SelectSubjects.ForObject.Type}

				if query.SelectSubjects.ForObject.Id != warrant.Wildcard {
					matchFilters.ObjectId = []string{query.SelectSubjects.ForObject.Id}
				}

				if query.SelectSubjects.SubjectTypes[0] != warrant.Wildcard {
					matchFilters.SubjectType = query.SelectSubjects.SubjectTypes
				}
			}

			res, err := svc.matchRelation(ctx, selectSubjects, objectType, relation, matchFilters, query.Expand)
			if err != nil {
				return nil, err
			}

			resultSet = resultSet.Union(res)
		}
	}

	return resultSet, nil
}

func (svc QueryService) matchRelation(ctx context.Context, selectSubjects bool, objectType string, relation string, matchFilters warrant.FilterParams, expand bool) (*ResultSet, error) {
	objectTypeDef, err := svc.objectTypeSvc.GetByTypeId(ctx, objectType)
	if err != nil {
		return nil, err
	}

	resultSet := NewResultSet()
	if _, exists := objectTypeDef.Relations[relation]; !exists {
		return resultSet, nil
	}

	// match any warrants at this level
	matchedWarrants, _, _, err := svc.matchWarrants(ctx, warrant.FilterParams{
		ObjectType: []string{objectType},
		ObjectId:   matchFilters.ObjectId,
		Relation:   []string{relation},
	})
	if err != nil {
		return nil, err
	}

	for _, matchedWarrant := range matchedWarrants {
		// match any encountered group warrants
		//nolint:gocritic
		if matchedWarrant.Subject.Relation != "" {
			// only expand group warrants if requested
			if !expand {
				continue
			}

			res, err := svc.matchRelation(ctx, selectSubjects, matchedWarrant.Subject.ObjectType, matchedWarrant.Subject.Relation, warrant.FilterParams{
				ObjectId:    []string{matchedWarrant.Subject.ObjectId},
				SubjectType: matchFilters.SubjectType,
				SubjectId:   matchFilters.SubjectId,
			}, expand)
			if err != nil {
				return nil, err
			}

			if selectSubjects {
				resultSet = resultSet.Union(res)
			} else if res.Len() > 0 {
				//nolint:gocritic
				if matchedWarrant.ObjectId != warrant.Wildcard {
					resultSet.Add(matchedWarrant.ObjectType, matchedWarrant.ObjectId, matchedWarrant)
				} else if len(matchFilters.ObjectId) > 0 {
					resultSet.Add(matchedWarrant.ObjectType, matchFilters.ObjectId[0], matchedWarrant)
				} else {
					//nolint:gocritic
					if matchedWarrant.ObjectId != warrant.Wildcard {
						resultSet.Add(matchedWarrant.ObjectType, matchedWarrant.ObjectId, matchedWarrant)
					} else if len(matchFilters.ObjectId) > 0 {
						resultSet.Add(matchedWarrant.ObjectType, matchFilters.ObjectId[0], matchedWarrant)
					} else {
						wcWarrantMatches, _, _, err := svc.matchWarrants(ctx, warrant.FilterParams{
							ObjectType: []string{matchedWarrant.ObjectType},
						})
						if err != nil {
							return nil, err
						}

						for _, wcWarrantMatch := range wcWarrantMatches {
							if wcWarrantMatch.ObjectId != warrant.Wildcard {
								resultSet.Add(wcWarrantMatch.ObjectType, wcWarrantMatch.ObjectId, matchedWarrant)
							}
						}
					}
				}
			}
		} else if selectSubjects {
			resultSet.Add(matchedWarrant.Subject.ObjectType, matchedWarrant.Subject.ObjectId, matchedWarrant)
		} else if matches(matchFilters.SubjectType, matchedWarrant.Subject.ObjectType) && matches(matchFilters.SubjectId, matchedWarrant.Subject.ObjectId) {
			resultSet.Add(matchedWarrant.ObjectType, matchedWarrant.ObjectId, matchedWarrant)
		}
	}

	// explore following levels if requested
	if expand {
		rule := objectTypeDef.Relations[relation]
		res, err := svc.matchRule(ctx, selectSubjects, objectType, relation, &rule, matchFilters, expand)
		if err != nil {
			return nil, err
		}
		resultSet = resultSet.Union(res)
	}

	return resultSet, nil
}

func (svc QueryService) matchRule(ctx context.Context, selectSubjects bool, objectType string, relation string, rule *objecttype.RelationRule, matchFilters warrant.FilterParams, expand bool) (*ResultSet, error) {
	switch rule.InheritIf {
	case "":
		// Do nothing, explicit matches already explored in matchRelation
		return NewResultSet(), nil
	case objecttype.InheritIfAllOf, objecttype.InheritIfAnyOf, objecttype.InheritIfNoneOf:
		return svc.matchSetRule(ctx, selectSubjects, objectType, relation, rule.InheritIf, rule.Rules, matchFilters, expand)
	default:
		// inherit relation if subject has:
		// (1) InheritIf on this object
		if rule.OfType == "" && rule.WithRelation == "" {
			return svc.matchRelation(ctx, selectSubjects, objectType, rule.InheritIf, matchFilters, expand)
		}

		// inherit relation if subject has:
		// (1) InheritIf on object (2) of type OfType
		// (3) with relation WithRelation on this object
		matchedWarrants, _, _, err := svc.matchWarrants(ctx, warrant.FilterParams{
			ObjectType:  []string{objectType},
			Relation:    []string{rule.WithRelation},
			ObjectId:    matchFilters.ObjectId,
			SubjectType: []string{rule.OfType},
		})
		if err != nil {
			return nil, err
		}

		resultSet := NewResultSet()
		for _, matchedWarrant := range matchedWarrants {
			res, err := svc.matchRelation(ctx, selectSubjects, rule.OfType, rule.InheritIf, warrant.FilterParams{
				ObjectType:  matchFilters.ObjectType,
				ObjectId:    []string{matchedWarrant.Subject.ObjectId},
				SubjectType: matchFilters.SubjectType,
				SubjectId:   matchFilters.SubjectId,
			}, expand)
			if err != nil {
				return nil, err
			}

			if selectSubjects {
				resultSet = resultSet.Union(res)
			} else if res.Len() > 0 {
				//nolint:gocritic
				if matchedWarrant.ObjectId != warrant.Wildcard {
					resultSet.Add(matchedWarrant.ObjectType, matchedWarrant.ObjectId, matchedWarrant)
				} else if len(matchFilters.ObjectId) > 0 {
					resultSet.Add(matchedWarrant.ObjectType, matchFilters.ObjectId[0], matchedWarrant)
				} else {
					wcWarrantMatches, _, _, err := svc.matchWarrants(ctx, warrant.FilterParams{
						ObjectType: []string{matchedWarrant.ObjectType},
					})
					if err != nil {
						return nil, err
					}

					for _, wcWarrantMatch := range wcWarrantMatches {
						if wcWarrantMatch.ObjectId != warrant.Wildcard {
							resultSet.Add(wcWarrantMatch.ObjectType, wcWarrantMatch.ObjectId, matchedWarrant)
						}
					}
				}
			}
		}

		return resultSet, nil
	}
}

func (svc QueryService) matchSetRule(
	ctx context.Context,
	selectSubjects bool,
	objectType string,
	relation string,
	setRuleType string,
	rules []objecttype.RelationRule,
	matchFilters warrant.FilterParams,
	expand bool,
) (*ResultSet, error) {
	switch setRuleType {
	case objecttype.InheritIfAllOf:
		var resultSet *ResultSet
		for i := range rules {
			res, err := svc.matchRule(ctx, selectSubjects, objectType, relation, &rules[i], matchFilters, expand)
			if err != nil {
				return nil, err
			}

			// short-circuit if no matches found for a rule
			if res.Len() == 0 {
				return NewResultSet(), nil
			}

			if resultSet == nil {
				resultSet = res
			} else {
				resultSet = resultSet.Intersect(res)
			}
		}

		return resultSet, nil
	case objecttype.InheritIfAnyOf:
		resultSet := NewResultSet()
		for i := range rules {
			res, err := svc.matchRule(ctx, selectSubjects, objectType, relation, &rules[i], matchFilters, expand)
			if err != nil {
				return nil, err
			}
			resultSet = resultSet.Union(res)
		}

		return resultSet, nil
	case objecttype.InheritIfNoneOf:
		return nil, service.NewInvalidRequestError("cannot query authorization models with object types that use the 'noneOf' operator.")
	default:
		return nil, ErrInvalidQuery
	}
}

func (svc QueryService) matchWarrants(ctx context.Context, matchFilters warrant.FilterParams) ([]warrant.WarrantSpec, *service.Cursor, *service.Cursor, error) {
	warrantListParams := service.DefaultListParams(warrant.WarrantListParamParser{})
	warrantListParams.Limit = MaxEdges
	return svc.warrantSvc.List(ctx, matchFilters, warrantListParams)
}

func matches(set []string, target string) bool {
	if len(set) == 0 || target == warrant.Wildcard {
		return true
	}

	for _, val := range set {
		if val == warrant.Wildcard || val == target {
			return true
		}
	}

	return false
}

func objectKey(objectType string, objectId string) string {
	return fmt.Sprintf("%s:%s", objectType, objectId)
}

func objectTypeAndObjectIdFromCursor(cursor *service.Cursor) (string, string, error) {
	objectType, objectId, found := strings.Cut(cursor.ID(), ":")
	if !found {
		return "", "", errors.New("invalid cursor")
	}

	return objectType, objectId, nil
}
