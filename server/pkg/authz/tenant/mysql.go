package tenant

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	objecttype "github.com/warrant-dev/warrant/server/pkg/authz/objecttype"
	warrant "github.com/warrant-dev/warrant/server/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/server/pkg/database"
	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

type MySQLRepository struct {
	database.SQLRepository
}

func NewMySQLRepository(db *database.MySQL) MySQLRepository {
	return MySQLRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo MySQLRepository) Create(ctx context.Context, tenant Tenant) (int64, error) {
	result, err := repo.DB.ExecContext(
		ctx,
		`
			INSERT INTO tenant (
				tenantId,
				objectId,
				name
			) VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE
				objectId = ?,
				name = ?,
				createdAt = NOW(),
				deletedAt = NULL
		`,
		tenant.TenantId,
		tenant.ObjectId,
		tenant.Name,
		tenant.ObjectId,
		tenant.Name,
	)

	if err != nil {
		mysqlErr, ok := err.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1062 {
			return 0, service.NewDuplicateRecordError("Tenant", tenant.TenantId, "Tenant with given tenantId already exists")
		}

		return 0, service.NewInternalError("Unable to create Tenant")
	}

	newTenantId, err := result.LastInsertId()
	if err != nil {
		return 0, service.NewInternalError("Unable to create Tenant")
	}

	return newTenantId, nil
}

func (repo MySQLRepository) BatchCreate(ctx context.Context, tenants []Tenant) error {
	result, err := repo.DB.NamedExecContext(
		ctx,
		`
			INSERT INTO tenant (
				tenantId,
				objectId,
				name
			) VALUES (
				:tenantId,
				:objectId,
				:name
			) ON DUPLICATE KEY UPDATE
				objectId = VALUES(objectId),
				name = VALUES(name),
				createdAt = NOW(),
				deletedAt = NULL
		`,
		tenants,
	)
	if err != nil {
		return service.NewInternalError("Unable to create tenants")
	}

	_, err = result.RowsAffected()
	if err != nil {
		return service.NewInternalError("Unable to create tenants")
	}

	return nil
}

func (repo MySQLRepository) UpdateByTenantId(ctx context.Context, tenantId string, tenant Tenant) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE tenant
			SET
				name = ?
			WHERE
				tenantId = ? AND
				deletedAt IS NULL
		`,
		tenant.Name,
		tenantId,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error updating tenant %d", tenant.ID))
	}

	return nil
}

func (repo MySQLRepository) DeleteByTenantId(ctx context.Context, tenantId string) error {
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
		time.Now(),
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

func (repo MySQLRepository) GetByTenantId(ctx context.Context, tenantId string) (*Tenant, error) {
	var tenant Tenant
	err := repo.DB.GetContext(
		ctx,
		&tenant,
		`
			SELECT id, objectId, tenantId, name, createdAt, updatedAt
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
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get Tenant %s from mysql", tenantId))
		}
	}

	return &tenant, nil
}

func (repo MySQLRepository) BatchGet(ctx context.Context, tenantIds []string) ([]Tenant, error) {
	tenants := make([]Tenant, 0)

	query, args, err := sqlx.In(
		`
			SELECT id, objectId, tenantId, name, createdAt, updatedAt
			FROM tenant
			WHERE
				tenantId IN (?) AND
				deletedAt IS NULL
		`,
		tenantIds,
	)
	if err != nil {
		return tenants, err
	}

	err = repo.DB.SelectContext(
		ctx,
		&tenants,
		query,
		args...,
	)
	if err != nil {
		return tenants, err
	}

	return tenants, nil
}

