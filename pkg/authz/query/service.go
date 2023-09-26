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

	"github.com/rs/zerolog/log"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	object "github.com/warrant-dev/warrant/pkg/object"
	"github.com/warrant-dev/warrant/pkg/service"
	baseSvc "github.com/warrant-dev/warrant/pkg/service"
)

var ErrInvalidQuery = errors.New("invalid query")

type QueryService struct {
	baseSvc.BaseService
	objectTypeSvc *objecttype.ObjectTypeService
	warrantSvc    *warrant.WarrantService
	objectSvc     *object.ObjectService
}

func NewService(env service.Env, objectTypeSvc *objecttype.ObjectTypeService, warrantSvc *warrant.WarrantService, objectSvc *object.ObjectService) QueryService {
	return QueryService{
		BaseService:   baseSvc.NewBaseService(env),
		objectTypeSvc: objectTypeSvc,
		warrantSvc:    warrantSvc,
		objectSvc:     objectSvc,
	}
}

func (svc QueryService) Query(ctx context.Context, query *Query, listParams service.ListParams) (*Result, error) {
	queryResults := []QueryResult{}
	resultMap := make(map[string]int)
	objects := make(map[string][]string, 0)
	selectedObjectTypes := make(map[string]bool)

	if (query.SelectObjects == nil && query.SelectSubjects == nil) || (query.SelectObjects != nil && query.SelectSubjects != nil) {
		return nil, ErrInvalidQuery
	}

	objectTypeMap, err := svc.objectTypeSvc.GetTypeMap(ctx)
	if err != nil {
		return nil, err
	}

	resultSet, err := svc.query(ctx, query, objectTypeMap)
	if err != nil {
		return nil, err
	}

	for res := resultSet.List(); res != nil; res = res.Next() {
		var isImplicit bool
		//nolint:gocritic
		if query.SelectObjects != nil {
			isImplicit = !matches(query.SelectObjects.ObjectTypes, res.Warrant.ObjectType) || !matches(query.SelectObjects.Relations, res.Warrant.Relation)
		} else if query.SelectSubjects != nil {
			isImplicit = !matches(query.SelectSubjects.SubjectTypes, res.Warrant.Subject.ObjectType) || !matches(query.SelectSubjects.Relations, res.Warrant.Relation)
		} else {
			return nil, ErrInvalidQuery
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
	case "objectType":
		switch listParams.SortOrder {
		case service.SortOrderAsc:
			sort.Sort(ByObjectTypeAsc(queryResults))
		case service.SortOrderDesc:
			sort.Sort(ByObjectTypeDesc(queryResults))
		}
	case "objectId":
		switch listParams.SortOrder {
		case service.SortOrderAsc:
			sort.Sort(ByObjectIdAsc(queryResults))
		case service.SortOrderDesc:
			sort.Sort(ByObjectIdDesc(queryResults))
		}
	default:
		return nil, ErrInvalidQuery
	}

	index := 0
	// skip ahead if lastId passed in
	if listParams.AfterId != nil {
		lastIdSpec, err := StringToLastIdSpec(*listParams.AfterId)
		if err != nil {
			return nil, err
		}

		for index < len(queryResults) && (queryResults[index].ObjectType != lastIdSpec.ObjectType || queryResults[index].ObjectId != lastIdSpec.ObjectId) {
			index++
		}
	}

	paginatedQueryResults := []QueryResult{}
	for len(paginatedQueryResults) < listParams.Limit && index < len(queryResults) {
		paginatedQueryResult := queryResults[index]
		paginatedQueryResults = append(paginatedQueryResults, paginatedQueryResult)
		selectedObjectTypes[paginatedQueryResult.ObjectType] = true
		objects[paginatedQueryResult.ObjectType] = append(objects[paginatedQueryResult.ObjectType], paginatedQueryResult.ObjectId)
		resultMap[objectKey(paginatedQueryResult.ObjectType, paginatedQueryResult.ObjectId)] = len(paginatedQueryResults) - 1
		index++
	}

	lastId := ""
	if index < len(queryResults) {
		lastId, err = LastIdSpecToString(LastIdSpec{
			ObjectType: queryResults[index].ObjectType,
			ObjectId:   queryResults[index].ObjectId,
		})
		if err != nil {
			return nil, err
		}
	}

	for selectedObjectType := range selectedObjectTypes {
		if len(objects[selectedObjectType]) > 0 {
			objectSpecs, err := svc.objectSvc.BatchGetByObjectTypeAndIds(ctx, selectedObjectType, objects[selectedObjectType])
			if err != nil {
				return nil, err
			}

			for _, objectSpec := range objectSpecs {
				paginatedQueryResults[resultMap[objectKey(selectedObjectType, objectSpec.ObjectId)]].Meta = objectSpec.Meta
			}
		}
	}

	return &Result{
		Results: paginatedQueryResults,
		LastId:  lastId,
	}, nil
}

func (svc QueryService) query(ctx context.Context, query *Query, objectTypeMap objecttype.ObjectTypeMap) (*ResultSet, error) {
	var objectTypes []string
	var selectSubjects bool
	//nolint:gocritic
	if query.SelectObjects != nil {
		selectSubjects = false
		if query.SelectObjects.ObjectTypes[0] == warrant.Wildcard {
			for objectType := range objectTypeMap {
				objectTypes = append(objectTypes, objectType)
			}
		} else {
			objectTypes = append(objectTypes, query.SelectObjects.ObjectTypes...)
		}
	} else if query.SelectSubjects != nil {
		selectSubjects = true
		if query.SelectSubjects.ForObject != nil {
			objectTypes = append(objectTypes, query.SelectSubjects.ForObject.Type)
		} else {
			for objectType := range objectTypeMap {
				objectTypes = append(objectTypes, objectType)
			}
		}
	} else {
		return nil, ErrInvalidQuery
	}

	resultSet := NewResultSet()
	for _, objectType := range objectTypes {
		var relations []string
		objectTypeDef, err := objectTypeMap.GetByTypeId(objectType)
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

			res, err := svc.matchRelation(ctx, selectSubjects, objectTypeMap, objectType, relation, matchFilters, query.Expand)
			if err != nil {
				return nil, err
			}

			resultSet = resultSet.Union(res)
		}
	}

	return resultSet, nil
}

