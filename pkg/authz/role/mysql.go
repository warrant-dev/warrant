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

type MySQLRepository struct {
	database.SQLRepository
}

func NewMySQLRepository(db *database.MySQL) MySQLRepository {
	return MySQLRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo MySQLRepository) Create(ctx context.Context, role Model) (int64, error) {
	result, err := repo.DB.ExecContext(
		ctx,
		`
			INSERT INTO role (
				objectId,
				roleId,
				name,
				description
			) VALUES (?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				objectId = ?,
				name = ?,
				description = ?,
				createdAt = CURRENT_TIMESTAMP(6),
				deletedAt = NULL
		`,
		role.GetObjectId(),
		role.GetRoleId(),
		role.GetName(),
		role.GetDescription(),
		role.GetObjectId(),
		role.GetName(),
		role.GetDescription(),
	)

	if err != nil {
		return -1, errors.Wrap(err, "error creating role")
	}

	newRoleId, err := result.LastInsertId()
	if err != nil {
		return -1, errors.Wrap(err, "error creating role")
	}

	return newRoleId, err
}

func (repo MySQLRepository) GetById(ctx context.Context, id int64) (Model, error) {
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
			return nil, errors.Wrapf(err, "error getting role %d", id)
		}
	}

	return &role, nil
}

func (repo MySQLRepository) GetByRoleId(ctx context.Context, roleId string) (Model, error) {
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
			return nil, errors.Wrapf(err, "error getting role %s", roleId)
		}
	}

	return &role, nil
}

func (repo MySQLRepository) List(ctx context.Context, listParams middleware.ListParams) ([]Model, error) {
	models := make([]Model, 0)
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
			return models, nil
		default:
			return models, errors.Wrap(err, "error listing roles")
		}
	}

	for i := range roles {
		models = append(models, &roles[i])
	}

	return models, nil
}

func (repo MySQLRepository) UpdateByRoleId(ctx context.Context, roleId string, model Model) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE role
			SET
				name = ?,
				description = ?
			WHERE
				roleId = ? AND
				deletedAt IS NULL
		`,
		model.GetName(),
		model.GetDescription(),
		roleId,
	)
	if err != nil {
		return errors.Wrapf(err, "error updating role %s", roleId)
	}

	return nil
}

func (repo MySQLRepository) DeleteByRoleId(ctx context.Context, roleId string) error {
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
			return errors.Wrapf(err, "error deleting role %s", roleId)
		}
	}

	return nil
}
