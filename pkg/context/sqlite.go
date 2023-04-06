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

type SQLiteRepository struct {
	database.SQLRepository
}

func NewSQLiteRepository(db *database.SQLite) SQLiteRepository {
	return SQLiteRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repository SQLiteRepository) CreateAll(ctx context.Context, contexts []Context) ([]Context, error) {
	now := time.Now().UTC()
	for _, model := range contexts {
		model.CreatedAt = now
		model.UpdatedAt = now
	}
	_, err := repository.DB.NamedExecContext(
		ctx,
		`
			INSERT INTO context (
				warrantId,
				name,
				value,
				createdAt,
				updatedAt
			) VALUES (
				:warrantId,
				:name,
				:value,
				:createdAt,
				:updatedAt
			) ON CONFLICT (warrantId, name) DO UPDATE SET
				createdAt = EXCLUDED.createdAt,
				deletedAt = NULL
		`,
		contexts,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create contexts")
	}

	return repository.ListByWarrantId(ctx, []int64{contexts[0].WarrantId})
}

func (repository SQLiteRepository) ListByWarrantId(ctx context.Context, warrantIds []int64) ([]Context, error) {
	contexts := make([]Context, 0)
	if len(warrantIds) == 0 {
		return contexts, nil
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
			return contexts, nil
		default:
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to list contexts for warrant ids %s from sqlite", strings.Join(warrantIdStrings, ", ")))
		}
	}

	return contexts, nil
}

func (repository SQLiteRepository) DeleteAllByWarrantId(ctx context.Context, warrantId int64) error {
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
