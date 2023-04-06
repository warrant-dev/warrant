package authz

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/lib/pq"
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
	var newWarrantId int64
	err := repo.DB.GetContext(
		ctx,
		&newWarrantId,
		`
			INSERT INTO warrant (
				object_type,
				object_id,
				relation,
				subject_type,
				subject_id,
				subject_relation,
				context_hash
			) VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (object_type, object_id, relation, subject_type, subject_id, subject_relation, context_hash) DO UPDATE SET
				created_at = CURRENT_TIMESTAMP(6),
				deleted_at = NULL
			RETURNING id
		`,
		model.GetObjectType(),
		model.GetObjectId(),
		model.GetRelation(),
		model.GetSubjectType(),
		model.GetSubjectId(),
		model.GetSubjectRelation(),
		model.GetContextHash(),
	)
	if err != nil {
		postgresErr, ok := err.(*pq.Error)
		if ok && postgresErr.Code.Name() == "duplicate_object" {
			return 0, service.NewDuplicateRecordError("Warrant", model, "Warrant for the given objectType, objectId, relation, and subject already exists")
		}

		return 0, errors.Wrap(err, "Unable to create warrant")
	}

	return newWarrantId, nil
}

func (repo PostgresRepository) DeleteById(ctx context.Context, id int64) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE warrant
			SET deleted_at = ?
			WHERE
				id = ? AND
				deleted_at IS NULL
		`,
		time.Now().UTC(),
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("Warrant", id)
		default:
			return err
		}
	}

	return nil
}

func (repo PostgresRepository) DeleteAllByObject(ctx context.Context, objectType string, objectId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE warrant
			SET
				deleted_at = ?
			WHERE
				object_type = ? AND
				object_id = ? AND
				deleted_at IS NULL
		`,
		time.Now().UTC(),
		objectType,
		objectId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil
		default:
			return errors.Wrap(err, fmt.Sprintf("Unable to delete warrants with object %s:%s from mysql", objectType, objectId))
		}
	}

	return nil
}

func (repo PostgresRepository) DeleteAllBySubject(ctx context.Context, subjectType string, subjectId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE warrant
			SET
				deleted_at = ?
			WHERE
				subject_type = ? AND
				subject_id = ? AND
				deleted_at IS NULL
		`,
		time.Now().UTC(),
		subjectType,
		subjectId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil
		default:
			return errors.Wrap(err, fmt.Sprintf("Unable to delete warrants with subject %s:%s from mysql", subjectType, subjectId))
		}
	}

	return nil
}

func (repo PostgresRepository) Get(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, contextHash string) (Model, error) {
	var warrant Warrant
	err := repo.DB.GetContext(
		ctx,
		&warrant,
		`
			SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, created_at, updated_at, deleted_at
			FROM warrant
			WHERE
				object_type = ? AND
				object_id = ? AND
				relation = ? AND
				subject_type = ? AND
				subject_id = ? AND
				subject_relation = ? AND
				context_hash = ? AND
				deleted_at IS NULL
		`,
		objectType,
		objectId,
		relation,
		subjectType,
		subjectId,
		subjectRelation,
		contextHash,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Warrant", fmt.Sprintf("%s, %s, %s, %s:%s#%s", objectType, objectId, relation, subjectType, subjectId, subjectRelation))
		default:
			return nil, err
		}
	}

	return &warrant, nil
}

func (repo PostgresRepository) GetWithContextMatch(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, contextHash string) (Model, error) {
	var warrant Warrant
	err := repo.DB.GetContext(
		ctx,
		&warrant,
		`
			SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, created_at, updated_at, deleted_at
			FROM warrant
			WHERE
				object_type = ? AND
				(object_id = ? OR object_id = '*') AND
				relation = ? AND
				subject_type = ? AND
				subject_id = ? AND
				subject_relation = ? AND
				(context_hash = ? OR context_hash = '') AND
				deleted_at IS NULL
		`,
		objectType,
		objectId,
		relation,
		subjectType,
		subjectId,
		subjectRelation,
		contextHash,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &warrant, nil
}

func (repo PostgresRepository) GetByID(ctx context.Context, id int64) (Model, error) {
	var warrant Warrant
	err := repo.DB.GetContext(
		ctx,
		&warrant,
		`
			SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, created_at, updated_at, deleted_at
			FROM warrant
			WHERE
				id = ? AND
				deleted_at IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Warrant", id)
		default:
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get warrant %d from mysql", id))
		}
	}

	return &warrant, nil
}

