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
			INSERT INTO objectType (
				typeId,
				definition
			) VALUES (?, ?)
			ON DUPLICATE KEY UPDATE
				definition = ?,
				createdAt = IF(objectType.deletedAt IS NULL, objectType.createdAt, CURRENT_TIMESTAMP(6)),
				updatedAt = CURRENT_TIMESTAMP(6),
				deletedAt = NULL
		`,
		model.GetTypeId(),
		model.GetDefinition(),
		model.GetDefinition(),
	)
	if err != nil {
		return -1, errors.Wrap(err, "error creating object type")
	}

	newObjectTypeId, err := result.LastInsertId()
	if err != nil {
		return -1, errors.Wrap(err, "error creating object type")
	}

	return newObjectTypeId, nil
}

func (repo MySQLRepository) GetById(ctx context.Context, id int64) (Model, error) {
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

func (repo MySQLRepository) GetByTypeId(ctx context.Context, typeId string) (Model, error) {
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

func (repo MySQLRepository) ListAll(ctx context.Context) ([]Model, error) {
	models := make([]Model, 0)
	objectTypes := make([]ObjectType, 0)
	err := repo.DB.SelectContext(
		ctx,
		&objectTypes,
		`
			SELECT id, typeId, definition, createdAt, updatedAt, deletedAt
			FROM objectType
			WHERE
				deletedAt IS NULL
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

func (repo MySQLRepository) List(ctx context.Context, listParams service.ListParams) ([]Model, *service.Cursor, *service.Cursor, error) {
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
			if listParams.SortBy == primarySortKey {
				query = fmt.Sprintf("%s AND %s %s ?", query, primarySortKey, comparisonOp)
				replacements = append(replacements, listParams.NextCursor.ID())
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, listParams.SortBy, primarySortKey, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.NextCursor.ID(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, primarySortKey, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.NextCursor.ID(),
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, primarySortKey, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.NextCursor.Value(),
					listParams.NextCursor.ID(),
					listParams.NextCursor.Value(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, listParams.SortBy, primarySortKey, comparisonOp, listParams.SortBy)
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
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, primarySortKey, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.PrevCursor.ID(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, listParams.SortBy, primarySortKey, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.PrevCursor.ID(),
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, listParams.SortBy, primarySortKey, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.PrevCursor.Value(),
					listParams.PrevCursor.ID(),
					listParams.PrevCursor.Value(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, primarySortKey, comparisonOp, listParams.SortBy)
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
				query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderDesc, primarySortKey, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit+1)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderAsc, primarySortKey, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit+1)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s, %s %s", query, listParams.SortBy, listParams.SortOrder, primarySortKey, listParams.SortOrder)
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
		if listParams.SortBy != primarySortKey {
			query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, primarySortKey, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit+1)
		} else {
			query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, primarySortKey, listParams.SortOrder)
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
		return nil, nil, nil, errors.Wrap(err, "error listing warrants")
	}

	if len(objectTypes) == 0 {
		return models, nil, nil, nil
	}

	for i := 0; i < len(objectTypes) && i < listParams.Limit; i++ {
		models = append(models, &objectTypes[i])
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

func (repo MySQLRepository) UpdateByTypeId(ctx context.Context, typeId string, model Model) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE objectType
			SET
				definition = ?,
				updatedAt = CURRENT_TIMESTAMP(6)
			WHERE
				typeId = ? AND
				deletedAt IS NULL
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

func (repo MySQLRepository) DeleteByTypeId(ctx context.Context, typeId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE objectType
			SET
				updatedAt = CURRENT_TIMESTAMP(6),
				deletedAt = CURRENT_TIMESTAMP(6)
			WHERE
				typeId = ? AND
				deletedAt IS NULL
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
