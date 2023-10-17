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
	"database/sql"
	"fmt"
	"regexp"

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
	var newObjectTypeId int64
	err := repo.DB.GetContext(
		database.CtxWithWriterOverride(ctx),
		&newObjectTypeId,
		`
			INSERT INTO object_type (
				type_id,
				definition
			) VALUES (?, ?)
			ON CONFLICT (type_id) DO UPDATE SET
				definition = ?,
				created_at = CASE
					WHEN object_type.deleted_at IS NULL THEN object_type.created_at
					ELSE CURRENT_TIMESTAMP(6)
				END,
				updated_at = CURRENT_TIMESTAMP(6),
				deleted_at = NULL
			RETURNING id
		`,
		model.GetTypeId(),
		model.GetDefinition(),
		model.GetDefinition(),
	)
	if err != nil {
		return -1, errors.Wrap(err, "error creating object type")
	}

	return newObjectTypeId, nil
}

func (repo PostgresRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var objectType ObjectType
	err := repo.DB.GetContext(
		ctx,
		&objectType,
		`
			SELECT id, type_id, definition, created_at, updated_at, deleted_at
			FROM object_type
			WHERE
				id = ? AND
				deleted_at IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return &objectType, service.NewRecordNotFoundError("ObjectType", id)
		default:
			return &objectType, errors.Wrapf(err, "error getting object type %d", id)
		}
	}

	return &objectType, nil
}

func (repo PostgresRepository) GetByTypeId(ctx context.Context, typeId string) (Model, error) {
	var objectType ObjectType
	err := repo.DB.GetContext(
		ctx,
		&objectType,
		`
			SELECT id, type_id, definition, created_at, updated_at, deleted_at
			FROM object_type
			WHERE
				type_id = ? AND
				deleted_at IS NULL
		`,
		typeId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &objectType, service.NewRecordNotFoundError("ObjectType", typeId)
		}
		return &objectType, errors.Wrapf(err, "error getting object type %s", typeId)
	}

	return &objectType, nil
}

func (repo PostgresRepository) ListAll(ctx context.Context) ([]Model, error) {
	models := make([]Model, 0)
	objectTypes := make([]ObjectType, 0)
	err := repo.DB.SelectContext(
		ctx,
		&objectTypes,
		`
			SELECT id, type_id, definition, created_at, updated_at, deleted_at
			FROM object_type
			WHERE
				deleted_at IS NULL
		`,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models, nil
		}

		return models, errors.Wrap(err, "error listing all object types")
	}

	for i := range objectTypes {
		models = append(models, &objectTypes[i])
	}

	return models, nil
}

func (repo PostgresRepository) List(ctx context.Context, listParams service.ListParams) ([]Model, *service.Cursor, *service.Cursor, error) {
	models := make([]Model, 0)
	objectTypes := make([]ObjectType, 0)
	replacements := make([]interface{}, 0)
	query := `
		SELECT id, type_id, definition, created_at, updated_at, deleted_at
		FROM object_type
		WHERE
			deleted_at IS NULL
	`
	primaryKeyColumn := sortRegexp.ReplaceAllString(PrimarySortKey, `_$1`)
	sortBy := sortRegexp.ReplaceAllString(listParams.SortBy, `_$1`)

	if listParams.Query != nil {
		searchTermReplacement := fmt.Sprintf("%%%s%%", *listParams.Query)
		query = fmt.Sprintf("%s AND %s LIKE ?", query, "type_id")
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
			if listParams.SortBy == PrimarySortKey {
				query = fmt.Sprintf("%s AND %s %s ?", query, primaryKeyColumn, comparisonOp)
				replacements = append(replacements, listParams.NextCursor.ID())
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, sortBy, primaryKeyColumn, comparisonOp, sortBy)
				replacements = append(replacements,
					listParams.NextCursor.ID(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, primaryKeyColumn, comparisonOp, sortBy)
				replacements = append(replacements,
					listParams.NextCursor.ID(),
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, sortBy, comparisonOp, primaryKeyColumn, comparisonOp, sortBy)
				replacements = append(replacements,
					listParams.NextCursor.Value(),
					listParams.NextCursor.ID(),
					listParams.NextCursor.Value(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, sortBy, comparisonOp, sortBy, primaryKeyColumn, comparisonOp, sortBy)
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
			if listParams.SortBy == PrimarySortKey {
				query = fmt.Sprintf("%s AND %s %s ?", query, primaryKeyColumn, comparisonOp)
				replacements = append(replacements, listParams.PrevCursor.ID())
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, primaryKeyColumn, comparisonOp, sortBy)
				replacements = append(replacements,
					listParams.PrevCursor.ID(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, sortBy, primaryKeyColumn, comparisonOp, sortBy)
				replacements = append(replacements,
					listParams.PrevCursor.ID(),
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, sortBy, comparisonOp, sortBy, primaryKeyColumn, comparisonOp, sortBy)
				replacements = append(replacements,
					listParams.PrevCursor.Value(),
					listParams.PrevCursor.ID(),
					listParams.PrevCursor.Value(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, sortBy, comparisonOp, primaryKeyColumn, comparisonOp, sortBy)
				replacements = append(replacements,
					listParams.PrevCursor.Value(),
					listParams.PrevCursor.ID(),
					listParams.PrevCursor.Value(),
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

	if listParams.PrevCursor != nil {
		if listParams.SortBy != PrimarySortKey {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortBy, service.SortOrderDesc, invertedNullSortClause, primaryKeyColumn, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit+1)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortBy, service.SortOrderAsc, invertedNullSortClause, primaryKeyColumn, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit+1)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s %s, %s %s", query, sortBy, listParams.SortOrder, nullSortClause, primaryKeyColumn, listParams.SortOrder)
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, sortBy, service.SortOrderDesc, invertedNullSortClause)
				replacements = append(replacements, listParams.Limit+1)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, sortBy, service.SortOrderAsc, invertedNullSortClause)
				replacements = append(replacements, listParams.Limit+1)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s %s", query, sortBy, listParams.SortOrder, nullSortClause)
		}
	} else {
		if listParams.SortBy != PrimarySortKey {
			query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortBy, listParams.SortOrder, nullSortClause, primaryKeyColumn, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit+1)
		} else {
			query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, primaryKeyColumn, listParams.SortOrder, nullSortClause)
			replacements = append(replacements, listParams.Limit+1)
		}
	}

	err := repo.DB.SelectContext(
		ctx,
		&objectTypes,
		query,
		replacements...,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models, nil, nil, nil
		}
		return nil, nil, nil, errors.Wrap(err, "error listing object types")
	}

	if len(objectTypes) == 0 {
		return models, nil, nil, nil
	}

	i := 0
	if listParams.PrevCursor != nil && len(objectTypes) > listParams.Limit {
		i = 1
	}
	for i < len(objectTypes) && len(models) < listParams.Limit {
		models = append(models, &objectTypes[i])
		i++
	}

	//nolint:gosec
	firstElem := models[0]
	lastElem := models[len(models)-1]
	var firstValue interface{} = nil
	var lastValue interface{} = nil
	switch listParams.SortBy {
	case PrimarySortKey:
		// do nothing
	case "createdAt":
		firstValue = firstElem.GetCreatedAt()
		lastValue = lastElem.GetCreatedAt()
	}

	prevCursor := service.NewCursor(firstElem.GetTypeId(), firstValue)
	nextCursor := service.NewCursor(lastElem.GetTypeId(), lastValue)
	if len(objectTypes) <= listParams.Limit {
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

func (repo PostgresRepository) UpdateByTypeId(ctx context.Context, typeId string, model Model) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE object_type
			SET
				definition = ?,
				updated_at = CURRENT_TIMESTAMP(6)
			WHERE
				type_id = ? AND
				deleted_at IS NULL
		`,
		model.GetDefinition(),
		typeId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.NewRecordNotFoundError("ObjectType", typeId)
		}
		return errors.Wrapf(err, "error updating object type %s", typeId)
	}

	return nil
}

func (repo PostgresRepository) DeleteByTypeId(ctx context.Context, typeId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE object_type
			SET
				updated_at = CURRENT_TIMESTAMP(6),
				deleted_at = CURRENT_TIMESTAMP(6)
			WHERE
				type_id = ? AND
				deleted_at IS NULL
		`,
		typeId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.NewRecordNotFoundError("ObjectType", typeId)
		}
		return errors.Wrapf(err, "error deleting object type %s", typeId)
	}

	return nil
}
