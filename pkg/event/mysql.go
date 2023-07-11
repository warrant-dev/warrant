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

type MySQLRepository struct {
	database.SQLRepository
}

func NewMySQLRepository(db *database.MySQL) *MySQLRepository {
	return &MySQLRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo MySQLRepository) TrackResourceEvent(ctx context.Context, resourceEvent ResourceEventModel) error {
	return repo.TrackResourceEvents(ctx, []ResourceEventModel{resourceEvent})
}

func (repo MySQLRepository) TrackResourceEvents(ctx context.Context, models []ResourceEventModel) error {
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
			  meta
		   ) VALUES (
			  UUID_TO_BIN(:id),
			  :type,
			  :source,
			  :resourceType,
			  :resourceId,
			  :meta
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

func (repo MySQLRepository) ListResourceEvents(ctx context.Context, listParams ListResourceEventParams) ([]ResourceEventModel, string, error) {
	models := make([]ResourceEventModel, 0)
	resourceEvents := make([]ResourceEvent, 0)
	query := `
		SELECT BIN_TO_UUID(id) id, type, source, resourceType, resourceId, meta, createdAt
		FROM resourceEvent
		WHERE
	`
	conditions := []string{}
	replacements := []interface{}{}

	if listParams.Type != "" {
		conditions = append(conditions, "type = ?")
		replacements = append(replacements, listParams.Type)
	}

	if listParams.Source != "" {
		conditions = append(conditions, "source = ?")
		replacements = append(replacements, listParams.Source)
	}

	if listParams.ResourceType != "" {
		conditions = append(conditions, "resourceType = ?")
		replacements = append(replacements, listParams.ResourceType)
	}

	if listParams.ResourceId != "" {
		conditions = append(conditions, "resourceId = ?")
		replacements = append(replacements, listParams.ResourceId)
	}

	if listParams.LastId != "" {
		lastIdSpec, err := StringToLastIdSpec(listParams.LastId)
		if err != nil {
			return models, "", service.NewInvalidParameterError("lastId", "")
		}

		conditions = append(conditions, "(createdAt, BIN_TO_UUID(id)) < (?, ?)")
		replacements = append(replacements, lastIdSpec.CreatedAt)
		replacements = append(replacements, lastIdSpec.ID)
	}

	conditions = append(conditions, "createdAt BETWEEN ? AND ?")
	replacements = append(replacements, listParams.Since)
	replacements = append(replacements, listParams.Until)

	query = fmt.Sprintf("%s %s ORDER BY createdAt DESC, BIN_TO_UUID(id) DESC LIMIT ?", query, strings.Join(conditions, " AND "))
	replacements = append(replacements, listParams.Limit)
	err := repo.DB.SelectContext(
		ctx,
		&resourceEvents,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, "", nil
		default:
			return models, "", errors.Wrap(err, "error listing resource events")
		}
	}

	for i := range resourceEvents {
		models = append(models, &resourceEvents[i])
	}

	if len(resourceEvents) == 0 || len(resourceEvents) < int(listParams.Limit) {
		return models, "", nil
	}

	lastResourceEvent := resourceEvents[len(resourceEvents)-1]
	lastIdStr, err := LastIdSpecToString(LastIdSpec{
		ID:        lastResourceEvent.GetID(),
		CreatedAt: lastResourceEvent.GetCreatedAt(),
	})
	if err != nil {
		return models, "", errors.Wrap(err, "error listing resource events")
	}

	return models, lastIdStr, nil
}

func (repo MySQLRepository) TrackAccessEvent(ctx context.Context, accessEvent AccessEventModel) error {
	return repo.TrackAccessEvents(ctx, []AccessEventModel{accessEvent})
}

func (repo MySQLRepository) TrackAccessEvents(ctx context.Context, models []AccessEventModel) error {
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
			  meta
		   ) VALUES (
			  UUID_TO_BIN(:id),
			  :type,
			  :source,
			  :objectType,
			  :objectId,
			  :relation,
			  :subjectType,
			  :subjectId,
			  :subjectRelation,
			  :meta
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

func (repo MySQLRepository) ListAccessEvents(ctx context.Context, listParams ListAccessEventParams) ([]AccessEventModel, string, error) {
	models := make([]AccessEventModel, 0)
	accessEvents := make([]AccessEvent, 0)
	query := `
		SELECT BIN_TO_UUID(id) id, type, source, objectType, objectId, relation, subjectType, subjectId, subjectRelation, meta, createdAt
		FROM accessEvent
		WHERE
	`
	conditions := []string{}
	replacements := []interface{}{}

	if listParams.Type != "" {
		conditions = append(conditions, "type = ?")
		replacements = append(replacements, listParams.Type)
	}

	if listParams.Source != "" {
		conditions = append(conditions, "source = ?")
		replacements = append(replacements, listParams.Source)
	}

	if listParams.ObjectType != "" {
		conditions = append(conditions, "objectType = ?")
		replacements = append(replacements, listParams.ObjectType)
	}

	if listParams.ObjectId != "" {
		conditions = append(conditions, "objectId = ?")
		replacements = append(replacements, listParams.ObjectId)
	}

	if listParams.Relation != "" {
		conditions = append(conditions, "relation = ?")
		replacements = append(replacements, listParams.Relation)
	}

	if listParams.SubjectType != "" {
		conditions = append(conditions, "subjectType = ?")
		replacements = append(replacements, listParams.SubjectType)
	}

	if listParams.SubjectId != "" {
		conditions = append(conditions, "subjectId = ?")
		replacements = append(replacements, listParams.SubjectId)
	}

	if listParams.SubjectRelation != "" {
		conditions = append(conditions, "subjectRelation = ?")
		replacements = append(replacements, listParams.SubjectRelation)
	}

	if listParams.LastId != "" {
		lastIdSpec, err := StringToLastIdSpec(listParams.LastId)
		if err != nil {
			return models, "", service.NewInvalidParameterError("lastId", "")
		}

		conditions = append(conditions, "(createdAt, BIN_TO_UUID(id)) < (?, ?)")
		replacements = append(replacements, lastIdSpec.CreatedAt)
		replacements = append(replacements, lastIdSpec.ID)
	}

	conditions = append(conditions, "createdAt BETWEEN ? AND ?")
	replacements = append(replacements, listParams.Since)
	replacements = append(replacements, listParams.Until)

	query = fmt.Sprintf("%s %s ORDER BY createdAt DESC, BIN_TO_UUID(id) DESC LIMIT ?", query, strings.Join(conditions, " AND "))
	replacements = append(replacements, listParams.Limit)
	err := repo.DB.SelectContext(
		ctx,
		&accessEvents,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, "", nil
		default:
			return models, "", errors.Wrap(err, "error listing access events")
		}
	}

	for i := range accessEvents {
		models = append(models, &accessEvents[i])
	}

	if len(accessEvents) == 0 || len(accessEvents) < int(listParams.Limit) {
		return models, "", nil
	}

	lastAccessEvent := accessEvents[len(accessEvents)-1]
	lastIdStr, err := LastIdSpecToString(LastIdSpec{
		ID:        lastAccessEvent.GetID(),
		CreatedAt: lastAccessEvent.GetCreatedAt(),
	})
	if err != nil {
		return models, "", errors.Wrap(err, "error listing access events")
	}

	return models, lastIdStr, nil
}
