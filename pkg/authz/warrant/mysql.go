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

func (repo MySQLRepository) Create(ctx context.Context, model Model) (int64, error) {
	result, err := repo.DB.ExecContext(
		ctx,
		`
			INSERT INTO warrant (
				objectType,
				objectId,
				relation,
				subjectType,
				subjectId,
				subjectRelation,
				contextHash
			) VALUES (?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				createdAt = CURRENT_TIMESTAMP(6),
				deletedAt = NULL
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
		return -1, errors.Wrap(err, "error creating warrant")
	}

	newWarrantId, err := result.LastInsertId()
	if err != nil {
		return -1, errors.Wrap(err, "error creating warrant")
	}

	return newWarrantId, nil
}

func (repo MySQLRepository) DeleteById(ctx context.Context, id int64) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE warrant
			SET deletedAt = ?
			WHERE
				id = ? AND
				deletedAt IS NULL
		`,
		time.Now().UTC(),
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("Warrant", id)
		default:
			return errors.Wrapf(err, "error deleting warrant %d", id)
		}
	}

	return nil
}

func (repo MySQLRepository) DeleteAllByObject(ctx context.Context, objectType string, objectId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE warrant
			SET
				deletedAt = ?
			WHERE
				objectType = ? AND
				objectId = ? AND
				deletedAt IS NULL
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
			return errors.Wrapf(err, "error deleting warrants with object %s:%s", objectType, objectId)
		}
	}

	return nil
}

func (repo MySQLRepository) DeleteAllBySubject(ctx context.Context, subjectType string, subjectId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE warrant
			SET
				deletedAt = ?
			WHERE
				subjectType = ? AND
				subjectId = ? AND
				deletedAt IS NULL
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
			return errors.Wrapf(err, "error deleting warrants with subject %s:%s", subjectType, subjectId)
		}
	}

	return nil
}

func (repo MySQLRepository) Get(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation *string, contextHash string) (Model, error) {
	var warrant Warrant
	query := `
		SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, createdAt, updatedAt, deletedAt
		FROM warrant
		WHERE
			objectType = ? AND
			objectId = ? AND
			relation = ? AND
			subjectType = ? AND
			subjectId = ? AND
			contextHash = ? AND
			deletedAt IS NULL
	`
	replacements := []interface{}{
		objectType,
		objectId,
		relation,
		subjectType,
		subjectId,
		contextHash,
	}
	if subjectRelation != nil {
		query = fmt.Sprintf("%s AND subjectRelation = ?", query)
		replacements = append(replacements, subjectRelation)
	} else {
		query = fmt.Sprintf("%s AND subjectRelation IS NULL", query)
	}

	err := repo.DB.GetContext(
		ctx,
		&warrant,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			wntErrorId := fmt.Sprintf("%s:%s#%s@%s:%s", objectType, objectId, relation, subjectType, subjectId)
			if subjectRelation != nil {
				wntErrorId = fmt.Sprintf("%s#%s", wntErrorId, *subjectRelation)
			}

			return nil, service.NewRecordNotFoundError("Warrant", wntErrorId)
		default:
			return nil, errors.Wrap(err, "error getting warrant")
		}
	}

	return &warrant, nil
}

func (repo MySQLRepository) GetWithContextMatch(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation *string, contextHash string) (Model, error) {
	var warrant Warrant
	query := `
		SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, createdAt, updatedAt, deletedAt
		FROM warrant
		WHERE
			objectType = ? AND
			(objectId = ? OR objectId = "*") AND
			relation = ? AND
			subjectType = ? AND
			subjectId = ? AND
			(contextHash = ? OR contextHash = "") AND
			deletedAt IS NULL
	`
	replacements := []interface{}{
		objectType,
		objectId,
		relation,
		subjectType,
		subjectId,
		contextHash,
	}
	if subjectRelation != nil {
		query = fmt.Sprintf("%s AND subjectRelation = ?", query)
		replacements = append(replacements, subjectRelation)
	} else {
		query = fmt.Sprintf("%s AND subjectRelation IS NULL", query)
	}

	err := repo.DB.GetContext(
		ctx,
		&warrant,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, errors.Wrap(err, "error getting warrant with context match")
		}
	}

	return &warrant, nil
}

