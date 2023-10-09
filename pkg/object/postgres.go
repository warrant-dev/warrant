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
	"regexp"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

var sortRegexp = regexp.MustCompile("([A-Z])")

type PostgresRepository struct {
	database.SQLRepository
}

func NewPostgresRepository(db *database.Postgres) *PostgresRepository {
	return &PostgresRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo PostgresRepository) Create(ctx context.Context, model Model) (int64, error) {
	var newObjectId int64
	err := repo.DB.GetContext(
		database.CtxWithWriterOverride(ctx),
		&newObjectId,
		`
			INSERT INTO object (
				object_type,
				object_id,
				meta
			) VALUES (?, ?, ?)
			ON CONFLICT (object_type, object_id) DO UPDATE SET
				meta = ?,
				created_at = CASE
					WHEN object.deleted_at IS NULL THEN object.created_at
					ELSE CURRENT_TIMESTAMP(6)
				END,
				updated_at = CURRENT_TIMESTAMP(6),
				deleted_at = NULL
			RETURNING id
		`,
		model.GetObjectType(),
		model.GetObjectId(),
		model.GetMeta(),
		model.GetMeta(),
	)
	if err != nil {
		return -1, errors.Wrap(err, "error creating object")
	}

	return newObjectId, nil
}

func (repo PostgresRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var object Object
	err := repo.DB.GetContext(
		ctx,
		&object,
		`
			SELECT id, object_type, object_id, meta, created_at, updated_at, deleted_at
			FROM object
			WHERE
				id = ? AND
				deleted_at IS NULL
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

func (repo PostgresRepository) GetByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (Model, error) {
	var object Object
	err := repo.DB.GetContext(
		ctx,
		&object,
		`
			SELECT id, object_type, object_id, meta, created_at, updated_at, deleted_at
			FROM object
			WHERE
				object_type = ? AND
				object_id = ? AND
				deleted_at IS NULL
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

func (repo PostgresRepository) BatchGetByObjectTypeAndIds(ctx context.Context, objectType string, objectIds []string) ([]Model, error) {
	models := make([]Model, 0)
	objects := make([]Object, 0)
	if len(objectIds) == 0 {
		return models, nil
	}

	query, args, err := sqlx.In(
		`
			SELECT id, object_type, object_id, meta, created_at, updated_at, deleted_at
			FROM object
			WHERE
				object_type = ? AND
				object_id IN (?) AND
				deleted_at IS NULL
			ORDER BY object_id ASC
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

func (repo PostgresRepository) List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]Model, error) {
	models := make([]Model, 0)
	objects := make([]Object, 0)
	query := `
		SELECT id, object_type, object_id, meta, created_at, updated_at, deleted_at
		FROM object
		WHERE
			deleted_at IS NULL
	`
	replacements := []interface{}{}
	objectIdColumn := sortRegexp.ReplaceAllString("objectId", `_$1`)

	var sortByColumn string
	if IsObjectSortBy(listParams.SortBy) {
		sortByColumn = sortRegexp.ReplaceAllString(listParams.SortBy, `_$1`)
	} else {
		sortByColumn = fmt.Sprintf("meta->>'%s'", sortRegexp.ReplaceAllString(listParams.SortBy, `_$1`))
	}

	if filterOptions != nil && filterOptions.ObjectType != "" {
		query = fmt.Sprintf("%s AND object_type = ?", query)
		replacements = append(replacements, filterOptions.ObjectType)
	}

	if listParams.Query != nil {
		searchTermReplacement := fmt.Sprintf("%%%s%%", *listParams.Query)
		query = fmt.Sprintf("%s AND (%s LIKE ? OR meta::text LIKE ?)", query, objectIdColumn)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.AfterId != nil {
		comparisonOp := "<"
		if listParams.SortOrder == service.SortOrderAsc {
			comparisonOp = ">"
		}

		switch listParams.AfterValue {
		case nil:
			//nolint:gocritic
			if listParams.SortBy == listParams.DefaultSortBy() {
				query = fmt.Sprintf("%s AND %s %s ?", query, objectIdColumn, comparisonOp)
				replacements = append(replacements, listParams.AfterId)
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, sortByColumn, objectIdColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.AfterId,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, objectIdColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.AfterId,
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, objectIdColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, sortByColumn, objectIdColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		}
	}

	if listParams.BeforeId != nil {
		comparisonOp := ">"
		if listParams.SortOrder == service.SortOrderAsc {
			comparisonOp = "<"
		}

		switch listParams.BeforeValue {
		case nil:
			//nolint:gocritic
			if listParams.SortBy == listParams.DefaultSortBy() {
				query = fmt.Sprintf("%s AND %s %s ?", query, objectIdColumn, comparisonOp)
				replacements = append(replacements, listParams.BeforeId)
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, objectIdColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.BeforeId,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, sortByColumn, objectIdColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.BeforeId,
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, sortByColumn, objectIdColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, objectIdColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		}
	}

	nullSortClause := "NULLS LAST"
	invertedNullSortClause := "NULLS FIRST"
	if listParams.SortOrder == service.SortOrderAsc {
		nullSortClause = "NULLS FIRST"
		invertedNullSortClause = "NULLS LAST"
	}

	if listParams.BeforeId != nil {
		if listParams.SortBy != listParams.DefaultSortBy() {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortByColumn, service.SortOrderDesc, invertedNullSortClause, objectIdColumn, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortByColumn, service.SortOrderAsc, invertedNullSortClause, objectIdColumn, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s %s, %s %s", query, sortByColumn, listParams.SortOrder, nullSortClause, objectIdColumn, listParams.SortOrder)
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, sortByColumn, service.SortOrderDesc, invertedNullSortClause)
				replacements = append(replacements, listParams.Limit)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, sortByColumn, service.SortOrderAsc, invertedNullSortClause)
				replacements = append(replacements, listParams.Limit)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s %s", query, sortByColumn, listParams.SortOrder, nullSortClause)
		}
	} else {
		if listParams.SortBy != listParams.DefaultSortBy() {
			query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortByColumn, listParams.SortOrder, nullSortClause, objectIdColumn, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, objectIdColumn, listParams.SortOrder, nullSortClause)
			replacements = append(replacements, listParams.Limit)
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
			return models, nil
		} else {
			return nil, errors.Wrap(err, "error listing objects")
		}
	}

	for i := range objects {
		models = append(models, &objects[i])
	}

	return models, nil
}

func (repo PostgresRepository) UpdateByObjectTypeAndId(ctx context.Context, objectType string, objectId string, model Model) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE object
			SET
				meta = ?,
				updated_at = CURRENT_TIMESTAMP(6)
			WHERE
				object_type = ? AND
				object_id = ? AND
				deleted_at IS NULL
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

func (repo PostgresRepository) DeleteByObjectTypeAndId(ctx context.Context, objectType string, objectId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE object
			SET
				updated_at = CURRENT_TIMESTAMP(6),
				deleted_at = CURRENT_TIMESTAMP(6)
			WHERE
				object_type = ? AND
				object_id = ? AND
				deleted_at IS NULL
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

func (repo PostgresRepository) DeleteWarrantsMatchingObject(ctx context.Context, objectType string, objectId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE warrant
			SET
				updated_at = CURRENT_TIMESTAMP(6),
				deleted_at = CURRENT_TIMESTAMP(6)
			WHERE
				object_type = ? AND
				object_id = ? AND
				deleted_at IS NULL
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

func (repo PostgresRepository) DeleteWarrantsMatchingSubject(ctx context.Context, subjectType string, subjectId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE warrant
			SET
				updated_at = CURRENT_TIMESTAMP(6),
				deleted_at = CURRENT_TIMESTAMP(6)
			WHERE
				subject_type = ? AND
				subject_id = ? AND
				deleted_at IS NULL
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
