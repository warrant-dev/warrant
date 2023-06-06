package authz

import (
	"context"
	"database/sql"
	"fmt"
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
				policy,
				policyHash
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
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
		model.GetPolicy(),
		model.GetPolicyHash(),
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

func (repo MySQLRepository) get(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, policyHash string) (Model, error) {
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

func (repo MySQLRepository) getByID(ctx context.Context, id int64) (Model, error) {
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
			return nil, errors.Wrapf(err, "error getting warrant %d", id)
		}
	}

	return &warrant, nil
}

func (repo MySQLRepository) List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]Model, error) {
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

func (repo MySQLRepository) GetAllMatchingObjectRelationAndSubject(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string) ([]Model, error) {
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

func (repo MySQLRepository) GetAllMatchingObjectAndRelation(ctx context.Context, objectType string, objectId string, relation string) ([]Model, error) {
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

func (repo MySQLRepository) GetAllMatchingObjectAndRelationBySubjectType(ctx context.Context, objectType string, objectId string, relation string, subjectType string) ([]Model, error) {
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