func (repo MySQLRepository) GetById(ctx context.Context, id int64) (*Tenant, error) {
	var tenant Tenant
	err := repo.DB.GetContext(
		ctx,
		&tenant,
		`
			SELECT id, objectId, tenantId, name, createdAt, updatedAt
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
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get Tenant %d from mysql", id))
		}
	}

	return &tenant, nil
}

func (repo MySQLRepository) List(ctx context.Context, listParams middleware.ListParams) ([]Tenant, error) {
	tenants := make([]Tenant, 0)
	query := `
		SELECT id, objectId, tenantId, name, createdAt, updatedAt
		FROM tenant
		WHERE
			deletedAt IS NULL

	`
	replacements := []interface{}{}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (tenantId LIKE ? OR name LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.UseCursorPagination() {
		if listParams.AfterId != "" {
			if listParams.AfterValue != nil {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND (%s > ? OR (tenantId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.AfterValue,
						listParams.AfterId,
						listParams.AfterValue,
					)
				} else {
					query = fmt.Sprintf("%s AND (%s < ? OR (tenantId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.AfterValue,
						listParams.AfterId,
						listParams.AfterValue,
					)
				}
			} else {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND tenantId > ?", query)
					replacements = append(replacements, listParams.AfterId)
				} else {
					query = fmt.Sprintf("%s AND tenantId < ?", query)
					replacements = append(replacements, listParams.AfterId)
				}
			}
		}

		if listParams.BeforeId != "" {
			if listParams.BeforeValue != nil {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND (%s < ? OR (tenantId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.BeforeValue,
						listParams.BeforeId,
						listParams.BeforeValue,
					)
				} else {
					query = fmt.Sprintf("%s AND (%s > ? OR (tenantId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.BeforeValue,
						listParams.BeforeId,
						listParams.BeforeValue,
					)
				}
			} else {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND tenantId < ?", query)
					replacements = append(replacements, listParams.AfterId)
				} else {
					query = fmt.Sprintf("%s AND tenantId > ?", query)
					replacements = append(replacements, listParams.AfterId)
				}
			}
		}

		if listParams.SortBy != "" && listParams.SortBy != "tenantId" {
			query = fmt.Sprintf("%s ORDER BY %s %s, tenantId %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY tenantId %s LIMIT ?", query, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		}
	} else {
		offset := (listParams.Page - 1) * listParams.Limit

		if listParams.SortBy != "" && listParams.SortBy != "tenantId" {
			query = fmt.Sprintf("%s ORDER BY %s %s, tenantId %s LIMIT ?, ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
			replacements = append(replacements, offset, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY tenantId %s LIMIT ?, ?", query, listParams.SortOrder)
			replacements = append(replacements, offset, listParams.Limit)
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
			return tenants, nil
		default:
			return tenants, service.NewInternalError("Unable to list tenants")
		}
	}

	return tenants, nil
}

func (repo MySQLRepository) ListByUserId(ctx context.Context, userId string, listParams middleware.ListParams) ([]UserTenant, error) {
	tenants := make([]UserTenant, 0)
	query := `
		SELECT tenant.id, tenant.tenantId, tenant.name, warrant.relation, tenant.objectId, tenant.createdAt, tenant.updatedAt
		FROM tenant
		INNER JOIN warrant
			ON tenant.tenantId = warrant.objectId
		WHERE
			tenant.deletedAt IS NULL AND
			warrant.objectType = ? AND
			warrant.relation IN (?, ?, ?) AND
			warrant.subject = ? AND
			warrant.deletedAt IS NULL
	`
	replacements := []interface{}{
		objecttype.ObjectTypeTenant,
		objecttype.RelationAdmin,
		objecttype.RelationManager,
		objecttype.RelationMember,
		warrant.UserIdToSubjectString(userId),
	}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (tenant.tenantId LIKE ? OR tenant.name LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.UseCursorPagination() {
		if listParams.AfterId != "" {
			if listParams.AfterValue != nil {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND (%s > ? OR (tenant.tenantId > ? AND tenant.%s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.AfterValue,
						listParams.AfterId,
						listParams.AfterValue,
					)
				} else {
					query = fmt.Sprintf("%s AND (%s < ? OR (tenant.tenantId < ? AND tenant.%s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.AfterValue,
						listParams.AfterId,
						listParams.AfterValue,
					)
				}
			} else {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND tenant.tenantId > ?", query)
					replacements = append(replacements, listParams.AfterId)
				} else {
					query = fmt.Sprintf("%s AND tenant.tenantId < ?", query)
					replacements = append(replacements, listParams.AfterId)
				}
			}
		}

		if listParams.BeforeId != "" {
			if listParams.BeforeValue != nil {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND (%s < ? OR (tenant.tenantId < ? AND tenant.%s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.BeforeValue,
						listParams.BeforeId,
						listParams.BeforeValue,
					)
				} else {
					query = fmt.Sprintf("%s AND (%s > ? OR (tenant.tenantId > ? AND tenant.%s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.BeforeValue,
						listParams.BeforeId,
						listParams.BeforeValue,
					)
				}
			} else {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND tenant.tenantId < ?", query)
					replacements = append(replacements, listParams.AfterId)
				} else {
					query = fmt.Sprintf("%s AND tenant.tenantId > ?", query)
					replacements = append(replacements, listParams.AfterId)
				}
			}
		}

		if listParams.SortBy != "" && listParams.SortBy != "tenantId" {
			query = fmt.Sprintf("%s ORDER BY tenant.%s %s, tenant.tenantId %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY tenant.tenantId %s LIMIT ?", query, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		}
	} else {
		offset := (listParams.Page - 1) * listParams.Limit

		if listParams.SortBy != "" && listParams.SortBy != "tenantId" {
			query = fmt.Sprintf("%s ORDER BY %s %s, tenant.tenantId %s LIMIT ?, ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
			replacements = append(replacements, offset, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY tenant.tenantId %s LIMIT ?, ?", query, listParams.SortOrder)
			replacements = append(replacements, offset, listParams.Limit)
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
			return tenants, nil
		default:
			return tenants, service.NewInternalError(fmt.Sprintf("Unable to list tenants for user %s", userId))
		}
	}

	return tenants, nil
}