func (svc QueryService) matchRelation(ctx context.Context, selectSubjects bool, objectTypes objecttype.ObjectTypeMap, objectType string, relation string, matchFilters warrant.FilterParams, expand bool) (*ResultSet, error) {
	log.Ctx(ctx).Debug().
		Str("objectType", objectType).
		Str("relation", relation).
		Str("filters", matchFilters.String()).
		Msg("matchRelation")
	objectTypeDef, err := objectTypes.GetByTypeId(objectType)
	if err != nil {
		return nil, err
	}

	if _, exists := objectTypeDef.Relations[relation]; !exists {
		return NewResultSet(), nil
	}

	resultSet := NewResultSet()
	// match any warrants at this level
	matchedWarrants, err := svc.matchWarrants(ctx, warrant.FilterParams{
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
			res, err := svc.matchRelation(ctx, selectSubjects, objectTypes, matchedWarrant.Subject.ObjectType, matchedWarrant.Subject.Relation, warrant.FilterParams{
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
						wcWarrantMatches, err := svc.matchWarrants(ctx, warrant.FilterParams{
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
		res, err := svc.matchRule(ctx, selectSubjects, objectTypes, objectType, relation, &rule, matchFilters, expand)
		if err != nil {
			return nil, err
		}
		resultSet = resultSet.Union(res)
	}

	return resultSet, nil
}

func (svc QueryService) matchRule(ctx context.Context, selectSubjects bool, objectTypes objecttype.ObjectTypeMap, objectType string, relation string, rule *objecttype.RelationRule, matchFilters warrant.FilterParams, expand bool) (*ResultSet, error) {
	switch rule.InheritIf {
	case "":
		// Do nothing, explicit matches already explored in matchRelation
		return NewResultSet(), nil
	case objecttype.InheritIfAllOf, objecttype.InheritIfAnyOf, objecttype.InheritIfNoneOf:
		return svc.matchSetRule(ctx, selectSubjects, objectTypes, objectType, relation, rule.InheritIf, rule.Rules, matchFilters, expand)
	default:
		// inherit relation if subject has:
		// (1) InheritIf on this object
		if rule.OfType == "" && rule.WithRelation == "" {
			return svc.matchRelation(ctx, selectSubjects, objectTypes, objectType, rule.InheritIf, matchFilters, expand)
		}

		// inherit relation if subject has:
		// (1) InheritIf on object (2) of type OfType
		// (3) with relation WithRelation on this object
		matchedWarrants, err := svc.matchWarrants(ctx, warrant.FilterParams{
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
			res, err := svc.matchRelation(ctx, selectSubjects, objectTypes, rule.OfType, rule.InheritIf, warrant.FilterParams{
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
					wcWarrantMatches, err := svc.matchWarrants(ctx, warrant.FilterParams{
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
	objectTypes objecttype.ObjectTypeMap,
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
			res, err := svc.matchRule(ctx, selectSubjects, objectTypes, objectType, relation, &rules[i], matchFilters, expand)
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
			res, err := svc.matchRule(ctx, selectSubjects, objectTypes, objectType, relation, &rules[i], matchFilters, expand)
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

func (svc QueryService) matchWarrants(ctx context.Context, matchFilters warrant.FilterParams) ([]warrant.WarrantSpec, error) {
	warrantListParams := service.DefaultListParams(warrant.WarrantListParamParser{})
	warrantListParams.Limit = 1000 // explore up to 1000 edges
	return svc.warrantSvc.List(ctx, &matchFilters, warrantListParams)
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
