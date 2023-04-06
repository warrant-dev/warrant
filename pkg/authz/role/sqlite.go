package authz

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

type SQLiteRepository struct {
	database.SQLRepository
}

func NewSQLiteRepository(db *database.SQLite) SQLiteRepository {
	return SQLiteRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo SQLiteRepository) Create(ctx context.Context, role Role) (int64, error) {
	var newRoleId int64
	now := time.Now().UTC()
	err := repo.DB.GetContext(
		ctx,
		&newRoleId,
		`
			INSERT INTO role (
				objectId,
				roleId,
				name,
				description,
				createdAt,
				updatedAt
			) VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT (roleId) DO UPDATE SET
				objectId = ?,
				roleId = ?,
				name = ?,
				description = ?,
				createdAt = ?,
				deletedAt = NULL
			RETURNING id
		`,
		role.ObjectId,
		role.RoleId,
		role.Name,
		role.Description,
		now,
		now,
		role.ObjectId,
		role.RoleId,
		role.Name,
		role.Description,
		now,
	)

	if err != nil {
		return 0, errors.Wrap(err, "Unable to create role")
	}

	return newRoleId, err
}

func (repo SQLiteRepository) GetById(ctx context.Context, id int64) (*Role, error) {
	var role Role
	err := repo.DB.GetContext(
		ctx,
		&role,
		`
			SELECT id, objectId, roleId, name, description, createdAt, updatedAt, deletedAt
			FROM role
			WHERE
				id = ? AND
				deletedAt IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Role", id)
		default:
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get role id %d from sqlite", id))
		}
	}

	return &role, nil
}

func (repo SQLiteRepository) GetByRoleId(ctx context.Context, roleId string) (*Role, error) {
	var role Role
	err := repo.DB.GetContext(
		ctx,
		&role,
		`
			SELECT id, objectId, roleId, name, description, createdAt, updatedAt, deletedAt
			FROM role
			WHERE
				roleId = ? AND
				deletedAt IS NULL
		`,
		roleId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Role", roleId)
		default:
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get role %s from sqlite", roleId))
		}
	}

	return &role, nil
}

func (repo SQLiteRepository) List(ctx context.Context, listParams middleware.ListParams) ([]Role, error) {
	roles := make([]Role, 0)
	query := `
		SELECT id, objectId, roleId, name, description, createdAt, updatedAt, deletedAt
		FROM role
		WHERE
			deletedAt IS NULL
	`
	replacements := []interface{}{}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (roleId LIKE ? OR name LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.AfterId != "" {
		if listParams.AfterValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s > ? OR (roleId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s < ? OR (roleId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND roleId > ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND roleId < ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.BeforeId != "" {
		if listParams.BeforeValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s < ? OR (roleId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s > ? OR (roleId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND roleId < ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND roleId > ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.SortBy != "roleId" {
		query = fmt.Sprintf("%s ORDER BY %s %s, roleId %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	} else {
		query = fmt.Sprintf("%s ORDER BY roleId %s LIMIT ?", query, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	}

	err := repo.DB.SelectContext(
		ctx,
		&roles,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return roles, nil
		default:
			return roles, service.NewInternalError("Unable to list roles")
		}
	}

	return roles, nil
}

func (repo SQLiteRepository) UpdateByRoleId(ctx context.Context, roleId string, role Role) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE role
			SET
				name = ?,
				description = ?,
				updatedAt = ?
			WHERE
				roleId = ? AND
				deletedAt IS NULL
		`,
		role.Name,
		role.Description,
		time.Now().UTC(),
		roleId,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error updating role %s", roleId))
	}

	return nil
}

func (repo SQLiteRepository) DeleteByRoleId(ctx context.Context, roleId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE role
			SET
				deletedAt = ?
			WHERE
				roleId = ? AND
				deletedAt IS NULL
		`,
		time.Now().UTC(),
		roleId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("Role", roleId)
		default:
			return err
		}
	}

	return nil
}
