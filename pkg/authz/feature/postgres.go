package authz

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

func (repo PostgresRepository) Create(ctx context.Context, model Model) (int64, error) {
	var newFeatureId int64
	err := repo.DB.GetContext(
		ctx,
		&newFeatureId,
		`
			INSERT INTO feature (
				object_id,
				feature_id,
				name,
				description
			) VALUES (?, ?, ?, ?)
			ON CONFLICT (feature_id) DO UPDATE SET
				object_id = ?,
				name = ?,
				description = ?,
				created_at = CURRENT_TIMESTAMP(6),
				deleted_at = NULL
			RETURNING id
		`,
		model.GetObjectId(),
		model.GetFeatureId(),
		model.GetName(),
		model.GetDescription(),
		model.GetObjectId(),
		model.GetName(),
		model.GetDescription(),
	)

	if err != nil {
		return 0, errors.Wrap(err, "Unable to create feature")
	}

	return newFeatureId, err
}

func (repo PostgresRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var feature Feature
	err := repo.DB.GetContext(
		ctx,
		&feature,
		`
			SELECT id, object_id, feature_id, name, description, created_at, updated_at, deleted_at
			FROM feature
			WHERE
				id = ? AND
				deleted_at IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Feature", id)
		default:
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get feature id %d from postgres", id))
		}
	}

	return &feature, nil
}

func (repo PostgresRepository) GetByFeatureId(ctx context.Context, featureId string) (Model, error) {
	var feature Feature
	err := repo.DB.GetContext(
		ctx,
		&feature,
		`
			SELECT id, object_id, feature_id, name, description, created_at, updated_at, deleted_at
			FROM feature
			WHERE
				feature_id = ? AND
				deleted_at IS NULL
		`,
		featureId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Feature", featureId)
		default:
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get feature %s from postgres", featureId))
		}
	}

	return &feature, nil
}

func (repo PostgresRepository) List(ctx context.Context, listParams middleware.ListParams) ([]Model, error) {
	models := make([]Model, 0)
	features := make([]Feature, 0)
	query := `
		SELECT id, object_id, feature_id, name, description, created_at, updated_at, deleted_at
		FROM feature
		WHERE
			deleted_at IS NULL
	`
	replacements := []interface{}{}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (feature_id LIKE ? OR name LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	sortBy := regexp.MustCompile("([A-Z])").ReplaceAllString(listParams.SortBy, `_$1`)
	if listParams.AfterId != "" {
		if listParams.AfterValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s > ? OR (feature_id > ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s < ? OR (feature_id < ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND feature_id > ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND feature_id < ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.BeforeId != "" {
		if listParams.BeforeValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s < ? OR (feature_id < ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s > ? OR (feature_id > ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND feature_id < ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND feature_id > ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	nullSortClause := "NULLS LAST"
	if listParams.SortOrder == middleware.SortOrderAsc {
		nullSortClause = "NULLS FIRST"
	}

	if listParams.SortBy != "featureId" {
		query = fmt.Sprintf("%s ORDER BY %s %s %s, feature_id %s LIMIT ?", query, sortBy, listParams.SortOrder, nullSortClause, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	} else {
		query = fmt.Sprintf("%s ORDER BY feature_id %s %s LIMIT ?", query, listParams.SortOrder, nullSortClause)
		replacements = append(replacements, listParams.Limit)
	}

	err := repo.DB.SelectContext(
		ctx,
		&features,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return models, service.NewInternalError("Unable to list features")
		}
	}

	for i := range features {
		models = append(models, &features[i])
	}

	return models, nil
}

func (repo PostgresRepository) UpdateByFeatureId(ctx context.Context, featureId string, model Model) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE feature
			SET
				name = ?,
				description = ?
			WHERE
				feature_id = ? AND
				deleted_at IS NULL
		`,
		model.GetName(),
		model.GetDescription(),
		featureId,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error updating feature %s", featureId))
	}

	return nil
}

func (repo PostgresRepository) DeleteByFeatureId(ctx context.Context, featureId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE feature
			SET
				deleted_at = ?
			WHERE
				feature_id = ? AND
				deleted_at IS NULL
		`,
		time.Now().UTC(),
		featureId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("Feature", featureId)
		default:
			return err
		}
	}

	return nil
}