func (repo PostgresRepository) List(ctx context.Context, filterOptions *FilterOptions, listParams middleware.ListParams) ([]Model, error) {
	offset := (listParams.Page - 1) * listParams.Limit
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	query := `
		SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, created_at, updated_at, deleted_at
		FROM warrant
		WHERE
			deleted_at IS NULL
	`
	replacements := []interface{}{}

	if filterOptions.ObjectType != "" {
		query = fmt.Sprintf(`%s AND object_type = ?`, query)
		replacements = append(replacements, filterOptions.ObjectType)
	}

	if filterOptions.ObjectId != "" {
		query = fmt.Sprintf(`%s AND object_id = ?`, query)
		replacements = append(replacements, filterOptions.ObjectId)
	}

	if filterOptions.Relation != "" {
		query = fmt.Sprintf(`%s AND relation = ?`, query)
		replacements = append(replacements, filterOptions.Relation)
	}

	if filterOptions.Subject != nil {
		query = fmt.Sprintf(`%s AND subject_type = ? AND subject_id = ? AND subject_relation = ?`, query)
		replacements = append(replacements, filterOptions.Subject.ObjectType, filterOptions.Subject.ObjectId, filterOptions.Subject.Relation)
	}

	if listParams.SortBy != "" {
		sortBy := regexp.MustCompile("([A-Z])").ReplaceAllString(listParams.SortBy, `_$1`)
		query = fmt.Sprintf(`%s ORDER BY %s %s`, query, sortBy, listParams.SortOrder)
	} else {
		query = fmt.Sprintf(`%s ORDER BY created_at DESC, id DESC`, query)
	}

	query = fmt.Sprintf(`%s LIMIT ? OFFSET ?`, query)
	replacements = append(replacements, listParams.Limit, offset)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			// Do nothing
		default:
			return nil, err
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}

func (repo PostgresRepository) GetAllMatchingWildcard(ctx context.Context, objectType string, objectId string, relation string, contextHash string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT
				w2.id,
				w2.object_type,
				w2.object_id,
				w2.relation,
				w1.subject_type,
				w1.subject_id,
				w1.subject_relation,
				w2.context_hash,
				w2.created_at,
				w2.updated_at
			FROM warrant AS w1
			JOIN warrant AS w2 ON
				w1.id != w2.id AND
				w1.object_type = w2.object_type AND
				w1.relation = w2.relation AND
				w1.context_hash = w2.context_hash
			WHERE
				w1.object_type = ? AND
				w1.object_id = '*' AND
				w2.object_id = ? AND
				w1.relation = ? AND
				(w1.context_hash = ? OR w1.context_hash = '') AND
				w1.deleted_at IS NULL AND
				w2.deleted_at IS NULL
			ORDER BY w2.created_at DESC, w2.id DESC
		`,
		objectType,
		objectId,
		relation,
		contextHash,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get warrants matching object type %s and relation %s from mysql", objectType, relation))
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}

func (repo PostgresRepository) GetAllMatchingObjectAndRelation(ctx context.Context, objectType string, objectId string, relation string, subjectType string, contextHash string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, context_hash, created_at, updated_at, deleted_at
			FROM warrant
			WHERE
				object_type = ? AND
				object_id = ? AND
				relation = ? AND
				subject_type = ? AND
				(context_hash = ? OR context_hash = '') AND
				deleted_at IS NULL
			ORDER BY created_at DESC, id DESC
		`,
		objectType,
		objectId,
		relation,
		subjectType,
		contextHash,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get warrants with object type %s, object id %s, and relation %s from mysql", objectType, objectId, relation))
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}

func (repo PostgresRepository) GetAllMatchingObjectAndSubject(ctx context.Context, objectType string, objectId string, subjectType string, subjectId string, subjectRelation string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, created_at, updated_at, deleted_at
			FROM warrant
			WHERE
				((object_type = ? AND object_id = ?) OR (subject_type = ? AND subject_id = ? AND subject_relation = ?)) AND
				deleted_at IS NULL
			ORDER BY created_at DESC, id DESC
		`,
		objectType,
		objectId,
		subjectType,
		subjectId,
		subjectRelation,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get warrants for object type %s, object id %s, and subject %s:%s#%s from mysql", objectType, objectId, subjectType, subjectId, subjectRelation))
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}

func (repo PostgresRepository) GetAllMatchingSubjectAndRelation(ctx context.Context, objectType string, relation string, subjectType string, subjectId string, subjectRelation string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, created_at, updated_at, deleted_at
			FROM warrant
			WHERE
				object_type = ? AND
				relation = ? AND
				subject_type = ? AND
				subject_id = ? AND
				subject_relation = ? AND
				deleted_at IS NULL
			ORDER BY created_at DESC, id DESC
		`,
		objectType,
		relation,
		subjectType,
		subjectId,
		subjectRelation,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get warrants for object type %s, relation %s, and subject %s:%s#%s from mysql", objectType, relation, subjectType, subjectId, subjectRelation))
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}
