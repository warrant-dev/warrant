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

	"github.com/rs/zerolog/log"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	object "github.com/warrant-dev/warrant/pkg/object"
	"github.com/warrant-dev/warrant/pkg/service"
	baseSvc "github.com/warrant-dev/warrant/pkg/service"
)

var InvalidQueryErr = errors.New("invalid query")

type Relation struct {
	Name string
	objecttype.RelationRule
}

type QueryFilters struct {
	ObjectId    string
	SubjectType string
	SubjectId   string
}

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
	result := Result{Results: []QueryResult{}}
	resultMap := make(map[string]int)
	objects := make(map[string][]string, 0)
	selectedObjectTypes := make(map[string]bool)

	if (query.SelectObjects == nil && query.SelectSubjects == nil) || (query.SelectObjects != nil && query.SelectSubjects != nil) {
		return nil, InvalidQueryErr
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
		if query.SelectObjects != nil {
			isImplicit = !matches(query.SelectObjects.ObjectTypes, res.Warrant.ObjectType) || !matches(query.SelectObjects.Relations, res.Warrant.Relation)
		} else if query.SelectSubjects != nil {
			isImplicit = !matches(query.SelectSubjects.SubjectTypes, res.Warrant.Subject.ObjectType) || !matches(query.SelectSubjects.Relations, res.Warrant.Relation)
		} else {
			return nil, InvalidQueryErr
		}

		result.Results = append(result.Results, QueryResult{
			ObjectType: res.ObjectType,
			ObjectId:   res.ObjectId,
			Warrant:    res.Warrant,
			IsImplicit: isImplicit,
		})

		selectedObjectTypes[res.ObjectType] = true
		objects[res.ObjectType] = append(objects[res.ObjectType], res.ObjectId)
		resultMap[objectKey(res.ObjectType, res.ObjectId)] = len(result.Results) - 1
	}

	for selectedObjectType := range selectedObjectTypes {
		if len(objects[selectedObjectType]) > 0 {
			objectSpecs, err := svc.objectSvc.BatchGetByObjectTypeAndIds(ctx, selectedObjectType, objects[selectedObjectType])
			if err != nil {
				return nil, err
			}

			for _, objectSpec := range objectSpecs {
				result.Results[resultMap[objectKey(selectedObjectType, objectSpec.ObjectId)]].Meta = objectSpec.Meta
			}
		}
	}

	return &result, nil
}

func (svc QueryService) query(ctx context.Context, query *Query, objectTypeMap objecttype.ObjectTypeMap) (*ResultSet, error) {
	var objectTypes []string
	var selectObjects bool
	if query.SelectObjects != nil {
		selectObjects = true
		if query.SelectObjects.ObjectTypes[0] == Wildcard {
			for objectType := range objectTypeMap {
				objectTypes = append(objectTypes, objectType)
			}
		} else {
			objectTypes = append(objectTypes, query.SelectObjects.ObjectTypes...)
		}
	} else if query.SelectSubjects != nil {
		selectObjects = false
		if query.SelectSubjects.ForObject != nil {
			objectTypes = append(objectTypes, query.SelectSubjects.ForObject.Type)
		} else {
			for objectType := range objectTypeMap {
				objectTypes = append(objectTypes, objectType)
			}
		}
	} else {
		return nil, InvalidQueryErr
	}

	resultSet := NewResultSet()
	for _, objectType := range objectTypes {
		var relations []string
		objectTypeDef, err := objectTypeMap.GetByTypeId(objectType)
		if err != nil {
			return nil, err
		}

		if query.SelectObjects != nil {
			if query.SelectObjects.Relations[0] == Wildcard {
				for relation := range objectTypeDef.Relations {
					relations = append(relations, relation)
				}
			} else {
				relations = append(relations, query.SelectObjects.Relations...)
			}
		} else {
			if query.SelectSubjects.Relations[0] == Wildcard {
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

				if query.SelectObjects.WhereSubject.Id != Wildcard {
					matchFilters.SubjectId = []string{query.SelectObjects.WhereSubject.Id}
				}
			} else if query.SelectSubjects != nil && query.SelectSubjects.ForObject != nil {
				matchFilters.ObjectType = []string{query.SelectSubjects.ForObject.Type}

				if query.SelectSubjects.ForObject.Id != Wildcard {
					matchFilters.ObjectId = []string{query.SelectSubjects.ForObject.Id}
				}

				if query.SelectSubjects.SubjectTypes[0] != Wildcard {
					matchFilters.SubjectType = query.SelectSubjects.SubjectTypes
				}
			}

			res, err := svc.matchRelation(ctx, selectObjects, objectTypeMap, objectType, relation, matchFilters, query.Expand)
			if err != nil {
				return nil, err
			}

			resultSet = resultSet.Union(res)
		}
	}

	return resultSet, nil
}

