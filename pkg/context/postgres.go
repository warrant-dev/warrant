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

type PostgresRepository struct {
	database.SQLRepository
}

func NewPostgresRepository(db *database.Postgres) PostgresRepository {
	return PostgresRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repository PostgresRepository) CreateAll(ctx context.Context, models []Model) ([]Model, error) {
	contexts := make([]Context, 0)
	for _, model := range models {
		contexts = append(contexts, *NewContextFromModel(model))
	}

	_, err := repository.DB.NamedExecContext(
		ctx,
		`
			INSERT INTO context (
				warrant_id,
				name,
				value
			) VALUES (
				:warrant_id,
				:name,
				:value
			) ON CONFLICT (warrant_id, name) DO UPDATE SET
				created_at = CURRENT_TIMESTAMP(6),
				deleted_at = NULL
		`,
		contexts,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating contexts")
	}

	return repository.ListByWarrantId(ctx, []int64{contexts[0].GetWarrantId()})
}

func (repository PostgresRepository) ListByWarrantId(ctx context.Context, warrantIds []int64) ([]Model, error) {
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
				SELECT id, warrant_id, name, value, created_at, updated_at, deleted_at
				FROM context
				WHERE
					warrant_id IN (%s) AND
					deleted_at IS NULL
			`,
			strings.Join(warrantIdStrings, ", "),
		),
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return nil, errors.Wrapf(err, "error listing contexts for warrant ids %s", strings.Join(warrantIdStrings, ", "))
		}
	}

	for i := range contexts {
		models = append(models, &contexts[i])
	}

	return models, nil
}

func (repository PostgresRepository) DeleteAllByWarrantId(ctx context.Context, warrantId int64) error {
	_, err := repository.DB.ExecContext(
		ctx,
		`
			UPDATE context
			SET
				deleted_at = ?
			WHERE
				warrant_id = ? AND
				deleted_at IS NULL
		`,
		time.Now().UTC(),
		warrantId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("Warrant", warrantId)
		default:
			return errors.Wrapf(err, "error deleting contexts for warrant %d", warrantId)
		}
	}

	return nil
}
