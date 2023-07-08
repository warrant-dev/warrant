package authz

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

type SQLiteRepository struct {
	database.SQLRepository
}

func NewSQLiteRepository(db *database.SQLite) *SQLiteRepository {
	return &SQLiteRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo SQLiteRepository) Create(ctx context.Context, model Model) (int64, error) {
	var newWarrantId int64
	now := time.Now().UTC()
	err := repo.DB.GetContext(
		ctx,
		&newWarrantId,
		`
			INSERT INTO warrant (
				objectType,
				objectId,
				relation,
				subjectType,
				subjectId,
				subjectRelation,
				policy,
				policyHash,
				createdAt,
				updatedAt
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (objectType, objectId, relation, subjectType, subjectId, subjectRelation, policyHash) DO UPDATE SET
				createdAt = ?,
				deletedAt = NULL
			RETURNING id
		`,
		model.GetObjectType(),
		model.GetObjectId(),
		model.GetRelation(),
		model.GetSubjectType(),
		model.GetSubjectId(),
		model.GetSubjectRelation(),
		model.GetPolicy(),
		model.GetPolicy().Hash(),
		now,
		now,
		now,
	)
	if err != nil {
		return -1, errors.Wrap(err, "error creating warrant")
	}

	return newWarrantId, nil
}

func (repo SQLiteRepository) DeleteById(ctx context.Context, ids []int64) error {
	query, args, err := sqlx.In(
		`
			UPDATE warrant
			SET deletedAt = ?
			WHERE
				id IN (?) AND
				deletedAt IS NULL
		`,
		time.Now().UTC(),
		ids,
	)
	if err != nil {
		return errors.Wrapf(err, "error deleting warrants %v", ids)
	}
	_, err = repo.DB.ExecContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil
		default:
			return errors.Wrapf(err, "error deleting warrants %v", ids)
		}
	}

	return nil
}

func (repo SQLiteRepository) Delete(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, policyHash string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE warrant
			SET
				deletedAt = ?
			WHERE
				objectType = ? AND
				objectId = ? AND
				relation = ? AND
				subjectType = ? AND
				subjectId = ? AND
				subjectRelation = ? AND
				policyHash = ? AND
				deletedAt IS NULL
		`,
		time.Now().UTC(),
		objectType,
		objectId,
		relation,
		subjectType,
		subjectId,
		subjectRelation,
		policyHash,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			wntErrorId := fmt.Sprintf("%s:%s#%s@%s:%s", objectType, objectId, relation, subjectType, subjectId)
			if subjectRelation != "" {
				wntErrorId = fmt.Sprintf("%s#%s", wntErrorId, subjectRelation)
			}
			if policyHash != "" {
				wntErrorId = fmt.Sprintf("%s[%s]", wntErrorId, policyHash)
			}

			return service.NewRecordNotFoundError("Warrant", wntErrorId)
		default:
			return errors.Wrap(err, "error deleting warrant")
		}
	}

	return nil
}

func (repo SQLiteRepository) GetAllMatchingObject(ctx context.Context, objectType string, objectId string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, policy, createdAt, updatedAt, deletedAt
			FROM warrant
			WHERE
				objectType = ? AND
				objectId = ? AND
				deletedAt IS NULL
		`,
		objectType,
		objectId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return models, errors.Wrapf(err, "error deleting warrants with object %s:%s", objectType, objectId)
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}

