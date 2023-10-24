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
	"strconv"

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
				created_at = CASE
					WHEN warrant.deleted_at IS NULL THEN warrant.created_at
					ELSE CURRENT_TIMESTAMP(6)
				END,
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
		if errors.Is(err, sql.ErrNoRows) {
			wntErrorId := fmt.Sprintf("%s:%s#%s@%s:%s", objectType, objectId, relation, subjectType, subjectId)
			if subjectRelation != "" {
				wntErrorId = fmt.Sprintf("%s#%s", wntErrorId, subjectRelation)
			}
			if policyHash != "" {
				wntErrorId = fmt.Sprintf("%s[%s]", wntErrorId, policyHash)
			}

			return service.NewRecordNotFoundError("Warrant", wntErrorId)
		}
		return errors.Wrap(err, "error deleting warrant")
	}

	return nil
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
		if errors.Is(err, sql.ErrNoRows) {
			wntErrorId := fmt.Sprintf("%s:%s#%s@%s:%s", objectType, objectId, relation, subjectType, subjectId)
			if subjectRelation != "" {
				wntErrorId = fmt.Sprintf("%s#%s", wntErrorId, subjectRelation)
			}
			if policyHash != "" {
				wntErrorId = fmt.Sprintf("%s[%s]", wntErrorId, policyHash)
			}

			return nil, service.NewRecordNotFoundError("Warrant", wntErrorId)
		}
		return nil, errors.Wrap(err, "error getting warrant")
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.NewRecordNotFoundError("Warrant", id)
		}
		return nil, errors.Wrap(err, fmt.Sprintf("Unable to get warrant %d from mysql", id))
	}

	return &warrant, nil
}

func (repo PostgresRepository) List(ctx context.Context, filterParams FilterParams, listParams service.ListParams) ([]Model, *service.Cursor, *service.Cursor, error) {
	models := make([]Model, 0)
	warrants := make([]Warrant, 0)
	query := `
		SELECT id, object_type, object_id, relation, subject_type, subject_id, subject_relation, policy, created_at, updated_at, deleted_at
		FROM warrant
		WHERE
			deleted_at IS NULL
	`
	replacements := []interface{}{}
	primarySortKeyColumn := sortRegexp.ReplaceAllString(PrimarySortKey, `_$1`)
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

	if listParams.NextCursor != nil {
		comparisonOp := "<"
		if listParams.SortOrder == service.SortOrderAsc {
			comparisonOp = ">"
		}

		switch listParams.NextCursor.Value() {
		case nil:
			//nolint:gocritic
			if listParams.SortBy == PrimarySortKey {
				query = fmt.Sprintf("%s AND %s %s ?", query, primarySortKeyColumn, comparisonOp)
				replacements = append(replacements, listParams.NextCursor.ID())
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, sortByColumn, primarySortKeyColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.NextCursor.ID(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, primarySortKeyColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.NextCursor.ID(),
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, primarySortKeyColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.NextCursor.Value(),
					listParams.NextCursor.ID(),
					listParams.NextCursor.Value(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, sortByColumn, primarySortKeyColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.NextCursor.Value(),
					listParams.NextCursor.ID(),
					listParams.NextCursor.Value(),
				)
			}
		}
	}

	if listParams.PrevCursor != nil {
		comparisonOp := ">"
		if listParams.SortOrder == service.SortOrderAsc {
			comparisonOp = "<"
		}

		switch listParams.PrevCursor.Value() {
		case nil:
			//nolint:gocritic
			if listParams.SortBy == PrimarySortKey {
				query = fmt.Sprintf("%s AND %s %s ?", query, primarySortKeyColumn, comparisonOp)
				replacements = append(replacements, listParams.PrevCursor.ID())
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, primarySortKeyColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.PrevCursor.ID(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, sortByColumn, primarySortKeyColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.PrevCursor.ID(),
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, sortByColumn, primarySortKeyColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.PrevCursor.Value(),
					listParams.PrevCursor.ID(),
					listParams.PrevCursor.Value(),
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, sortByColumn, comparisonOp, primarySortKeyColumn, comparisonOp, sortByColumn)
				replacements = append(replacements,
					listParams.PrevCursor.Value(),
					listParams.PrevCursor.ID(),
					listParams.PrevCursor.Value(),
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

	if listParams.PrevCursor != nil {
		if listParams.SortBy != PrimarySortKey {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortByColumn, service.SortOrderDesc, invertedNullSortClause, primarySortKeyColumn, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit+1)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortByColumn, service.SortOrderAsc, invertedNullSortClause, primarySortKeyColumn, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit+1)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s %s, %s %s", query, sortByColumn, listParams.SortOrder, nullSortClause, primarySortKeyColumn, listParams.SortOrder)
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, sortByColumn, service.SortOrderDesc, invertedNullSortClause)
				replacements = append(replacements, listParams.Limit+1)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, sortByColumn, service.SortOrderAsc, invertedNullSortClause)
				replacements = append(replacements, listParams.Limit+1)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s %s", query, sortByColumn, listParams.SortOrder, nullSortClause)
		}
	} else {
		if listParams.SortBy != PrimarySortKey {
			query = fmt.Sprintf("%s ORDER BY %s %s %s, %s %s LIMIT ?", query, sortByColumn, listParams.SortOrder, nullSortClause, primarySortKeyColumn, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit+1)
		} else {
			query = fmt.Sprintf("%s ORDER BY %s %s %s LIMIT ?", query, primarySortKeyColumn, listParams.SortOrder, nullSortClause)
			replacements = append(replacements, listParams.Limit+1)
		}
	}

	err := repo.DB.SelectContext(
		ctx,
		&warrants,
		query,
		replacements...,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models, nil, nil, nil
		}
		return nil, nil, nil, errors.Wrap(err, "error listing warrants")
	}

	if len(warrants) == 0 {
		return models, nil, nil, nil
	}

	i := 0
	if listParams.PrevCursor != nil && len(warrants) > listParams.Limit {
		i = 1
	}
	for i < len(warrants) && len(models) < listParams.Limit {
		models = append(models, &warrants[i])
		i++
	}

	firstElem := models[0]
	lastElem := models[len(models)-1]
	var firstValue interface{} = nil
	var lastValue interface{} = nil
	switch listParams.SortBy {
	case PrimarySortKey:
		// do nothing
	case "createdAt":
		firstValue = firstElem.GetCreatedAt()
		lastValue = lastElem.GetCreatedAt()
	}

	prevCursor := service.NewCursor(strconv.FormatInt(firstElem.GetID(), 10), firstValue)
	nextCursor := service.NewCursor(strconv.FormatInt(lastElem.GetID(), 10), lastValue)
	if len(warrants) <= listParams.Limit {
		if listParams.PrevCursor != nil {
			return models, nil, nextCursor, nil
		}

		if listParams.NextCursor != nil {
			return models, prevCursor, nil, nil
		}

		return models, nil, nil, nil
	} else if listParams.PrevCursor == nil && listParams.NextCursor == nil {
		return models, nil, nextCursor, nil
	}

	return models, prevCursor, nextCursor, nil
}
