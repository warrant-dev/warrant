// Copyright 2023 Forerunner Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authz

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

var sortRegexp = regexp.MustCompile("([A-Z])")

type PostgresRepository struct {
	database.SQLRepository
}

func NewPostgresRepository(db *database.Postgres) *PostgresRepository {
	return &PostgresRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo PostgresRepository) Create(ctx context.Context, model Model) (int64, error) {
	var newWarrantId int64
	err := repo.DB.GetContext(
		database.CtxWithWriterOverride(ctx),
		&newWarrantId,
		`
			INSERT INTO warrant (
				object_type,
				object_id,
				relation,
				subject_type,
				subject_id,
				subject_relation,
				policy,
				policy_hash
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (object_type, object_id, relation, subject_type, subject_id, subject_relation, policy_hash) DO UPDATE SET
				created_at = CURRENT_TIMESTAMP(6),
				updated_at = CURRENT_TIMESTAMP(6),
				deleted_at = NULL
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
	)
	if err != nil {
		return -1, errors.Wrap(err, "error creating warrant")
	}

	return newWarrantId, nil
}

func (repo PostgresRepository) DeleteById(ctx context.Context, ids []int64) error {
	query, args, err := sqlx.In(
		`
			UPDATE warrant
			SET
				updated_at = CURRENT_TIMESTAMP(6),
				deleted_at = CURRENT_TIMESTAMP(6)
			WHERE
				id IN (?) AND
				deleted_at IS NULL
		`,
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

func (repo PostgresRepository) Delete(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, policyHash string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE warrant
			SET
				updated_at = CURRENT_TIMESTAMP(6),
				deleted_at = CURRENT_TIMESTAMP(6)
			WHERE
				object_type = ? AND
				object_id = ? AND
				relation = ? AND
				subject_type = ? AND
				subject_id = ? AND
				subject_relation = ? AND
				policy_hash = ? AND
				deleted_at IS NULL
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

			return service.NewRecordNotFoundError("Warrant", wntErrorId)
		default:
			return errors.Wrap(err, "error deleting warrant")
		}
	}

	return nil
}

func (repo PostgresRepository) GetAllMatchingObject(ctx context.Context, objectType string, objectId string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, policy, created_at, updated_at, deleted_at
			FROM warrant
			WHERE
				object_type = ? AND
				object_id = ? AND
				deleted_at IS NULL
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

func (repo PostgresRepository) GetAllMatchingSubject(ctx context.Context, subjectType string, subjectId string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, policy, created_at, updated_at, deleted_at
			FROM warrant
			WHERE
				subject_type = ? AND
				subject_id = ? AND
				deleted_at IS NULL
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

func (repo PostgresRepository) Get(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, policyHash string) (Model, error) {
	var warrant Warrant
	err := repo.DB.GetContext(
		ctx,
		&warrant,
		`
			SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, policy, created_at, updated_at, deleted_at
			FROM warrant
			WHERE
				object_type = ? AND
				object_id = ? AND
				relation = ? AND
				subject_type = ? AND
				subject_id = ? AND
				subject_relation = ? AND
				policy_hash = ? AND
				deleted_at IS NULL
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

func (repo PostgresRepository) GetByID(ctx context.Context, id int64) (Model, error) {
	var warrant Warrant
	err := repo.DB.GetContext(
		ctx,
		&warrant,
		`
			SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, policy, created_at, updated_at, deleted_at
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

func (repo PostgresRepository) List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]Model, error) {
	offset := (listParams.Page - 1) * listParams.Limit
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	query := `
		SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, policy, created_at, updated_at, deleted_at
		FROM warrant
		WHERE
			deleted_at IS NULL
	`
	replacements := []interface{}{}

	if len(filterOptions.ObjectType) > 0 {
		query = fmt.Sprintf("%s AND object_type IN (%s)", query, buildQuestionMarkString(len(filterOptions.ObjectType)))
		for _, objectType := range filterOptions.ObjectType {
			replacements = append(replacements, objectType)
		}
	}

	if len(filterOptions.ObjectId) > 0 {
		query = fmt.Sprintf("%s AND object_id IN (%s)", query, buildQuestionMarkString(len(filterOptions.ObjectId)))
		for _, objectId := range filterOptions.ObjectId {
			replacements = append(replacements, objectId)
		}
	}

	if len(filterOptions.Relation) > 0 {
		query = fmt.Sprintf("%s AND relation IN (%s)", query, buildQuestionMarkString(len(filterOptions.Relation)))
		for _, relation := range filterOptions.Relation {
			replacements = append(replacements, relation)
		}
	}

	if len(filterOptions.SubjectType) > 0 {
		query = fmt.Sprintf("%s AND subject_type IN (%s)", query, buildQuestionMarkString(len(filterOptions.SubjectType)))
		for _, subjectType := range filterOptions.SubjectType {
			replacements = append(replacements, subjectType)
		}
	}

	if len(filterOptions.SubjectId) > 0 {
		query = fmt.Sprintf("%s AND subject_id IN (%s)", query, buildQuestionMarkString(len(filterOptions.SubjectId)))
		for _, subjectId := range filterOptions.SubjectId {
			replacements = append(replacements, subjectId)
		}
	}

	if len(filterOptions.SubjectRelation) > 0 {
		query = fmt.Sprintf("%s AND subject_relation IN (%s)", query, buildQuestionMarkString(len(filterOptions.SubjectRelation)))
		for _, subjectRelation := range filterOptions.SubjectRelation {
			replacements = append(replacements, subjectRelation)
		}
	}

	if filterOptions.Policy != "" {
		query = fmt.Sprintf("%s AND policy_hash = ?", query)
		replacements = append(replacements, filterOptions.Policy.Hash())
	}

	if listParams.SortBy != "" {
		sortBy := sortRegexp.ReplaceAllString(listParams.SortBy, `_$1`)
		query = fmt.Sprintf(`%s ORDER BY %s %s`, query, sortBy, listParams.SortOrder)
	}

	query = fmt.Sprintf("%s LIMIT ? OFFSET ?", query)
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
			return nil, errors.Wrap(err, "error listing warrants")
		}
	}

	for i := range warrants {
		models = append(models, &warrants[i])
	}

	return models, nil
}

func (repo PostgresRepository) GetAllMatchingObjectRelationAndSubject(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, policy, created_at, updated_at, deleted_at
			FROM warrant
			WHERE
				object_type = ? AND
				(object_id = ? OR object_id = '*') AND
				relation = ? AND
				subject_type = ? AND
				subject_id = ? AND
				subject_relation = ? AND
				deleted_at IS NULL
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

func (repo PostgresRepository) GetAllMatchingObjectAndRelation(ctx context.Context, objectType string, objectId string, relation string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, policy, created_at, updated_at, deleted_at
			FROM warrant
			WHERE
				object_type = ? AND
				(object_id = ? OR object_id = '*') AND
				relation = ? AND
				deleted_at IS NULL
			ORDER BY created_at DESC, id DESC
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

func (repo PostgresRepository) GetAllMatchingObjectAndRelationBySubjectType(ctx context.Context, objectType string, objectId string, relation string, subjectType string) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		`
			SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, policy, created_at, updated_at, deleted_at
			FROM warrant
			WHERE
				object_type = ? AND
				(object_id = ? OR object_id = '*') AND
				relation = ? AND
				subject_type = ? AND
				deleted_at IS NULL
			ORDER BY created_at DESC, id DESC
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
