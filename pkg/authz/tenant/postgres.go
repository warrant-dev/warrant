package tenant

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

type PostgresRepository struct {
	database.SQLRepository
}

func NewPostgresRepository(db *database.Postgres) PostgresRepository {
	return PostgresRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo PostgresRepository) Create(ctx context.Context, tenant TenantModel) (int64, error) {
	var newTenantId int64
	err := repo.DB.GetContext(
		ctx,
		&newTenantId,
		`
			INSERT INTO tenant (
				tenant_id,
				object_id,
				name
			) VALUES (?, ?, ?)
			ON CONFLICT (tenant_id) DO UPDATE SET
				object_id = ?,
				name = ?,
				created_at = CURRENT_TIMESTAMP(6),
				deleted_at = NULL
			RETURNING id
		`,
		tenant.GetTenantId(),
		tenant.GetObjectId(),
		tenant.GetName(),
		tenant.GetObjectId(),
		tenant.GetName(),
	)

	if err != nil {
		return 0, errors.Wrap(err, "Unable to create Tenant")
	}

	return newTenantId, nil
}

func (repo PostgresRepository) GetById(ctx context.Context, id int64) (TenantModel, error) {
	var tenant Tenant
	err := repo.DB.GetContext(
		ctx,
		&tenant,
		`
			SELECT id, object_id, tenant_id, name, created_at, updated_at, deleted_at
			FROM tenant
			WHERE
				id = ? AND
				deleted_at IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Tenant", id)
		default:
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get Tenant %d from mysql", id))
		}
	}

	return &tenant, nil
}

func (repo PostgresRepository) GetByTenantId(ctx context.Context, tenantId string) (TenantModel, error) {
	var tenant Tenant
	err := repo.DB.GetContext(
		ctx,
		&tenant,
		`
			SELECT id, object_id, tenant_id, name, created_at, updated_at, deleted_at
			FROM tenant
			WHERE
				tenant_id = ? AND
				deleted_at IS NULL
		`,
		tenantId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Tenant", tenantId)
		default:
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get Tenant %s from mysql", tenantId))
		}
	}

	return &tenant, nil
}

func (repo PostgresRepository) List(ctx context.Context, listParams middleware.ListParams) ([]TenantModel, error) {
	tenants := make([]TenantModel, 0)
	query := `
		SELECT id, object_id, tenant_id, name, created_at, updated_at, deleted_at
		FROM tenant
		WHERE
			deleted_at IS NULL

	`
	replacements := []interface{}{}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (tenant_id LIKE ? OR name LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	sortBy := regexp.MustCompile("([A-Z])").ReplaceAllString(listParams.SortBy, `_$1`)
	if listParams.AfterId != "" {
		if listParams.AfterValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s > ? OR (tenant_id > ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s < ? OR (tenant_id < ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND tenant_id > ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND tenant_id < ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.BeforeId != "" {
		if listParams.BeforeValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s < ? OR (tenant_id < ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s > ? OR (tenant_id > ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND tenant_id < ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND tenant_id > ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	nullSortClause := "NULLS LAST"
	if listParams.SortOrder == middleware.SortOrderAsc {
		nullSortClause = "NULLS FIRST"
	}

	if listParams.SortBy != "tenantId" {
		query = fmt.Sprintf("%s ORDER BY %s %s %s, tenant_id %s LIMIT ?", query, sortBy, listParams.SortOrder, nullSortClause, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	} else {
		query = fmt.Sprintf("%s ORDER BY tenant_id %s %s LIMIT ?", query, listParams.SortOrder, nullSortClause)
		replacements = append(replacements, listParams.Limit)
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
			return tenants, nil
		default:
			return tenants, service.NewInternalError("Unable to list tenants")
		}
	}

	return tenants, nil
}

func (repo PostgresRepository) UpdateByTenantId(ctx context.Context, tenantId string, tenant TenantModel) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE tenant
			SET
				name = ?
			WHERE
				tenant_id = ? AND
				deleted_at IS NULL
		`,
		tenant.GetName(),
		tenantId,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error updating tenant %d", tenant.GetID()))
	}

	return nil
}

func (repo PostgresRepository) DeleteByTenantId(ctx context.Context, tenantId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE tenant
			SET
				deleted_at = ?
			WHERE
				tenant_id = ? AND
				deleted_at IS NULL
		`,
		time.Now().UTC(),
		tenantId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("Tenant", tenantId)
		default:
			return err
		}
	}

	return nil
}
