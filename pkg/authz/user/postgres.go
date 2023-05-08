package authz

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
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

func (repo PostgresRepository) Create(ctx context.Context, model Model) (int64, error) {
	var newUserId int64
	err := repo.DB.GetContext(
		ctx,
		&newUserId,
		`
			INSERT INTO "user" (
				user_id,
				object_id,
				email
			) VALUES (?, ?, ?)
			ON CONFLICT (user_id) DO UPDATE SET
				object_id = ?,
				email = ?,
				created_at = CURRENT_TIMESTAMP(6),
				deleted_at = NULL
			RETURNING id
		`,
		model.GetUserId(),
		model.GetObjectId(),
		model.GetEmail(),
		model.GetObjectId(),
		model.GetEmail(),
	)
	if err != nil {
		return -1, errors.Wrap(err, "error creating user")
	}

	return newUserId, nil
}

func (repo PostgresRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var user User
	err := repo.DB.GetContext(
		ctx,
		&user,
		`
			SELECT id, object_id, user_id, email, created_at, updated_at, deleted_at
			FROM "user"
			WHERE
				id = ? AND
				deleted_at IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("User", id)
		default:
			return nil, errors.Wrapf(err, "error getting user %d", id)
		}
	}

	return &user, nil
}

func (repo PostgresRepository) GetByUserId(ctx context.Context, userId string) (Model, error) {
	var user User
	err := repo.DB.GetContext(
		ctx,
		&user,
		`
			SELECT id, object_id, user_id, email, created_at, updated_at, deleted_at
			FROM "user"
			WHERE
				user_id = ? AND
				deleted_at IS NULL
		`,
		userId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("User", userId)
		default:
			return nil, errors.Wrapf(err, "error getting user %s", userId)
		}
	}

	return &user, nil
}

func (repo PostgresRepository) List(ctx context.Context, listParams service.ListParams) ([]Model, error) {
	models := make([]Model, 0)
	users := make([]User, 0)
	query := `
		SELECT id, object_id, user_id, email, created_at, updated_at, deleted_at
		FROM "user"
		WHERE
			deleted_at IS NULL
	`
	replacements := []interface{}{}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (user_id LIKE ? OR email LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	sortBy := regexp.MustCompile("([A-Z])").ReplaceAllString(listParams.SortBy, `_$1`)
	if listParams.AfterId != "" {
		if listParams.AfterValue != nil {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s > ? OR (user_id > ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s < ? OR (user_id < ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND user_id > ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND user_id < ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.BeforeId != "" {
		if listParams.BeforeValue != nil {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s < ? OR (user_id < ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s > ? OR (user_id > ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND user_id < ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND user_id > ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	nullSortClause := "NULLS LAST"
	if listParams.SortOrder == service.SortOrderAsc {
		nullSortClause = "NULLS FIRST"
	}

	if listParams.SortBy != "userId" {
		query = fmt.Sprintf("%s ORDER BY %s %s %s, user_id %s LIMIT ?", query, sortBy, listParams.SortOrder, nullSortClause, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	} else {
		query = fmt.Sprintf("%s ORDER BY user_id %s %s LIMIT ?", query, listParams.SortOrder, nullSortClause)
		replacements = append(replacements, listParams.Limit)
	}

	err := repo.DB.SelectContext(
		ctx,
		&users,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return nil, errors.Wrap(err, "error listing users")
		}
	}

	for i := range users {
		models = append(models, &users[i])
	}

	return models, nil
}

func (repo PostgresRepository) UpdateByUserId(ctx context.Context, userId string, model Model) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE "user"
			SET
				email = ?
			WHERE
				user_id = ? AND
				deleted_at IS NULL
		`,
		model.GetEmail(),
		model.GetUserId(),
	)
	if err != nil {
		return errors.Wrapf(err, "error updating user %s", userId)
	}

	return nil
}

func (repo PostgresRepository) DeleteByUserId(ctx context.Context, userId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE "user"
			SET deleted_at = ?
			WHERE
				user_id = ? AND
				deleted_at IS NULL
		`,
		time.Now().UTC(),
		userId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("User", userId)
		default:
			return errors.Wrapf(err, "error deleting user %s", userId)
		}
	}

	return nil
}
