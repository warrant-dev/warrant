package context

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
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

func (repository MySQLRepository) CreateAll(ctx context.Context, models []Model) ([]Model, error) {
	contexts := make([]Context, 0)
	for _, model := range models {
		contexts = append(contexts, *NewContextFromModel(model))
	}

	_, err := repository.DB.NamedExecContext(
		ctx,
		`
			INSERT INTO context (
				warrantId,
				name,
				value
			) VALUES (
				:warrantId,
				:name,
				:value
			) ON DUPLICATE KEY UPDATE
				createdAt = CURRENT_TIMESTAMP(6),
				deletedAt = NULL
		`,
		contexts,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create contexts")
	}

	return repository.ListByWarrantId(ctx, []int64{contexts[0].GetWarrantId()})
}

func (repository MySQLRepository) ListByWarrantId(ctx context.Context, warrantIds []int64) ([]Model, error) {
	models := make([]Model, 0)
	contexts := make([]Context, 0)
	if len(warrantIds) == 0 {
		return models, nil
	}

	warrantIdStrings := make([]string, 0)
	for _, warrantId := range warrantIds {
		warrantIdStrings = append(warrantIdStrings, strconv.FormatInt(warrantId, 10))
	}

	err := repository.DB.SelectContext(
		ctx,
		&contexts,
		fmt.Sprintf(
			`
				SELECT id, warrantId, name, value, createdAt, updatedAt, deletedAt
				FROM context
				WHERE
					warrantId IN (%s) AND
					deletedAt IS NULL
			`,
			strings.Join(warrantIdStrings, ", "),
		),
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to list contexts for warrant ids %s from mysql", strings.Join(warrantIdStrings, ", ")))
		}
	}

	for i := range contexts {
		models = append(models, &contexts[i])
	}

	return models, nil
}

func (repository MySQLRepository) DeleteAllByWarrantId(ctx context.Context, warrantId int64) error {
	_, err := repository.DB.ExecContext(
		ctx,
		`
			UPDATE context
			SET
				deletedAt = ?
			WHERE
				warrantId = ? AND
				deletedAt IS NULL
		`,
		time.Now().UTC(),
		warrantId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("Warrant", warrantId)
		default:
			return err
		}
	}

	return nil
}
