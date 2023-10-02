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

func (repo PostgresRepository) List(ctx context.Context, filterParams *FilterParams, listParams service.ListParams) ([]Model, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	query := `
		SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, policy, created_at, updated_at, deleted_at
		FROM warrant
		WHERE
			deleted_at IS NULL
	`
	replacements := []interface{}{}
	defaultSort := sortRegexp.ReplaceAllString(listParams.DefaultSortBy(), `_$1`)
	sortByColumn := sortRegexp.ReplaceAllString(listParams.SortBy, `_$1`)

	if len(filterParams.ObjectType) > 0 {
		query = fmt.Sprintf("%s AND object_type IN (%s)", query, BuildQuestionMarkString(len(filterParams.ObjectType)+1))
		for _, objectType := range filterParams.ObjectType {
			replacements = append(replacements, objectType)
		}
		replacements = append(replacements, Wildcard)
	}

	if len(filterParams.ObjectId) > 0 {
		query = fmt.Sprintf("%s AND object_id IN (%s)", query, BuildQuestionMarkString(len(filterParams.ObjectId)+1))
		for _, objectId := range filterParams.ObjectId {
			replacements = append(replacements, objectId)
		}
		replacements = append(replacements, Wildcard)
	}

	if len(filterParams.Relation) > 0 {
		query = fmt.Sprintf("%s AND relation IN (%s)", query, BuildQuestionMarkString(len(filterParams.Relation)+1))
		for _, relation := range filterParams.Relation {
			replacements = append(replacements, relation)
		}
		replacements = append(replacements, Wildcard)
	}

	if len(filterParams.SubjectType) > 0 {
		query = fmt.Sprintf("%s AND subject_type IN (%s)", query, BuildQuestionMarkString(len(filterParams.SubjectType)+1))
		for _, subjectType := range filterParams.SubjectType {
			replacements = append(replacements, subjectType)
		}
		replacements = append(replacements, Wildcard)
	}

	if len(filterParams.SubjectId) > 0 {
		query = fmt.Sprintf("%s AND subject_id IN (%s)", query, BuildQuestionMarkString(len(filterParams.SubjectId)+1))
		for _, subjectId := range filterParams.SubjectId {
			replacements = append(replacements, subjectId)
		}
		replacements = append(replacements, Wildcard)
	}

	if len(filterParams.SubjectRelation) > 0 {
		query = fmt.Sprintf("%s AND subject_relation IN (%s)", query, BuildQuestionMarkString(len(filterParams.SubjectRelation)+1))
		for _, subjectRelation := range filterParams.SubjectRelation {
			replacements = append(replacements, subjectRelation)
		}
		replacements = append(replacements, Wildcard)
	}

	if filterParams.Policy != "" {
		query = fmt.Sprintf("%s AND policy_hash = ?", query)
		replacements = append(replacements, filterParams.Policy.Hash())
	}

	if listParams.AfterId != nil {
		comparisonOp := "<"
		if listParams.SortOrder == service.SortOrderAsc {
			comparisonOp = ">"
		}

		switch listParams.AfterValue {
		case nil:
			//nolint:gocritic
			if listParams.SortBy == listParams.DefaultSortBy() {
				query = fmt.Sprintf("%s AND %s %s ?", query, defaultSort, comparisonOp)
				replacements = append(replacements, listParams.AfterId)
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, sortByColumn, defaultSort, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.AfterId,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, defaultSort, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.AfterId,
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, defaultSort, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, sortByColumn, defaultSort, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		}
	}

	if listParams.BeforeId != nil {
		comparisonOp := ">"
		if listParams.SortOrder == service.SortOrderAsc {
			comparisonOp = "<"
		}

		switch listParams.BeforeValue {
		case nil:
			//nolint:gocritic
			if listParams.SortBy == listParams.DefaultSortBy() {
				query = fmt.Sprintf("%s AND %s %s ?", query, defaultSort, comparisonOp)
				replacements = append(replacements, listParams.BeforeId)
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, defaultSort, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.BeforeId,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, sortByColumn, defaultSort, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.BeforeId,
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, sortByColumn, defaultSort, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, defaultSort, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		}
	}

	nullSortClause := "NULLS LAST"
	invertedNullSortClause := "NULLS FIRST"
	if listParams.SortOrder == service.SortOrderAsc {
		nullSortClause = "NULLS FIRST"
		invertedNullSortClause = "NULLS LAST"
	}

	if listParams.BeforeId != nil {
		if listParams.SortBy != listParams.DefaultSortBy() {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortByColumn, service.SortOrderDesc, invertedNullSortClause, defaultSort, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortByColumn, service.SortOrderAsc, invertedNullSortClause, defaultSort, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s %s, %s %s", query, sortByColumn, listParams.SortOrder, nullSortClause, defaultSort, listParams.SortOrder)
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, sortByColumn, service.SortOrderDesc, invertedNullSortClause)
				replacements = append(replacements, listParams.Limit)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, sortByColumn, service.SortOrderAsc, invertedNullSortClause)
				replacements = append(replacements, listParams.Limit)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s %s", query, sortByColumn, listParams.SortOrder, nullSortClause)
		}
	} else {
		if listParams.SortBy != listParams.DefaultSortBy() {
			query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortByColumn, listParams.SortOrder, nullSortClause, defaultSort, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, defaultSort, listParams.SortOrder, nullSortClause)
			replacements = append(replacements, listParams.Limit)
		}
	}

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
