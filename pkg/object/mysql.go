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

package object

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
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

func (repo MySQLRepository) Create(ctx context.Context, model Model) (int64, error) {
	result, err := repo.DB.ExecContext(
		ctx,
		`
			INSERT INTO object (
				objectType,
				objectId,
				meta
			) VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE
				meta = ?,
				createdAt = IF(object.deletedAt IS NULL, object.createdAt, CURRENT_TIMESTAMP(6)),
				updatedAt = CURRENT_TIMESTAMP(6),
				deletedAt = NULL
		`,
		model.GetObjectType(),
		model.GetObjectId(),
		model.GetMeta(),
		model.GetMeta(),
	)
	if err != nil {
		return -1, errors.Wrap(err, "error creating object")
	}

	newObjectId, err := result.LastInsertId()
	if err != nil {
		return -1, errors.Wrap(err, "error creating object")
	}

	return newObjectId, nil
}

func (repo MySQLRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var object Object
	err := repo.DB.GetContext(
		ctx,
		&object,
		`
			SELECT id, objectType, objectId, meta, createdAt, updatedAt, deletedAt
			FROM object
			WHERE
				id = ? AND
				deletedAt IS NULL
		`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.NewRecordNotFoundError("Object", id)
		} else {
			return nil, errors.Wrapf(err, "error getting object %d", id)
		}
	}

	return &object, nil
}

func (repo MySQLRepository) GetByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (Model, error) {
	var object Object
	err := repo.DB.GetContext(
		ctx,
		&object,
		`
			SELECT id, objectType, objectId, meta, createdAt, updatedAt, deletedAt
			FROM object
			WHERE
				objectType = ? AND
				objectId = ? AND
				deletedAt IS NULL
		`,
		objectType,
		objectId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.NewRecordNotFoundError(objectType, objectId)
		} else {
			return nil, errors.Wrapf(err, "error getting object %s:%s", objectType, objectId)
		}
	}

	return &object, nil
}

func (repo MySQLRepository) BatchGetByObjectTypeAndIds(ctx context.Context, objectType string, objectIds []string) ([]Model, error) {
	models := make([]Model, 0)
	objects := make([]Object, 0)
	if len(objectIds) == 0 {
		return models, nil
	}

	query, args, err := sqlx.In(
		`
			SELECT id, objectType, objectId, meta, createdAt, updatedAt, deletedAt
			FROM object
			WHERE
				objectType = ? AND
				objectId IN (?) AND
				deletedAt IS NULL
			ORDER BY objectId ASC
		`,
		objectType,
		objectIds,
	)
	if err != nil {
		return models, errors.Wrap(err, "error getting objects batch")
	}

	err = repo.DB.SelectContext(
		ctx,
		&objects,
		query,
		args...,
	)
	if err != nil {
		return models, errors.Wrap(err, "error getting objects batch")
	}

	for i := range objects {
		models = append(models, &objects[i])
	}

	return models, nil
}

