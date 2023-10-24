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

package event

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

type SQLiteRepository struct {
	database.SQLRepository
}

func NewSQLiteRepository(db *database.SQLite) *SQLiteRepository {
	return &SQLiteRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo SQLiteRepository) TrackResourceEvent(ctx context.Context, resourceEvent ResourceEventModel) error {
	return repo.TrackResourceEvents(ctx, []ResourceEventModel{resourceEvent})
}

func (repo SQLiteRepository) TrackResourceEvents(ctx context.Context, models []ResourceEventModel) error {
	resourceEvents := make([]ResourceEvent, 0)
	for _, model := range models {
		resourceEvents = append(resourceEvents, *NewResourceEventFromModel(model))
	}

	result, err := repo.DB.NamedExecContext(
		ctx,
		`
		   INSERT INTO resourceEvent (
			  id,
			  type,
			  source,
			  resourceType,
			  resourceId,
			  meta,
			  createdAt
		   ) VALUES (
			  :id,
			  :type,
			  :source,
			  :resourceType,
			  :resourceId,
			  :meta,
			  :createdAt
		   )
		`,
		resourceEvents,
	)
	if err != nil {
		return errors.Wrap(err, "error creating resource events")
	}

	_, err = result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "error creating resource events")
	}

	return nil
}

func (repo SQLiteRepository) ListResourceEvents(ctx context.Context, filterParams ResourceEventFilterParams, listParams service.ListParams) ([]ResourceEventModel, *service.Cursor, *service.Cursor, error) {
	models := make([]ResourceEventModel, 0)
	resourceEvents := make([]ResourceEvent, 0)
	query := `
		SELECT id, type, source, resourceType, resourceId, meta, createdAt
		FROM resourceEvent
		WHERE
	`
	conditions := []string{}
	replacements := []interface{}{}

	if filterParams.Type != "" {
		conditions = append(conditions, "type = ?")
		replacements = append(replacements, filterParams.Type)
	}

	if filterParams.Source != "" {
		conditions = append(conditions, "source = ?")
		replacements = append(replacements, filterParams.Source)
	}

	if filterParams.ResourceType != "" {
		conditions = append(conditions, "resourceType = ?")
		replacements = append(replacements, filterParams.ResourceType)
	}

	if filterParams.ResourceId != "" {
		conditions = append(conditions, "resourceId = ?")
		replacements = append(replacements, filterParams.ResourceId)
	}

	conditions = append(conditions, "createdAt BETWEEN ? AND ?")
	replacements = append(replacements, filterParams.Since)
	replacements = append(replacements, filterParams.Until)

	if listParams.NextCursor != nil {
		conditions = append(conditions, "(createdAt, id) < (?, ?)")
		replacements = append(replacements, listParams.NextCursor.Value())
		replacements = append(replacements, listParams.NextCursor.ID())
	}

	if listParams.PrevCursor != nil {
		conditions = append(conditions, "(createdAt, id) > (?, ?)")
		replacements = append(replacements, listParams.PrevCursor.Value())
		replacements = append(replacements, listParams.PrevCursor.ID())
		query = fmt.Sprintf("With result_set AS (%s %s ORDER BY createdAt ASC, id ASC LIMIT ?) SELECT * FROM result_set ORDER BY createdAt DESC, id DESC", query, strings.Join(conditions, " AND "))
		replacements = append(replacements, listParams.Limit+1)
	} else {
		query = fmt.Sprintf("%s %s ORDER BY createdAt DESC, id DESC LIMIT ?", query, strings.Join(conditions, " AND "))
		replacements = append(replacements, listParams.Limit+1)
	}

	err := repo.DB.SelectContext(
		ctx,
		&resourceEvents,
		query,
		replacements...,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models, nil, nil, nil
		}
		return nil, nil, nil, errors.Wrap(err, "error listing resource events")
	}

	if len(resourceEvents) == 0 {
		return models, nil, nil, nil
	}

	i := 0
	if listParams.PrevCursor != nil && len(resourceEvents) > listParams.Limit {
		i = 1
	}
	for i < len(resourceEvents) && len(models) < listParams.Limit {
		models = append(models, resourceEvents[i])
		i++
	}

	firstElem := models[0]
	lastElem := models[len(models)-1]
	var firstValue interface{} = nil
	var lastValue interface{} = nil
	switch listParams.SortBy {
	case "id":
		// do nothing
	case "createdAt":
		firstValue = firstElem.GetCreatedAt()
		lastValue = lastElem.GetCreatedAt()
	}

	prevCursor := service.NewCursor(firstElem.GetID(), firstValue)
	nextCursor := service.NewCursor(lastElem.GetID(), lastValue)
	if len(resourceEvents) <= listParams.Limit {
		if listParams.PrevCursor != nil {
			return models, nil, nextCursor, nil
		}

		if listParams.NextCursor != nil {
			return models, prevCursor, nil, nil
		}

		return models, nil, nil, nil
	} else if listParams.PrevCursor == nil && listParams.NextCursor == nil {
		return models, nil, nextCursor, nil
	}

	return models, prevCursor, nextCursor, nil
}

func (repo SQLiteRepository) TrackAccessEvent(ctx context.Context, accessEvent AccessEventModel) error {
	return repo.TrackAccessEvents(ctx, []AccessEventModel{accessEvent})
}

