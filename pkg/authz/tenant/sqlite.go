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

package tenant

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
	var newTenantId int64
	now := time.Now().UTC()
	err := repo.DB.GetContext(
		database.CtxWithWriterOverride(ctx),
		&newTenantId,
		`
			INSERT INTO tenant (
				tenantId,
				objectId,
				name,
				createdAt,
				updatedAt
			) VALUES (?, ?, ?, ?, ?)
			ON CONFLICT (tenantId) DO UPDATE SET
				objectId = ?,
				name = ?,
				createdAt = ?,
				deletedAt = NULL
			RETURNING id
		`,
		model.GetTenantId(),
		model.GetObjectId(),
		model.GetName(),
		now,
		now,
		model.GetObjectId(),
		model.GetName(),
		now,
	)

	if err != nil {
		return -1, errors.Wrap(err, "error creating tenant")
	}

	return newTenantId, nil
}

func (repo SQLiteRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var tenant Tenant
	err := repo.DB.GetContext(
		ctx,
		&tenant,
		`
			SELECT id, objectId, tenantId, name, createdAt, updatedAt, deletedAt
			FROM tenant
			WHERE
				id = ? AND
				deletedAt IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Tenant", id)
		default:
			return nil, errors.Wrapf(err, "error getting tenant %d", id)
		}
	}

	return &tenant, nil
}

func (repo SQLiteRepository) GetByTenantId(ctx context.Context, tenantId string) (Model, error) {
	var tenant Tenant
	err := repo.DB.GetContext(
		ctx,
		&tenant,
		`
			SELECT id, objectId, tenantId, name, createdAt, updatedAt, deletedAt
			FROM tenant
			WHERE
				tenantId = ? AND
				deletedAt IS NULL
		`,
		tenantId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Tenant", tenantId)
		default:
			return nil, errors.Wrapf(err, "error getting tenant %s", tenantId)
		}
	}

	return &tenant, nil
}

func (repo SQLiteRepository) List(ctx context.Context, listParams service.ListParams) ([]Model, error) {
	models := make([]Model, 0)
	tenants := make([]Tenant, 0)
	query := `
		SELECT id, objectId, tenantId, name, createdAt, updatedAt, deletedAt
		FROM tenant
		WHERE
			deletedAt IS NULL

	`
	replacements := []interface{}{}

	if listParams.Query != nil {
		searchTermReplacement := fmt.Sprintf("%%%s%%", *listParams.Query)
		query = fmt.Sprintf("%s AND (%s LIKE ? OR name LIKE ?)", query, DefaultSortBy)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.AfterId != nil {
		comparisonOp := "<"
		if listParams.SortOrder == service.SortOrderAsc {
			comparisonOp = ">"
		}

		switch listParams.AfterValue {
		case nil:
			if listParams.SortBy == DefaultSortBy {
				query = fmt.Sprintf("%s AND %s %s ?", query, DefaultSortBy, comparisonOp)
				replacements = append(replacements, listParams.AfterId)
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, listParams.SortBy, DefaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterId,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, DefaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterId,
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, DefaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, listParams.SortBy, DefaultSortBy, comparisonOp, listParams.SortBy)
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
			if listParams.SortBy == DefaultSortBy {
				query = fmt.Sprintf("%s AND %s %s ?", query, DefaultSortBy, comparisonOp)
				replacements = append(replacements, listParams.BeforeId)
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, DefaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeId,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, listParams.SortBy, DefaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeId,
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, listParams.SortBy, DefaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, DefaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		}
	}

	if listParams.BeforeId != nil {
		if listParams.SortBy != DefaultSortBy {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderDesc, DefaultSortBy, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderAsc, DefaultSortBy, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s, %s %s", query, listParams.SortBy, listParams.SortOrder, DefaultSortBy, listParams.SortOrder)
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s", query, listParams.SortBy, listParams.SortOrder)
		}
	} else {
		if listParams.SortBy != DefaultSortBy {
			query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, DefaultSortBy, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, DefaultSortBy, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		}
	}

	err := repo.DB.SelectContext(
		ctx,
		&tenants,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return models, errors.Wrap(err, "error listing tenants")
		}
	}

	for i := range tenants {
		models = append(models, &tenants[i])
	}

	return models, nil
}

func (repo SQLiteRepository) UpdateByTenantId(ctx context.Context, tenantId string, model Model) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE tenant
			SET
				name = ?,
				updatedAt = ?
			WHERE
				tenantId = ? AND
				deletedAt IS NULL
		`,
		model.GetName(),
		time.Now().UTC(),
		tenantId,
	)
	if err != nil {
		return errors.Wrapf(err, "error updating tenant %s", tenantId)
	}

	return nil
}

func (repo SQLiteRepository) DeleteByTenantId(ctx context.Context, tenantId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE tenant
			SET
				deletedAt = ?
			WHERE
				tenantId = ? AND
				deletedAt IS NULL
		`,
		time.Now().UTC(),
		tenantId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("Tenant", tenantId)
		default:
			return errors.Wrapf(err, "error deleting tenant %s", tenantId)
		}
	}

	return nil
}