func (repo SQLiteRepository) GetAllMatchingSubject(ctx context.Context, subjectType string, subjectId string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, policy, createdAt, updatedAt, deletedAt
			FROM warrant
			WHERE
				subjectType = ? AND
				subjectId = ? AND
				deletedAt IS NULL
		`,
		subjectType,
		subjectId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return models, errors.Wrapf(err, "error deleting warrants with subject %s:%s", subjectType, subjectId)
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}

func (repo SQLiteRepository) Get(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, policyHash string) (Model, error) {
	var warrant Warrant
	err := repo.DB.GetContext(
		ctx,
		&warrant,
		`
			SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, policy, createdAt, updatedAt, deletedAt
			FROM warrant
			WHERE
				objectType = ? AND
				objectId = ? AND
				relation = ? AND
				subjectType = ? AND
				subjectId = ? AND
				subjectRelation = ? AND
				policyHash = ? AND
				deletedAt IS NULL
		`,
		objectType,
		objectId,
		relation,
		subjectType,
		subjectId,
		subjectRelation,
		policyHash,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			wntErrorId := fmt.Sprintf("%s:%s#%s@%s:%s", objectType, objectId, relation, subjectType, subjectId)
			if subjectRelation != "" {
				wntErrorId = fmt.Sprintf("%s#%s", wntErrorId, subjectRelation)
			}
			if policyHash != "" {
				wntErrorId = fmt.Sprintf("%s[%s]", wntErrorId, policyHash)
			}

			return nil, service.NewRecordNotFoundError("Warrant", wntErrorId)
		default:
			return nil, errors.Wrap(err, "error getting warrant")
		}
	}

	return &warrant, nil
}

func (repo SQLiteRepository) GetByID(ctx context.Context, id int64) (Model, error) {
	var warrant Warrant
	err := repo.DB.GetContext(
		ctx,
		&warrant,
		`
			SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, policy, createdAt, updatedAt, deletedAt
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
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get warrant %d from sqlite", id))
		}
	}

	return &warrant, nil
}

func (repo SQLiteRepository) List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]Model, error) {
	offset := (listParams.Page - 1) * listParams.Limit
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	query := `
		SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, policy, createdAt, updatedAt, deletedAt
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

	if filterOptions.Subject.ObjectType != "" {
		query = fmt.Sprintf("%s AND subjectType = ?", query)
		replacements = append(replacements, filterOptions.Subject.ObjectType)
	}

	if filterOptions.Subject.ObjectId != "" {
		query = fmt.Sprintf("%s AND subjectId = ?", query)
		replacements = append(replacements, filterOptions.Subject.ObjectId)
	}

	if filterOptions.Subject.Relation != "" {
		query = fmt.Sprintf("%s AND subjectRelation = ?", query)
		replacements = append(replacements, filterOptions.Subject.Relation)
	}

	if filterOptions.Policy != "" {
		query = fmt.Sprintf("%s AND policyHash = ?", query)
		replacements = append(replacements, filterOptions.Policy.Hash())
	}

	if listParams.SortBy != "" {
		query = fmt.Sprintf("%s ORDER BY %s %s", query, listParams.SortBy, listParams.SortOrder)
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

func (repo SQLiteRepository) GetAllMatchingObjectRelationAndSubject(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, policy, createdAt, updatedAt, deletedAt
			FROM warrant
			WHERE
				objectType = ? AND
				(objectId = ? OR objectId = "*") AND
				relation = ? AND
				subjectType = ? AND
				subjectId = ? AND
				subjectRelation = ? AND
				deletedAt IS NULL
		`,
		objectType,
		objectId,
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
			return nil, errors.Wrapf(err, "error getting warrants with object type %s, object id %s, relation %s, subject type %s, subject id %s, and subject relation %s", objectType, objectId, relation, subjectType, subjectId, subjectRelation)
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}

func (repo SQLiteRepository) GetAllMatchingObjectAndRelation(ctx context.Context, objectType string, objectId string, relation string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, policy, createdAt, updatedAt, deletedAt
			FROM warrant
			WHERE
				objectType = ? AND
				(objectId = ? OR objectId = "*") AND
				relation = ? AND
				deletedAt IS NULL
			ORDER BY createdAt DESC, id DESC
		`,
		objectType,
		objectId,
		relation,
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

func (repo SQLiteRepository) GetAllMatchingObjectAndRelationBySubjectType(ctx context.Context, objectType string, objectId string, relation string, subjectType string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, policy, createdAt, updatedAt, deletedAt
			FROM warrant
			WHERE
				objectType = ? AND
				(objectId = ? OR objectId = "*") AND
				relation = ? AND
				subjectType = ? AND
				deletedAt IS NULL
			ORDER BY createdAt DESC, id DESC
		`,
		objectType,
		objectId,
		relation,
		subjectType,
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