func (repo SQLiteRepository) TrackAccessEvents(ctx context.Context, models []AccessEventModel) error {
	accessEvents := make([]AccessEvent, 0)
	for _, model := range models {
		accessEvents = append(accessEvents, *NewAccessEventFromModel(model))
	}

	result, err := repo.DB.NamedExecContext(
		ctx,
		`
		   INSERT INTO accessEvent (
			  id,
			  type,
			  source,
			  objectType,
			  objectId,
			  relation,
			  subjectType,
			  subjectId,
			  subjectRelation,
			  meta,
			  createdAt
		   ) VALUES (
			  :id,
			  :type,
			  :source,
			  :objectType,
			  :objectId,
			  :relation,
			  :subjectType,
			  :subjectId,
			  :subjectRelation,
			  :meta,
			  :createdAt
		   )
		`,
		accessEvents,
	)
	if err != nil {
		return errors.Wrap(err, "error creating access events")
	}

	_, err = result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "error creating access events")
	}

	return nil
}

func (repo SQLiteRepository) ListAccessEvents(ctx context.Context, filterParams AccessEventFilterParams, listParams service.ListParams) ([]AccessEventModel, *service.Cursor, *service.Cursor, error) {
	models := make([]AccessEventModel, 0)
	accessEvents := make([]AccessEvent, 0)
	query := `
		SELECT id, type, source, objectType, objectId, relation, subjectType, subjectId, subjectRelation, meta, createdAt
		FROM accessEvent
		WHERE
	`
	conditions := []string{}
	replacements := []interface{}{}

	if filterParams.Type != "" {
		conditions = append(conditions, "type = ?")
		replacements = append(replacements, filterParams.Type)
	}

	if filterParams.Source != "" {
		conditions = append(conditions, "source = ?")
		replacements = append(replacements, filterParams.Source)
	}

	if filterParams.ObjectType != "" {
		conditions = append(conditions, "objectType = ?")
		replacements = append(replacements, filterParams.ObjectType)
	}

	if filterParams.ObjectId != "" {
		conditions = append(conditions, "objectId = ?")
		replacements = append(replacements, filterParams.ObjectId)
	}

	if filterParams.Relation != "" {
		conditions = append(conditions, "relation = ?")
		replacements = append(replacements, filterParams.Relation)
	}

	if filterParams.SubjectType != "" {
		conditions = append(conditions, "subjectType = ?")
		replacements = append(replacements, filterParams.SubjectType)
	}

	if filterParams.SubjectId != "" {
		conditions = append(conditions, "subjectId = ?")
		replacements = append(replacements, filterParams.SubjectId)
	}

	if filterParams.SubjectRelation != "" {
		conditions = append(conditions, "subjectRelation = ?")
		replacements = append(replacements, filterParams.SubjectRelation)
	}

	conditions = append(conditions, "createdAt BETWEEN ? AND ?")
	replacements = append(replacements, filterParams.Since)
	replacements = append(replacements, filterParams.Until)

	if listParams.NextCursor != nil {
		conditions = append(conditions, "(createdAt, id) < (?, ?)")
		replacements = append(replacements, listParams.NextCursor.Value())
		replacements = append(replacements, listParams.NextCursor.ID())
	}

	if listParams.PrevCursor != nil {
		conditions = append(conditions, "(createdAt, id) > (?, ?)")
		replacements = append(replacements, listParams.PrevCursor.Value())
		replacements = append(replacements, listParams.PrevCursor.ID())
		query = fmt.Sprintf("With result_set AS (%s %s ORDER BY createdAt ASC, id ASC LIMIT ?) SELECT * FROM result_set ORDER BY createdAt DESC, id DESC", query, strings.Join(conditions, " AND "))
		replacements = append(replacements, listParams.Limit+1)
	} else {
		query = fmt.Sprintf("%s %s ORDER BY createdAt DESC, id DESC LIMIT ?", query, strings.Join(conditions, " AND "))
		replacements = append(replacements, listParams.Limit+1)
	}

	err := repo.DB.SelectContext(
		ctx,
		&accessEvents,
		query,
		replacements...,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models, nil, nil, nil
		}
		return nil, nil, nil, errors.Wrap(err, "error listing access events")
	}

	if len(accessEvents) == 0 {
		return models, nil, nil, nil
	}

	i := 0
	if listParams.PrevCursor != nil && len(accessEvents) > listParams.Limit {
		i = 1
	}
	for i < len(accessEvents) && len(models) < listParams.Limit {
		models = append(models, accessEvents[i])
		i++
	}

	firstElem := models[0]
	lastElem := models[len(models)-1]
	var firstValue interface{} = nil
	var lastValue interface{} = nil
	switch listParams.SortBy {
	case "id":
		// do nothing
	case "createdAt":
		firstValue = firstElem.GetCreatedAt()
		lastValue = lastElem.GetCreatedAt()
	}

	prevCursor := service.NewCursor(firstElem.GetID(), firstValue)
	nextCursor := service.NewCursor(lastElem.GetID(), lastValue)
	if len(accessEvents) <= listParams.Limit {
		if listParams.PrevCursor != nil {
			return models, nil, nextCursor, nil
		}

		if listParams.NextCursor != nil {
			return models, prevCursor, nil, nil
		}

		return models, nil, nil, nil
	} else if listParams.PrevCursor == nil && listParams.NextCursor == nil {
		return models, nil, nextCursor, nil
	}

	return models, prevCursor, nextCursor, nil
}
