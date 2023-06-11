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
	var newPricingTierId int64
	err := repo.DB(ctx).GetContext(
		ctx,
		&newPricingTierId,
		`
			INSERT INTO pricing_tier (
				object_id,
				pricing_tier_id,
				name,
				description
			) VALUES (?, ?, ?, ?)
			ON CONFLICT (pricing_tier_id) DO UPDATE SET
				object_id = ?,
				name = ?,
				description = ?,
				created_at = CURRENT_TIMESTAMP(6),
				deleted_at = NULL
			RETURNING id
		`,
		model.GetObjectId(),
		model.GetPricingTierId(),
		model.GetName(),
		model.GetDescription(),
		model.GetObjectId(),
		model.GetName(),
		model.GetDescription(),
	)

	if err != nil {
		return -1, errors.Wrap(err, "error creating pricing tier")
	}

	return newPricingTierId, err
}

func (repo PostgresRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var pricingTier PricingTier
	err := repo.DB(ctx).GetContext(
		ctx,
		&pricingTier,
		`
			SELECT id, object_id, pricing_tier_id, name, description, created_at, updated_at, deleted_at
			FROM pricing_tier
			WHERE
				id = ? AND
				deleted_at IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Model", id)
		default:
			return nil, errors.Wrapf(err, "error getting pricing tier id %d", id)
		}
	}

	return &pricingTier, nil
}

func (repo PostgresRepository) GetByPricingTierId(ctx context.Context, pricingTierId string) (Model, error) {
	var pricingTier PricingTier
	err := repo.DB(ctx).GetContext(
		ctx,
		&pricingTier,
		`
			SELECT id, object_id, pricing_tier_id, name, description, created_at, updated_at, deleted_at
			FROM pricing_tier
			WHERE
				pricing_tier_id = ? AND
				deleted_at IS NULL
		`,
		pricingTierId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Model", pricingTierId)
		default:
			return nil, errors.Wrapf(err, "error getting pricing tier %s", pricingTierId)
		}
	}

	return &pricingTier, nil
}

func (repo PostgresRepository) List(ctx context.Context, listParams service.ListParams) ([]Model, error) {
	models := make([]Model, 0)
	pricingTiers := make([]PricingTier, 0)
	query := `
		SELECT id, object_id, pricing_tier_id, name, description, created_at, updated_at, deleted_at
		FROM pricing_tier
		WHERE
			deleted_at IS NULL
	`
	replacements := []interface{}{}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (pricing_tier_id LIKE ? OR name LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	sortBy := regexp.MustCompile("([A-Z])").ReplaceAllString(listParams.SortBy, `_$1`)
	if listParams.AfterId != "" {
		if listParams.AfterValue != nil {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s > ? OR (pricing_tier_id > ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s < ? OR (pricing_tier_id < ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND pricing_tier_id > ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND pricing_tier_id < ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.BeforeId != "" {
		if listParams.BeforeValue != nil {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s < ? OR (pricing_tier_id < ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s > ? OR (pricing_tier_id > ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND pricing_tier_id < ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND pricing_tier_id > ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	nullSortClause := "NULLS LAST"
	if listParams.SortOrder == service.SortOrderAsc {
		nullSortClause = "NULLS FIRST"
	}

	if listParams.SortBy != "pricingTierId" {
		query = fmt.Sprintf("%s ORDER BY %s %s %s, pricing_tier_id %s LIMIT ?", query, sortBy, listParams.SortOrder, nullSortClause, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	} else {
		query = fmt.Sprintf("%s ORDER BY pricing_tier_id %s %s LIMIT ?", query, listParams.SortOrder, nullSortClause)
		replacements = append(replacements, listParams.Limit)
	}

	err := repo.DB(ctx).SelectContext(
		ctx,
		&pricingTiers,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return models, errors.Wrap(err, "error listing pricing tiers")
		}
	}

	for i := range pricingTiers {
		models = append(models, &pricingTiers[i])
	}

	return models, nil
}

func (repo PostgresRepository) UpdateByPricingTierId(ctx context.Context, pricingTierId string, model Model) error {
	_, err := repo.DB(ctx).ExecContext(
		ctx,
		`
			UPDATE pricing_tier
			SET
				name = ?,
				description = ?
			WHERE
				pricing_tier_id = ? AND
				deleted_at IS NULL
		`,
		model.GetName(),
		model.GetDescription(),
		pricingTierId,
	)
	if err != nil {
		return errors.Wrapf(err, "error updating pricing tier %s", pricingTierId)
	}

	return nil
}

func (repo PostgresRepository) DeleteByPricingTierId(ctx context.Context, pricingTierId string) error {
	_, err := repo.DB(ctx).ExecContext(
		ctx,
		`
			UPDATE pricing_tier
			SET
				deleted_at = ?
			WHERE
				pricing_tier_id = ? AND
				deleted_at IS NULL
		`,
		time.Now().UTC(),
		pricingTierId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("PricingTier", pricingTierId)
		default:
			return errors.Wrapf(err, "error deleting pricing tier %s", pricingTierId)
		}
	}

	return nil
}