func (svc QueryService) matchRelation(ctx context.Context, selectObjects bool, objectTypes objecttype.ObjectTypeMap, objectType string, relation string, matchFilters warrant.FilterParams, expand bool) (*ResultSet, error) {
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
		ObjectType:      []string{objectType},
		ObjectId:        matchFilters.ObjectId,
		Relation:        []string{relation},
		SubjectType:     matchFilters.SubjectType,
		SubjectId:       matchFilters.SubjectId,
		SubjectRelation: []string{""}, // ignore group warrants
	})
	if err != nil {
		return nil, err
	}

	for _, matchedWarrant := range matchedWarrants {
		if selectObjects {
			resultSet.Add(matchedWarrant.ObjectType, matchedWarrant.ObjectId, matchedWarrant)
		} else {
			resultSet.Add(matchedWarrant.Subject.ObjectType, matchedWarrant.Subject.ObjectId, matchedWarrant)
		}
	}

	// explore following levels if requested
	if expand {
		rule := objectTypeDef.Relations[relation]
		res, err := svc.matchRule(ctx, selectObjects, objectTypes, objectType, relation, &rule, matchFilters, expand)
		if err != nil {
			return nil, err
		}
		resultSet = resultSet.Union(res)
	}

	return resultSet, nil
}

func (svc QueryService) matchRule(ctx context.Context, selectObjects bool, objectTypes objecttype.ObjectTypeMap, objectType string, relation string, rule *objecttype.RelationRule, matchFilters warrant.FilterParams, expand bool) (*ResultSet, error) {
	log.Ctx(ctx).Debug().
		Str("objectType", objectType).
		Str("relation", relation).
		Str("rule.InheritIf", rule.InheritIf).
		Str("rule.OfType", rule.OfType).
		Str("rule.WithRelation", rule.WithRelation).
		Str("filters", matchFilters.String()).
		Msg("matchRule    ")
	switch rule.InheritIf {
	case "":
		// Do nothing, explicit matches already explored in matchRelation
		return NewResultSet(), nil
	case objecttype.InheritIfAllOf, objecttype.InheritIfAnyOf, objecttype.InheritIfNoneOf:
		return svc.matchSetRule(ctx, selectObjects, objectTypes, objectType, relation, rule.InheritIf, rule.Rules, matchFilters, expand)
	default:
		// inherit relation if subject has:
		// (1) InheritIf on this object
		if rule.OfType == "" && rule.WithRelation == "" {
			return svc.matchRelation(ctx, selectObjects, objectTypes, objectType, rule.InheritIf, matchFilters, expand)
		}

		// inherit relation if subject has:
		// (1) InheritIf on object (2) of type OfType
		// (3) with relation WithRelation on this object
		matchedWarrants, err := svc.matchWarrants(ctx, warrant.FilterParams{
			ObjectType:      []string{objectType},
			Relation:        []string{rule.WithRelation},
			ObjectId:        matchFilters.ObjectId,
			SubjectType:     []string{rule.OfType},
			SubjectRelation: []string{""}, // ignore group warrants
		})
		if err != nil {
			return nil, err
		}

		resultSet := NewResultSet()
		for _, matchedWarrant := range matchedWarrants {
			res, err := svc.matchRelation(ctx, selectObjects, objectTypes, rule.OfType, rule.InheritIf, warrant.FilterParams{
				ObjectType:  matchFilters.ObjectType,
				ObjectId:    []string{matchedWarrant.Subject.ObjectId},
				SubjectType: matchFilters.SubjectType,
				SubjectId:   matchFilters.SubjectId,
			}, expand)
			if err != nil {
				return nil, err
			}

			if res.Len() > 0 {
				if selectObjects {
					resultSet.Add(matchedWarrant.ObjectType, matchedWarrant.ObjectId, matchedWarrant)
				} else {
					for iter := res.List(); iter != nil; iter = iter.Next() {
						resultSet.Add(iter.Warrant.Subject.ObjectType, iter.Warrant.Subject.ObjectId, iter.Warrant)
					}
				}
			}
		}

		return resultSet, nil
	}
}

func (svc QueryService) matchSetRule(
	ctx context.Context,
	selectObjects bool,
	objectTypes objecttype.ObjectTypeMap,
	objectType string,
	relation string,
	setRuleType string,
	rules []objecttype.RelationRule,
	matchFilters warrant.FilterParams,
	expand bool,
) (*ResultSet, error) {
	log.Ctx(ctx).Debug().
		Str("objectType", objectType).
		Str("relation", relation).
		Str("setRuleType", setRuleType).
		Str("filters", matchFilters.String()).
		Msg("matchSetRule ")
	switch setRuleType {
	case objecttype.InheritIfAllOf:
		var resultSet *ResultSet
		for _, rule := range rules {
			res, err := svc.matchRule(ctx, selectObjects, objectTypes, objectType, relation, &rule, matchFilters, expand)
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
		for _, rule := range rules {
			res, err := svc.matchRule(ctx, selectObjects, objectTypes, objectType, relation, &rule, matchFilters, expand)
			if err != nil {
				return nil, err
			}
			resultSet = resultSet.Union(res)
		}

		return resultSet, nil
	case objecttype.InheritIfNoneOf:
		return nil, service.NewInvalidRequestError("cannot query object-types or relations that use 'noneOf'")
	default:
		return nil, InvalidQueryErr
	}
}

func (svc QueryService) matchWarrants(ctx context.Context, matchFilters warrant.FilterParams) ([]warrant.WarrantSpec, error) {
	warrantListParams := service.DefaultListParams(warrant.WarrantListParamParser{})
	warrantListParams.Limit = 1000 // explore up to 1000 edges
	return svc.warrantSvc.List(ctx, &matchFilters, warrantListParams)
}

func matches(set []string, target string) bool {
	if target == Wildcard {
		return true
	}

	for _, val := range set {
		if val == Wildcard || val == target {
			return true
		}
	}

	return false
}

func objectKey(objectType string, objectId string) string {
	return fmt.Sprintf("%s:%s", objectType, objectId)
}