func (repo MySQLRepository) List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]Model, *service.Cursor, *service.Cursor, error) {
	models := make([]Model, 0)
	objects := make([]Object, 0)
	query := `
		SELECT id, objectType, objectId, meta, createdAt, updatedAt, deletedAt
		FROM object
		WHERE
			deletedAt IS NULL
	`
	replacements := []interface{}{}

	var sortByColumn string
	if IsObjectSortBy(listParams.SortBy) {
		sortByColumn = listParams.SortBy
	} else {
		sortByColumn = fmt.Sprintf("meta->>'$.%s'", listParams.SortBy)
	}

	if filterOptions != nil && filterOptions.ObjectType != "" {
		query = fmt.Sprintf("%s AND objectType = ?", query)
		replacements = append(replacements, filterOptions.ObjectType)
	}

	if listParams.Query != nil {
		searchTermReplacement := fmt.Sprintf("%%%s%%", *listParams.Query)
		query = fmt.Sprintf("%s AND (objectId LIKE ? OR meta LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.NextCursor != nil {
		comparisonOp := "<"
		if listParams.SortOrder == service.SortOrderAsc {
			comparisonOp = ">"
		}

		switch listParams.NextCursor.Value() {
		case nil:
			//nolint:gocritic
			if listParams.SortBy == primarySortKey {
				query = fmt.Sprintf("%s AND %s %s ?", query, primarySortKey, comparisonOp)
				replacements = append(replacements, listParams.NextCursor.ID())
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, sortByColumn, primarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements, listParams.NextCursor.ID())
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, primarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements, listParams.NextCursor.ID())
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, primarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.NextCursor.Value(),
					listParams.NextCursor.ID(),
					listParams.NextCursor.Value(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, sortByColumn, primarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.NextCursor.Value(),
					listParams.NextCursor.ID(),
					listParams.NextCursor.Value(),
				)
			}
		}
	}

	if listParams.PrevCursor != nil {
		comparisonOp := ">"
		if listParams.SortOrder == service.SortOrderAsc {
			comparisonOp = "<"
		}

		switch listParams.PrevCursor.Value() {
		case nil:
			//nolint:gocritic
			if listParams.SortBy == primarySortKey {
				query = fmt.Sprintf("%s AND %s %s ?", query, primarySortKey, comparisonOp)
				replacements = append(replacements, listParams.PrevCursor.ID())
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, primarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements, listParams.PrevCursor.ID())
			} else {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, sortByColumn, primarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements, listParams.PrevCursor.ID())
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, sortByColumn, primarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.PrevCursor.Value(),
					listParams.PrevCursor.ID(),
					listParams.PrevCursor.Value(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, primarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.PrevCursor.Value(),
					listParams.PrevCursor.ID(),
					listParams.PrevCursor.Value(),
				)
			}
		}
	}

	if listParams.PrevCursor != nil {
		if listParams.SortBy != primarySortKey {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, sortByColumn, service.SortOrderDesc, primarySortKey, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit+1)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, sortByColumn, service.SortOrderAsc, primarySortKey, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit+1)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s, %s %s", query, sortByColumn, listParams.SortOrder, primarySortKey, listParams.SortOrder)
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, sortByColumn, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit+1)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, sortByColumn, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit+1)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s", query, sortByColumn, listParams.SortOrder)
		}
	} else {
		if listParams.SortBy != primarySortKey {
			query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, sortByColumn, listParams.SortOrder, primarySortKey, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit+1)
		} else {
			query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, primarySortKey, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit+1)
		}
	}

	err := repo.DB.SelectContext(
		ctx,
		&objects,
		query,
		replacements...,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models, nil, nil, nil
		}
		return nil, nil, nil, errors.Wrap(err, "error listing warrants")
	}

	if len(objects) == 0 {
		return models, nil, nil, nil
	}

	for i := 0; i < len(objects) && i < listParams.Limit; i++ {
		models = append(models, &objects[i])
	}

	//nolint:gosec
	firstElem := models[0]
	lastElem := models[len(models)-1]
	var firstValue interface{} = nil
	var lastValue interface{} = nil
	switch listParams.SortBy {
	case primarySortKey:
		// do nothing
	case "createdAt":
		firstValue = firstElem.GetCreatedAt()
		lastValue = lastElem.GetCreatedAt()
	case "objectType":
		firstValue = firstElem.GetObjectType()
		lastValue = lastElem.GetObjectType()
	default:
		firstSpec, err := firstElem.ToObjectSpec()
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "error listing warrants")
		}

		lastSpec, err := firstElem.ToObjectSpec()
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "error listing warrants")
		}

		firstValue = firstSpec.Meta[listParams.SortBy]
		lastValue = lastSpec.Meta[listParams.SortBy]
	}

	prevCursor := service.NewCursor(firstElem.GetObjectId(), firstValue)
	nextCursor := service.NewCursor(lastElem.GetObjectId(), lastValue)
	if len(objects) <= listParams.Limit {
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

func (repo MySQLRepository) UpdateByObjectTypeAndId(ctx context.Context, objectType string, objectId string, model Model) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE object
			SET
				meta = ?,
				updatedAt = CURRENT_TIMESTAMP(6)
			WHERE
				objectType = ? AND
				objectId = ? AND
				deletedAt IS NULL
		`,
		model.GetMeta(),
		objectType,
		objectId,
	)
	if err != nil {
		return errors.Wrapf(err, "error updating object %s:%s", objectType, objectId)
	}

	return nil
}

func (repo MySQLRepository) DeleteByObjectTypeAndId(ctx context.Context, objectType string, objectId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE object
			SET
				updatedAt = CURRENT_TIMESTAMP(6),
				deletedAt = CURRENT_TIMESTAMP(6)
			WHERE
				objectType = ? AND
				objectId = ? AND
				deletedAt IS NULL
		`,
		objectType,
		objectId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		} else {
			return errors.Wrapf(err, "error deleting object %s:%s", objectType, objectId)
		}
	}

	return nil
}

func (repo MySQLRepository) DeleteWarrantsMatchingObject(ctx context.Context, objectType string, objectId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE warrant
			SET
				updatedAt = CURRENT_TIMESTAMP(6),
				deletedAt = CURRENT_TIMESTAMP(6)
			WHERE
				objectType = ? AND
				objectId = ? AND
				deletedAt IS NULL
		`,
		objectType,
		objectId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		} else {
			return errors.Wrapf(err, "error deleting warrants matching object %s:%s", objectType, objectId)
		}
	}

	return nil
}

func (repo MySQLRepository) DeleteWarrantsMatchingSubject(ctx context.Context, subjectType string, subjectId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE warrant
			SET
				updatedAt = CURRENT_TIMESTAMP(6),
				deletedAt = CURRENT_TIMESTAMP(6)
			WHERE
				subjectType = ? AND
				subjectId = ? AND
				deletedAt IS NULL
		`,
		subjectType,
		subjectId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		} else {
			return errors.Wrapf(err, "error deleting warrants matching subject %s:%s", subjectType, subjectId)
		}
	}

	return nil
}
