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
	"database/sql"
	"fmt"
	"time"

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
	var newObjectTypeId int64
	now := time.Now().UTC()
	err := repo.DB.GetContext(
		database.CtxWithWriterOverride(ctx),
		&newObjectTypeId,
		`
			INSERT INTO objectType (
				typeId,
				definition,
				createdAt,
				updatedAt
			) VALUES (?, ?, ?, ?)
			ON CONFLICT (typeId) DO UPDATE SET
				definition = ?,
				createdAt = IIF(objectType.deletedAt IS NULL, objectType.createdAt, ?),
				updatedAt = ?,
				deletedAt = NULL
			RETURNING id
		`,
		model.GetTypeId(),
		model.GetDefinition(),
		now,
		now,
		model.GetDefinition(),
		now,
		now,
	)
	if err != nil {
		return -1, errors.Wrap(err, "error creating object type")
	}

	return newObjectTypeId, nil
}

func (repo SQLiteRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var objectType ObjectType
	err := repo.DB.GetContext(
		ctx,
		&objectType,
		`
			SELECT id, typeId, definition, createdAt, updatedAt, deletedAt
			FROM objectType
			WHERE
				id = ? AND
				deletedAt IS NULL
		`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &objectType, service.NewRecordNotFoundError("ObjectType", id)
		}
		return &objectType, errors.Wrapf(err, "error getting object type %d", id)
	}

	return &objectType, nil
}

func (repo SQLiteRepository) GetByTypeId(ctx context.Context, typeId string) (Model, error) {
	var objectType ObjectType
	err := repo.DB.GetContext(
		ctx,
		&objectType,
		`
			SELECT id, typeId, definition, createdAt, updatedAt, deletedAt
			FROM objectType
			WHERE
				typeId = ? AND
				deletedAt IS NULL
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

func (repo SQLiteRepository) List(ctx context.Context, listParams service.ListParams) ([]Model, *service.Cursor, *service.Cursor, error) {
	models := make([]Model, 0)
	objectTypes := make([]ObjectType, 0)
	replacements := make([]interface{}, 0)
	query := `
		SELECT id, typeId, definition, createdAt, updatedAt, deletedAt
		FROM objectType
		WHERE
			deletedAt IS NULL
	`

	if listParams.Query != nil {
		searchTermReplacement := fmt.Sprintf("%%%s%%", *listParams.Query)
		query = fmt.Sprintf("%s AND %s LIKE ?", query, "typeId")
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
				query = fmt.Sprintf("%s AND %s %s ?", query, PrimarySortKey, comparisonOp)
				replacements = append(replacements, listParams.NextCursor.ID())
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, listParams.SortBy, PrimarySortKey, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.NextCursor.ID(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, PrimarySortKey, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.NextCursor.ID(),
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, PrimarySortKey, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.NextCursor.Value(),
					listParams.NextCursor.ID(),
					listParams.NextCursor.Value(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, listParams.SortBy, PrimarySortKey, comparisonOp, listParams.SortBy)
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
				query = fmt.Sprintf("%s AND %s %s ?", query, PrimarySortKey, comparisonOp)
				replacements = append(replacements, listParams.PrevCursor.ID())
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, PrimarySortKey, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.PrevCursor.ID(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, listParams.SortBy, PrimarySortKey, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.PrevCursor.ID(),
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, listParams.SortBy, PrimarySortKey, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.PrevCursor.Value(),
					listParams.PrevCursor.ID(),
					listParams.PrevCursor.Value(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, PrimarySortKey, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.PrevCursor.Value(),
					listParams.PrevCursor.ID(),
					listParams.PrevCursor.Value(),
				)
			}
		}
	}

	if listParams.PrevCursor != nil {
		if listParams.SortBy != PrimarySortKey {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderDesc, PrimarySortKey, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit+1)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderAsc, PrimarySortKey, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit+1)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s, %s %s", query, listParams.SortBy, listParams.SortOrder, PrimarySortKey, listParams.SortOrder)
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit+1)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit+1)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s", query, listParams.SortBy, listParams.SortOrder)
		}
	} else {
		if listParams.SortBy != PrimarySortKey {
			query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, PrimarySortKey, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit+1)
		} else {
			query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, PrimarySortKey, listParams.SortOrder)
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

func (repo SQLiteRepository) UpdateByTypeId(ctx context.Context, typeId string, model Model) error {
	now := time.Now().UTC()
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE objectType
			SET
				definition = ?,
				updatedAt = ?
			WHERE
				typeId = ? AND
				deletedAt IS NULL
		`,
		model.GetDefinition(),
		now,
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

func (repo SQLiteRepository) DeleteByTypeId(ctx context.Context, typeId string) error {
	now := time.Now().UTC()
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE objectType
			SET
				updatedAt = ?,
				deletedAt = ?
			WHERE
				typeId = ? AND
				deletedAt IS NULL
		`,
		now,
		now,
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