func (repo MySQLRepository) GetByID(ctx context.Context, id int64) (Model, error) {
	var warrant Warrant
	err := repo.DB.GetContext(
		ctx,
		&warrant,
		`
			SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, createdAt, updatedAt, deletedAt
			FROM warrant
			WHERE
				id = ? AND
				deletedAt IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Warrant", id)
		default:
			return nil, errors.Wrapf(err, "error getting warrant %d", id)
		}
	}

	return &warrant, nil
}

func (repo MySQLRepository) List(ctx context.Context, filterOptions *FilterOptions, listParams middleware.ListParams) ([]Model, error) {
	offset := (listParams.Page - 1) * listParams.Limit
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	query := `
		SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, createdAt, updatedAt, deletedAt
		FROM warrant
		WHERE
			deletedAt IS NULL
	`
	replacements := []interface{}{}

	if filterOptions.ObjectType != "" {
		query = fmt.Sprintf("%s AND objectType = ?", query)
		replacements = append(replacements, filterOptions.ObjectType)
	}

	if filterOptions.ObjectId != "" {
		query = fmt.Sprintf("%s AND objectId = ?", query)
		replacements = append(replacements, filterOptions.ObjectId)
	}

	if filterOptions.Relation != "" {
		query = fmt.Sprintf("%s AND relation = ?", query)
		replacements = append(replacements, filterOptions.Relation)
	}

	if filterOptions.Subject != nil {
		query = fmt.Sprintf("%s AND subjectType = ? AND subjectId = ?", query)
		replacements = append(replacements, filterOptions.Subject.ObjectType, filterOptions.Subject.ObjectId)

		if filterOptions.Subject.Relation != nil {
			query = fmt.Sprintf("%s AND subjectRelation = ?", query)
			replacements = append(replacements, filterOptions.Subject.Relation)
		}
	}

	if listParams.SortBy != "" {
		query = fmt.Sprintf("%s ORDER BY %s %s", query, listParams.SortBy, listParams.SortOrder)
	} else {
		query = fmt.Sprintf("%s ORDER BY createdAt DESC, id DESC", query)
	}

	query = fmt.Sprintf("%s LIMIT ?, ?", query)
	replacements = append(replacements, offset, listParams.Limit)
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
			return nil, errors.Wrap(err, "error listing warrants")
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}

func (repo MySQLRepository) GetAllMatchingWildcard(ctx context.Context, objectType string, objectId string, relation string, contextHash string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT
				w2.id,
				w2.objectType,
				w2.objectId,
				w2.relation,
				w1.subjectType,
				w1.subjectId,
				w1.subjectRelation,
				w2.contextHash,
				w2.createdAt,
				w2.updatedAt
			FROM warrant AS w1
			JOIN warrant AS w2 ON
				w1.id != w2.id AND
				w1.objectType = w2.objectType AND
				w1.relation = w2.relation AND
				w1.contextHash = w2.contextHash
			WHERE
				w1.objectType = ? AND
				w1.objectId = "*" AND
				w2.objectId = ? AND
				w1.relation = ? AND
				(w1.contextHash = ? OR w1.contextHash = "") AND
				w1.deletedAt IS NULL AND
				w2.deletedAt IS NULL
			ORDER BY w2.createdAt DESC, w2.id DESC
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
			return nil, errors.Wrapf(err, "error getting warrants matching object type %s and relation %s", objectType, relation)
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}

func (repo MySQLRepository) GetAllMatchingObjectAndRelation(ctx context.Context, objectType string, objectId string, relation string, subjectType string, contextHash string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, contextHash, createdAt, updatedAt, deletedAt
			FROM warrant
			WHERE
				objectType = ? AND
				objectId = ? AND
				relation = ? AND
				subjectType = ? AND
				(contextHash = ? OR contextHash = "") AND
				deletedAt IS NULL
			ORDER BY createdAt DESC, id DESC
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
			return nil, errors.Wrapf(err, "error getting warrants with object type %s, object id %s, and relation %s", objectType, objectId, relation)
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}
