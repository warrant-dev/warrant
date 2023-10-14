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
	"time"

	"github.com/jmoiron/sqlx"
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

func (repo SQLiteRepository) Create(ctx context.Context, model Model) (int64, error) {
	var newObjectId int64
	now := time.Now().UTC()
	err := repo.DB.GetContext(
		database.CtxWithWriterOverride(ctx),
		&newObjectId,
		`
			INSERT INTO object (
				objectType,
				objectId,
				meta,
				createdAt,
				updatedAt
			) VALUES (?, ?, ?, ?, ?)
			ON CONFLICT (objectType, objectId) DO UPDATE SET
				meta = ?,
				createdAt = IIF(object.deletedAt IS NULL, object.createdAt, ?),
				updatedAt = ?,
				deletedAt = NULL
			RETURNING id
		`,
		model.GetObjectType(),
		model.GetObjectId(),
		model.GetMeta(),
		now,
		now,
		model.GetMeta(),
		now,
		now,
	)
	if err != nil {
		return -1, errors.Wrap(err, "error creating object")
	}

	return newObjectId, nil
}

func (repo SQLiteRepository) GetById(ctx context.Context, id int64) (Model, error) {
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

func (repo SQLiteRepository) GetByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (Model, error) {
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

func (repo SQLiteRepository) BatchGetByObjectTypeAndIds(ctx context.Context, objectType string, objectIds []string) ([]Model, error) {
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

func (repo SQLiteRepository) List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]Model, *service.Cursor, *service.Cursor, error) {
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
			if sortByColumn == PrimarySortKey {
				query = fmt.Sprintf("%s AND %s %s ?", query, PrimarySortKey, comparisonOp)
				replacements = append(replacements, listParams.NextCursor.ID())
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, sortByColumn, PrimarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.NextCursor.ID(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, PrimarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.NextCursor.ID(),
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, PrimarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.NextCursor.Value(),
					listParams.NextCursor.ID(),
					listParams.NextCursor.Value(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, sortByColumn, PrimarySortKey, comparisonOp, sortByColumn)
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
			if sortByColumn == PrimarySortKey {
				query = fmt.Sprintf("%s AND %s %s ?", query, PrimarySortKey, comparisonOp)
				replacements = append(replacements, listParams.PrevCursor.ID())
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, PrimarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.PrevCursor.ID(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, sortByColumn, PrimarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.PrevCursor.ID(),
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, sortByColumn, PrimarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.PrevCursor.Value(),
					listParams.PrevCursor.ID(),
					listParams.PrevCursor.Value(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, PrimarySortKey, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.PrevCursor.Value(),
					listParams.PrevCursor.ID(),
					listParams.PrevCursor.Value(),
				)
			}
		}
	}

	if listParams.PrevCursor != nil {
		if sortByColumn != PrimarySortKey {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, sortByColumn, service.SortOrderDesc, PrimarySortKey, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit+1)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, sortByColumn, service.SortOrderAsc, PrimarySortKey, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit+1)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s, %s %s", query, sortByColumn, listParams.SortOrder, PrimarySortKey, listParams.SortOrder)
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
		if sortByColumn != PrimarySortKey {
			query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, sortByColumn, listParams.SortOrder, PrimarySortKey, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit+1)
		} else {
			query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, PrimarySortKey, listParams.SortOrder)
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
		return nil, nil, nil, errors.Wrap(err, "error listing objects")
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
	switch sortByColumn {
	case PrimarySortKey:
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
			return nil, nil, nil, errors.Wrap(err, "error listing objects")
		}

		lastSpec, err := firstElem.ToObjectSpec()
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "error listing objects")
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

func (repo SQLiteRepository) UpdateByObjectTypeAndId(ctx context.Context, objectType string, objectId string, model Model) error {
	now := time.Now().UTC()
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE object
			SET
				meta = ?,
				updatedAt = ?
			WHERE
				objectType = ? AND
				objectId = ? AND
				deletedAt IS NULL
		`,
		model.GetMeta(),
		now,
		objectType,
		objectId,
	)
	if err != nil {
		return errors.Wrapf(err, "error updating object %s:%s", objectType, objectId)
	}

	return nil
}

func (repo SQLiteRepository) DeleteByObjectTypeAndId(ctx context.Context, objectType string, objectId string) error {
	now := time.Now().UTC()
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE object
			SET
				updatedAt = ?,
				deletedAt = ?
			WHERE
				objectType = ? AND
				objectId = ? AND
				deletedAt IS NULL
		`,
		now,
		now,
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

func (repo SQLiteRepository) DeleteWarrantsMatchingObject(ctx context.Context, objectType string, objectId string) error {
	now := time.Now().UTC()
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE warrant
			SET
				updatedAt = ?,
				deletedAt = ?
			WHERE
				objectType = ? AND
				objectId = ? AND
				deletedAt IS NULL
		`,
		now,
		now,
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

func (repo SQLiteRepository) DeleteWarrantsMatchingSubject(ctx context.Context, subjectType string, subjectId string) error {
	now := time.Now().UTC()
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE warrant
			SET
				updatedAt = ?,
				deletedAt = ?
			WHERE
				subjectType = ? AND
				subjectId = ? AND
				deletedAt IS NULL
		`,
		now,
		now,
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
