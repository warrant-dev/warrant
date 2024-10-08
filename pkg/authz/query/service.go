// Copyright 2024 WorkOS, Inc.
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
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/pkg/object"
	"github.com/warrant-dev/warrant/pkg/service"
)

const (
	MaxObjectTypes = 1000
	MaxEdges       = 10000
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

func (svc QueryService) Query(ctx context.Context, query Query, listParams service.ListParams) ([]QueryResult, *service.Cursor, *service.Cursor, error) {
	queryResults := make([]QueryResult, 0)
	resultMap := make(map[string][]int)
	objects := make(map[string][]string)
	selectedObjectTypes := make(map[string]bool)

	if (query.SelectObjects == nil && query.SelectSubjects == nil) || (query.SelectObjects != nil && query.SelectSubjects != nil) {
		return nil, nil, nil, ErrInvalidQuery
	}

	resultSet := NewResultSet()
	switch {
	case query.SelectObjects != nil:
		var (
			objectTypes []objecttype.ObjectTypeSpec
			err         error
		)
		if query.SelectObjects.ObjectTypes[0] == warrant.Wildcard {
			objectTypesListParams := service.DefaultListParams(objecttype.ObjectTypeListParamParser{})
			objectTypesListParams.WithLimit(MaxObjectTypes)
			objectTypes, _, _, err = svc.objectTypeSvc.List(ctx, objectTypesListParams)
			if err != nil {
				return nil, nil, nil, err
			}
		} else {
			for _, typeId := range query.SelectObjects.ObjectTypes {
				objectType, err := svc.objectTypeSvc.GetByTypeId(ctx, typeId)
				if err != nil {
					return nil, nil, nil, err
				}

				objectTypes = append(objectTypes, *objectType)
			}
		}

		for _, objectType := range objectTypes {
			var relations []string
			if query.SelectObjects.Relations[0] == warrant.Wildcard {
				for relation := range objectType.Relations {
					relations = append(relations, relation)
				}
			} else {
				for _, relation := range query.SelectObjects.Relations {
					if _, ok := objectType.Relations[relation]; ok {
						relations = append(relations, relation)
					}
				}
			}

			for _, relation := range relations {
				queryResult, err := svc.query(ctx, Query{
					Expand: query.Expand,
					SelectObjects: &SelectObjects{
						ObjectTypes:  []string{objectType.Type},
						Relations:    []string{relation},
						WhereSubject: query.SelectObjects.WhereSubject,
					},
					Context: query.Context,
				}, 0)
				if err != nil {
					return nil, nil, nil, err
				}

				for res := queryResult.List(); res != nil; res = res.Next() {
					resultSet.Add(res.ObjectType, res.ObjectId, relation, res.Warrant, res.Policy, res.IsImplicit)
				}
			}
		}
	case query.SelectSubjects != nil:
		var (
			subjectTypes []objecttype.ObjectTypeSpec
			err          error
		)
		if query.SelectSubjects.SubjectTypes[0] == warrant.Wildcard {
			objectTypesListParams := service.DefaultListParams(objecttype.ObjectTypeListParamParser{})
			objectTypesListParams.WithLimit(MaxObjectTypes)
			subjectTypes, _, _, err = svc.objectTypeSvc.List(ctx, objectTypesListParams)
			if err != nil {
				return nil, nil, nil, err
			}
		} else {
			for _, typeId := range query.SelectSubjects.SubjectTypes {
				subjectType, err := svc.objectTypeSvc.GetByTypeId(ctx, typeId)
				if err != nil {
					return nil, nil, nil, err
				}

				subjectTypes = append(subjectTypes, *subjectType)
			}
		}

		objectType, err := svc.objectTypeSvc.GetByTypeId(ctx, query.SelectSubjects.ForObject.Type)
		if err != nil {
			return nil, nil, nil, err
		}

		var relations []string
		if query.SelectSubjects.Relations[0] == warrant.Wildcard {
			for relation := range objectType.Relations {
				relations = append(relations, relation)
			}
		} else {
			for _, relation := range query.SelectSubjects.Relations {
				if _, ok := objectType.Relations[relation]; ok {
					relations = append(relations, relation)
				}
			}
		}

		for _, subjectType := range subjectTypes {
			for _, relation := range relations {
				queryResult, err := svc.query(ctx, Query{
					Expand: query.Expand,
					SelectSubjects: &SelectSubjects{
						SubjectTypes: []string{subjectType.Type},
						Relations:    []string{relation},
						ForObject:    query.SelectSubjects.ForObject,
					},
					Context: query.Context,
				}, 0)
				if err != nil {
					return nil, nil, nil, err
				}

				for res := queryResult.List(); res != nil; res = res.Next() {
					resultSet.Add(res.ObjectType, res.ObjectId, relation, res.Warrant, res.Policy, res.IsImplicit)
				}
			}
		}
	default:
		return nil, nil, nil, ErrInvalidQuery
	}

	for res := resultSet.List(); res != nil; res = res.Next() {
		var err error
		addResult := true
		if res.Policy != "" {
			addResult, err = res.Policy.Eval(query.Context)
			if err != nil {
				return nil, nil, nil, err
			}
		}

		if addResult {
			queryResults = append(queryResults, QueryResult{
				ObjectType: res.ObjectType,
				ObjectId:   res.ObjectId,
				Relation:   res.Relation,
				Warrant:    res.Warrant,
				IsImplicit: res.IsImplicit,
			})
		}
	}

	// handle sorting and pagination
	switch listParams.SortBy {
	case PrimarySortKey:
		switch listParams.SortOrder {
		case service.SortOrderAsc:
			sort.Sort(ByObjectTypeAndObjectIdAndRelationAsc(queryResults))
		case service.SortOrderDesc:
			sort.Sort(ByObjectTypeAndObjectIdAndRelationDesc(queryResults))
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
		lastObjectType, lastObjectId, lastRelation, err := objectTypeAndObjectIdAndRelationFromCursor(listParams.NextCursor)
		if err != nil {
			return nil, nil, nil, service.NewInvalidParameterError("nextCursor", "invalid cursor")
		}

		start = 0
		for start < len(queryResults) && (queryResults[start].ObjectType != lastObjectType || queryResults[start].ObjectId != lastObjectId || queryResults[start].Relation != lastRelation) {
			start++
		}

		end = start + listParams.Limit
	} else if listParams.PrevCursor != nil { // seek backward if PrevCursor passed in
		lastObjectType, lastObjectId, lastRelation, err := objectTypeAndObjectIdAndRelationFromCursor(listParams.PrevCursor)
		if err != nil {
			return nil, nil, nil, service.NewInvalidParameterError("prevCursor", "invalid cursor")
		}

		end = len(queryResults) - 1
		for end > 0 && (queryResults[end].ObjectType != lastObjectType || queryResults[end].ObjectId != lastObjectId || queryResults[end].Relation != lastRelation) {
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

		prevCursor = service.NewCursor(objectRelationKey(queryResults[start].ObjectType, queryResults[start].ObjectId, queryResults[start].Relation), value)
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

		nextCursor = service.NewCursor(objectRelationKey(queryResults[end].ObjectType, queryResults[end].ObjectId, queryResults[end].Relation), value)
	}

	for start < end && start < len(queryResults) {
		paginatedQueryResult := queryResults[start]
		paginatedQueryResults = append(paginatedQueryResults, paginatedQueryResult)
		selectedObjectTypes[paginatedQueryResult.ObjectType] = true
		objects[paginatedQueryResult.ObjectType] = append(objects[paginatedQueryResult.ObjectType], paginatedQueryResult.ObjectId)

		objKey := objectKey(paginatedQueryResult.ObjectType, paginatedQueryResult.ObjectId)
		resultMap[objKey] = append(resultMap[objKey], len(paginatedQueryResults)-1)
		start++
	}

	for selectedObjectType := range selectedObjectTypes {
		if len(objects[selectedObjectType]) > 0 {
			objectSpecs, err := svc.objectSvc.BatchGetByObjectTypeAndIds(ctx, selectedObjectType, objects[selectedObjectType])
			if err != nil {
				return nil, nil, nil, err
			}

			for _, objectSpec := range objectSpecs {
				for _, resultIdx := range resultMap[objectKey(selectedObjectType, objectSpec.ObjectId)] {
					paginatedQueryResults[resultIdx].Meta = objectSpec.Meta
				}
			}
		}
	}

	return paginatedQueryResults, prevCursor, nextCursor, nil
}

func (svc QueryService) query(ctx context.Context, query Query, level int) (*ResultSet, error) {
	switch {
	case query.SelectObjects != nil:
		objectType := query.SelectObjects.ObjectTypes[0]
		relation := query.SelectObjects.Relations[0]
		objectTypeDef, err := svc.objectTypeSvc.GetByTypeId(ctx, objectType)
		if err != nil {
			return nil, err
		}

		if _, found := objectTypeDef.Relations[relation]; !found {
			return nil, errors.New(fmt.Sprintf("query: relation %s does not exist on object type %s", relation, objectType))
		}

		// base case: explicit query
		matchedWarrants, err := svc.listWarrants(ctx, warrant.FilterParams{
			ObjectType: query.SelectObjects.ObjectTypes[0],
			Relation:   query.SelectObjects.Relations[0],
		})
		if err != nil {
			return nil, err
		}

		resultSet := NewResultSet()
		for _, matchedWarrant := range matchedWarrants {
			if matchedWarrant.Subject.Relation != "" {
				// handle group warrants
				subset, err := svc.query(ctx, Query{
					Expand: query.Expand,
					SelectSubjects: &SelectSubjects{
						Relations:    []string{matchedWarrant.Subject.Relation},
						SubjectTypes: []string{query.SelectObjects.WhereSubject.Type},
						ForObject: &Resource{
							Type: matchedWarrant.Subject.ObjectType,
							Id:   matchedWarrant.Subject.ObjectId,
						},
					},
					Context: query.Context,
				}, 0)
				if err != nil {
					return nil, err
				}

				for sub := subset.List(); sub != nil; sub = sub.Next() {
					if sub.ObjectType != query.SelectObjects.WhereSubject.Type || sub.ObjectId != query.SelectObjects.WhereSubject.Id {
						continue
					}

					policy := matchedWarrant.Policy.And(sub.Policy)
					if matchedWarrant.ObjectId == warrant.Wildcard {
						expandedObjects, err := svc.listObjectsByType(ctx, matchedWarrant.ObjectType)
						if err != nil {
							return nil, err
						}

						for _, obj := range expandedObjects {
							resultSet.Add(obj.ObjectType, obj.ObjectId, relation, matchedWarrant, policy, sub.IsImplicit || level > 0)
						}
					} else {
						resultSet.Add(matchedWarrant.ObjectType, matchedWarrant.ObjectId, relation, matchedWarrant, policy, sub.IsImplicit || level > 0)
					}
				}
			} else if query.SelectObjects.WhereSubject == nil ||
				(matchedWarrant.Subject.ObjectType == query.SelectObjects.WhereSubject.Type &&
					matchedWarrant.Subject.ObjectId == query.SelectObjects.WhereSubject.Id) {
				if matchedWarrant.ObjectId == warrant.Wildcard {
					expandedObjects, err := svc.listObjectsByType(ctx, matchedWarrant.ObjectType)
					if err != nil {
						return nil, err
					}

					for _, obj := range expandedObjects {
						resultSet.Add(obj.ObjectType, obj.ObjectId, relation, matchedWarrant, matchedWarrant.Policy, false)
					}
				} else {
					resultSet.Add(matchedWarrant.ObjectType, matchedWarrant.ObjectId, relation, matchedWarrant, matchedWarrant.Policy, false)
				}
			}
		}

		if query.Expand {
			implicitResultSet, err := svc.queryRule(ctx, query, level+1, relation, objectTypeDef.Relations[relation])
			if err != nil {
				return nil, err
			}

			for res := implicitResultSet.List(); res != nil; res = res.Next() {
				resultSet.Add(res.ObjectType, res.ObjectId, relation, res.Warrant, res.Policy, res.IsImplicit || level > 0)
			}
		}

		return resultSet, nil
	case query.SelectSubjects != nil:
		objectType := query.SelectSubjects.ForObject.Type
		relation := query.SelectSubjects.Relations[0]
		objectTypeDef, err := svc.objectTypeSvc.GetByTypeId(ctx, objectType)
		if err != nil {
			return nil, err
		}

		if _, found := objectTypeDef.Relations[relation]; !found {
			return nil, errors.New(fmt.Sprintf("query: relation %s does not exist on object type %s", relation, objectType))
		}

		// base case: explicit query
		matchedWarrants, err := svc.listWarrants(ctx, warrant.FilterParams{
			ObjectType: query.SelectSubjects.ForObject.Type,
			ObjectId:   query.SelectSubjects.ForObject.Id,
			Relation:   query.SelectSubjects.Relations[0],
		})
		if err != nil {
			return nil, err
		}

		resultSet := NewResultSet()
		for _, matchedWarrant := range matchedWarrants {
			if matchedWarrant.Subject.Relation != "" {
				// handle group warrants
				subset, err := svc.query(ctx, Query{
					Expand: query.Expand,
					SelectSubjects: &SelectSubjects{
						Relations:    []string{matchedWarrant.Subject.Relation},
						SubjectTypes: query.SelectSubjects.SubjectTypes,
						ForObject: &Resource{
							Type: matchedWarrant.Subject.ObjectType,
							Id:   matchedWarrant.Subject.ObjectId,
						},
					},
					Context: query.Context,
				}, 0)
				if err != nil {
					return nil, err
				}

				for sub := subset.List(); sub != nil; sub = sub.Next() {
					resultSet.Add(sub.ObjectType, sub.ObjectId, relation, matchedWarrant, matchedWarrant.Policy.And(sub.Policy), sub.IsImplicit || level > 0)
				}
			} else if query.SelectSubjects.SubjectTypes[0] == matchedWarrant.Subject.ObjectType {
				resultSet.Add(matchedWarrant.Subject.ObjectType, matchedWarrant.Subject.ObjectId, relation, matchedWarrant, matchedWarrant.Policy, false)
			}
		}

		if query.Expand {
			implicitResultSet, err := svc.queryRule(ctx, query, level+1, relation, objectTypeDef.Relations[relation])
			if err != nil {
				return nil, err
			}

			for res := implicitResultSet.List(); res != nil; res = res.Next() {
				resultSet.Add(res.ObjectType, res.ObjectId, relation, res.Warrant, res.Policy, res.IsImplicit || level > 0)
			}
		}

		return resultSet, nil
	default:
		return nil, ErrInvalidQuery
	}
}

func (svc QueryService) queryRule(ctx context.Context, query Query, level int, relation string, rule objecttype.RelationRule) (*ResultSet, error) {
	switch rule.InheritIf {
	case "":
		return NewResultSet(), nil
	case objecttype.InheritIfAllOf:
		var resultSet *ResultSet
		for _, r := range rule.Rules {
			res, err := svc.queryRule(ctx, query, level, relation, r)
			if err != nil {
				return nil, err
			}

			if resultSet == nil {
				resultSet = res
			} else {
				resultSet = resultSet.Intersect(res)
			}
		}

		return resultSet, nil
	case objecttype.InheritIfAnyOf:
		var resultSet *ResultSet
		for _, r := range rule.Rules {
			res, err := svc.queryRule(ctx, query, level, relation, r)
			if err != nil {
				return nil, err
			}

			if resultSet == nil {
				resultSet = res
			} else {
				resultSet = resultSet.Union(res)
			}
		}

		return resultSet, nil
	case objecttype.InheritIfNoneOf:
		return nil, service.NewInvalidRequestError("cannot query authorization models with object types that use the 'noneOf' operator.")
	default:
		switch {
		case query.SelectObjects != nil:
			if rule.OfType == "" && rule.WithRelation == "" {
				results, err := svc.query(ctx, Query{
					Expand: true,
					SelectObjects: &SelectObjects{
						ObjectTypes:  query.SelectObjects.ObjectTypes,
						WhereSubject: query.SelectObjects.WhereSubject,
						Relations:    []string{rule.InheritIf},
					},
					Context: query.Context,
				}, 0)
				if err != nil {
					return nil, err
				}

				resultSet := NewResultSet()
				for res := results.List(); res != nil; res = res.Next() {
					resultSet.Add(res.ObjectType, res.ObjectId, relation, res.Warrant, res.Policy, res.IsImplicit || level > 0)
				}

				return resultSet, nil
			} else {
				indirectWarrants, err := svc.listWarrants(ctx, warrant.FilterParams{
					ObjectType:  rule.OfType,
					Relation:    rule.InheritIf,
					SubjectType: query.SelectObjects.WhereSubject.Type,
					SubjectId:   query.SelectObjects.WhereSubject.Id,
				})
				if err != nil {
					return nil, err
				}

				resultSet := NewResultSet()
				for _, indirectWarrant := range indirectWarrants {
					if indirectWarrant.Subject.Relation != "" {
						continue
					}

					if indirectWarrant.ObjectId == warrant.Wildcard {
						expandedObjects, err := svc.listObjectsByType(ctx, indirectWarrant.ObjectType)
						if err != nil {
							return nil, err
						}

						for _, obj := range expandedObjects {
							if obj.ObjectId != indirectWarrant.Subject.ObjectId {
								inheritedResults, err := svc.query(ctx, Query{
									Expand: query.Expand,
									SelectObjects: &SelectObjects{
										ObjectTypes: query.SelectObjects.ObjectTypes,
										WhereSubject: &Resource{
											Type: obj.ObjectType,
											Id:   obj.ObjectId,
										},
										Relations: []string{rule.WithRelation},
									},
									Context: query.Context,
								}, 0)
								if err != nil {
									return nil, err
								}

								for res := inheritedResults.List(); res != nil; res = res.Next() {
									resultSet.Add(res.ObjectType, res.ObjectId, relation, res.Warrant, indirectWarrant.Policy.Or(res.Policy), res.IsImplicit || level > 0)
								}
							}
						}
					} else {
						inheritedResults, err := svc.query(ctx, Query{
							Expand: query.Expand,
							SelectObjects: &SelectObjects{
								ObjectTypes: query.SelectObjects.ObjectTypes,
								WhereSubject: &Resource{
									Type: indirectWarrant.ObjectType,
									Id:   indirectWarrant.ObjectId,
								},
								Relations: []string{rule.WithRelation},
							},
							Context: query.Context,
						}, 0)
						if err != nil {
							return nil, err
						}

						for res := inheritedResults.List(); res != nil; res = res.Next() {
							resultSet.Add(res.ObjectType, res.ObjectId, relation, res.Warrant, indirectWarrant.Policy.Or(res.Policy), res.IsImplicit || level > 0)
						}
					}
				}

				return resultSet, nil
			}
		case query.SelectSubjects != nil:
			if rule.OfType == "" && rule.WithRelation == "" {
				results, err := svc.query(ctx, Query{
					Expand: true,
					SelectSubjects: &SelectSubjects{
						SubjectTypes: query.SelectSubjects.SubjectTypes,
						Relations:    []string{rule.InheritIf},
						ForObject:    query.SelectSubjects.ForObject,
					},
					Context: query.Context,
				}, 0)
				if err != nil {
					return nil, err
				}

				resultSet := NewResultSet()
				for res := results.List(); res != nil; res = res.Next() {
					resultSet.Add(res.ObjectType, res.ObjectId, relation, res.Warrant, res.Policy, res.IsImplicit || level > 0)
				}

				return resultSet, nil
			} else {
				userset, err := svc.listWarrants(ctx, warrant.FilterParams{
					ObjectType:  query.SelectSubjects.ForObject.Type,
					ObjectId:    query.SelectSubjects.ForObject.Id,
					Relation:    rule.WithRelation,
					SubjectType: rule.OfType,
				})
				if err != nil {
					return nil, err
				}

				resultSet := NewResultSet()
				for _, w := range userset {
					if w.Subject.Relation != "" {
						continue
					}

					subset, err := svc.query(ctx, Query{
						Expand: query.Expand,
						SelectSubjects: &SelectSubjects{
							SubjectTypes: query.SelectSubjects.SubjectTypes,
							Relations:    []string{rule.InheritIf},
							ForObject: &Resource{
								Type: w.Subject.ObjectType,
								Id:   w.Subject.ObjectId,
							},
						},
						Context: query.Context,
					}, 0)
					if err != nil {
						return nil, err
					}

					for res := subset.List(); res != nil; res = res.Next() {
						resultSet.Add(res.ObjectType, res.ObjectId, relation, res.Warrant, w.Policy.Or(res.Policy), res.IsImplicit || level > 0)
					}
				}

				return resultSet, nil
			}
		default:
			return nil, ErrInvalidQuery
		}
	}
}

func (svc QueryService) listWarrants(ctx context.Context, filterParams warrant.FilterParams) ([]warrant.WarrantSpec, error) {
	var result []warrant.WarrantSpec
	listParams := service.DefaultListParams(warrant.WarrantListParamParser{})
	listParams.WithLimit(MaxEdges)
	for {
		warrantSpecs, _, nextCursor, err := svc.warrantSvc.List(
			ctx,
			filterParams,
			listParams,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, warrantSpecs...)

		if nextCursor == nil {
			return result, nil
		}

		listParams.NextCursor = nextCursor
	}
}

func (svc QueryService) listObjectsByType(ctx context.Context, objectType string) ([]object.ObjectSpec, error) {
	var result []object.ObjectSpec
	listParams := service.DefaultListParams(object.ObjectListParamParser{})
	listParams.WithLimit(MaxEdges)
	for {
		objectSpecs, _, nextCursor, err := svc.objectSvc.List(
			ctx,
			&object.FilterOptions{ObjectType: objectType},
			listParams,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, objectSpecs...)

		if nextCursor == nil {
			return result, nil
		}

		listParams.NextCursor = nextCursor
	}
}

func objectKey(objectType string, objectId string) string {
	return fmt.Sprintf("%s:%s", objectType, objectId)
}

func objectRelationKey(objectType string, objectId string, relation string) string {
	return fmt.Sprintf("%s:%s#%s", objectType, objectId, relation)
}

func objectTypeAndObjectIdAndRelationFromCursor(cursor *service.Cursor) (string, string, string, error) {
	objectType, objectIdRelation, found := strings.Cut(cursor.ID(), ":")
	if !found {
		return "", "", "", errors.New("invalid cursor")
	}

	objectId, relation, found := strings.Cut(objectIdRelation, "#")
	if !found {
		return "", "", "", errors.New("invalid cursor")
	}

	return objectType, objectId, relation, nil
}
