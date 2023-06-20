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
		database.NewSQLRepository(db),
	}
}

func (repo PostgresRepository) Create(ctx context.Context, model Model) (int64, error) {
	var newUserId int64
	err := repo.DB(ctx).GetContext(
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
	err := repo.DB(ctx).GetContext(
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
	err := repo.DB(ctx).GetContext(
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
	defaultSort := regexp.MustCompile("([A-Z])").ReplaceAllString(DefaultSortBy, `_$1`)
	sortBy := regexp.MustCompile("([A-Z])").ReplaceAllString(listParams.SortBy, `_$1`)

	if listParams.Query != nil {
		searchTermReplacement := fmt.Sprintf("%%%s%%", *listParams.Query)
		query = fmt.Sprintf("%s AND (%s LIKE ? OR email LIKE ?)", query, defaultSort)
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
				query = fmt.Sprintf("%s AND %s %s ?", query, defaultSort, comparisonOp)
				replacements = append(replacements, listParams.AfterId)
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, sortBy, defaultSort, comparisonOp, sortBy)
				replacements = append(replacements,
					listParams.AfterId,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, defaultSort, comparisonOp, sortBy)
				replacements = append(replacements,
					listParams.AfterId,
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, sortBy, comparisonOp, defaultSort, comparisonOp, sortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, sortBy, comparisonOp, sortBy, defaultSort, comparisonOp, sortBy)
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
				query = fmt.Sprintf("%s AND %s %s ?", query, defaultSort, comparisonOp)
				replacements = append(replacements, listParams.BeforeId)
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, defaultSort, comparisonOp, sortBy)
				replacements = append(replacements,
					listParams.BeforeId,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, sortBy, defaultSort, comparisonOp, sortBy)
				replacements = append(replacements,
					listParams.BeforeId,
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, sortBy, comparisonOp, sortBy, defaultSort, comparisonOp, sortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, sortBy, comparisonOp, defaultSort, comparisonOp, sortBy)
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
		if listParams.SortBy != DefaultSortBy {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortBy, service.SortOrderDesc, invertedNullSortClause, defaultSort, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortBy, service.SortOrderAsc, invertedNullSortClause, defaultSort, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s %s, %s %s", query, sortBy, listParams.SortOrder, nullSortClause, defaultSort, listParams.SortOrder)
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, sortBy, service.SortOrderDesc, invertedNullSortClause)
				replacements = append(replacements, listParams.Limit)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, sortBy, service.SortOrderAsc, invertedNullSortClause)
				replacements = append(replacements, listParams.Limit)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s %s", query, sortBy, listParams.SortOrder, nullSortClause)
		}
	} else {
		if listParams.SortBy != DefaultSortBy {
			query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortBy, listParams.SortOrder, nullSortClause, defaultSort, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, defaultSort, listParams.SortOrder, nullSortClause)
			replacements = append(replacements, listParams.Limit)
		}
	}

	err := repo.DB(ctx).SelectContext(
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
	_, err := repo.DB(ctx).ExecContext(
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
	_, err := repo.DB(ctx).ExecContext(
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
