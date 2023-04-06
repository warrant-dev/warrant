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

type SQLiteRepository struct {
	database.SQLRepository
}

func NewSQLiteRepository(db *database.SQLite) SQLiteRepository {
	return SQLiteRepository{
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
				contextHash,
				createdAt,
				updatedAt
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (objectType, objectId, relation, subjectType, subjectId, subjectRelation, contextHash) DO UPDATE SET
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
		model.GetContextHash(),
		now,
		now,
		now,
	)
	if err != nil {
		// TODO: Cast to appropriate SQLite error
		// sqliteErr, ok := err.(*sqlite.SQLiteError)
		// if ok && sqliteErr.Number == 1062 {
		// 	return 0, service.NewDuplicateRecordError("Warrant", warrant, "Warrant for the given objectType, objectId, relation, and subject already exists")
		// }

		return 0, errors.Wrap(err, "Unable to create warrant")
	}

	return newWarrantId, nil
}

func (repo SQLiteRepository) DeleteById(ctx context.Context, id int64) error {
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
			return err
		}
	}

	return nil
}

func (repo SQLiteRepository) DeleteAllByObject(ctx context.Context, objectType string, objectId string) error {
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
			return errors.Wrap(err, fmt.Sprintf("Unable to delete warrants with object %s:%s from sqlite", objectType, objectId))
		}
	}

	return nil
}

func (repo SQLiteRepository) DeleteAllBySubject(ctx context.Context, subjectType string, subjectId string) error {
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
			return errors.Wrap(err, fmt.Sprintf("Unable to delete warrants with subject %s:%s from sqlite", subjectType, subjectId))
		}
	}

	return nil
}

func (repo SQLiteRepository) Get(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, contextHash string) (Model, error) {
	var warrant Warrant
	err := repo.DB.GetContext(
		ctx,
		&warrant,
		`
			SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, createdAt, updatedAt, deletedAt
			FROM warrant
			WHERE
				objectType = ? AND
				objectId = ? AND
				relation = ? AND
				subjectType = ? AND
				subjectId = ? AND
				subjectRelation = ? AND
				contextHash = ? AND
				deletedAt IS NULL
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

func (repo SQLiteRepository) GetWithContextMatch(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, contextHash string) (Model, error) {
	var warrant Warrant
	err := repo.DB.GetContext(
		ctx,
		&warrant,
		`
			SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, createdAt, updatedAt, deletedAt
			FROM warrant
			WHERE
				objectType = ? AND
				(objectId = ? OR objectId = "*") AND
				relation = ? AND
				subjectType = ? AND
				subjectId = ? AND
				subjectRelation = ? AND
				(contextHash = ? OR contextHash = "") AND
				deletedAt IS NULL
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

func (repo SQLiteRepository) GetByID(ctx context.Context, id int64) (Model, error) {
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
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get warrant %d from sqlite", id))
		}
	}

	return &warrant, nil
}

func (repo SQLiteRepository) List(ctx context.Context, filterOptions *FilterOptions, listParams middleware.ListParams) ([]Model, error) {
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
		query = fmt.Sprintf("%s AND subjectType = ? AND subjectId = ? AND subjectRelation = ?", query)
		replacements = append(replacements, filterOptions.Subject.ObjectType, filterOptions.Subject.ObjectId, filterOptions.Subject.Relation)
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
			return nil, err
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}

func (repo SQLiteRepository) GetAllMatchingWildcard(ctx context.Context, objectType string, objectId string, relation string, contextHash string) ([]Model, error) {
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
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get warrants matching object type %s and relation %s from sqlite", objectType, relation))
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}

func (repo SQLiteRepository) GetAllMatchingObjectAndRelation(ctx context.Context, objectType string, objectId string, relation string, subjectType string, contextHash string) ([]Model, error) {
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
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get warrants with object type %s, object id %s, and relation %s from sqlite", objectType, objectId, relation))
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}

func (repo SQLiteRepository) GetAllMatchingObjectAndSubject(ctx context.Context, objectType string, objectId string, subjectType string, subjectId string, subjectRelation string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, createdAt, updatedAt, deletedAt
			FROM warrant
			WHERE
				((objectType = ? AND objectId = ?) OR (subjectType = ? AND subjectId = ? AND subjectRelation = ?)) AND
				deletedAt IS NULL
			ORDER BY createdAt DESC, id DESC
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
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get warrants for object type %s, object id %s, and subject %s:%s#%s from sqlite", objectType, objectId, subjectType, subjectId, subjectRelation))
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}

func (repo SQLiteRepository) GetAllMatchingSubjectAndRelation(ctx context.Context, objectType string, relation string, subjectType string, subjectId string, subjectRelation string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, objectType, objectId, relation, subjectType, subjectId, subjectRelation, createdAt, updatedAt, deletedAt
			FROM warrant
			WHERE
				objectType = ? AND
				relation = ? AND
				subjectType = ? AND
				subjectId = ? AND
				subjectRelation = ? AND
				deletedAt IS NULL
			ORDER BY createdAt DESC, id DESC
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
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get warrants for object type %s, relation %s, and subject %s:%s#%s from sqlite", objectType, relation, subjectType, subjectId, subjectRelation))
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}
