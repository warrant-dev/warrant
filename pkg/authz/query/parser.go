// Copyright 2024 WorkOS, Inc.
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
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/pkg/errors"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/pkg/service"
)

type selectClause struct {
	Explicit               bool     `parser:"@Explicit?"`
	ObjectTypesOrRelations []string `parser:"(@Wildcard | (@TypeOrRelation (Comma @TypeOrRelation)*))?"`
	SubjectTypes           []string `parser:"(OfType (@Wildcard | (@TypeOrRelation (Comma @TypeOrRelation)*)))?"`
}

type forClause struct {
	Object string `parser:"@Resource"`
}

type whereClause struct {
	Subject   string   `parser:"@Resource Is"`
	Relations []string `parser:"(@Wildcard | (@TypeOrRelation (Comma @TypeOrRelation)*))"`
}

type ast struct {
	SelectClause *selectClause `parser:"Select @@"`
	ForClause    *forClause    `parser:"(For @@)?"`
	WhereClause  *whereClause  `parser:"(Where @@)?"`
}

var participleLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Select", Pattern: `(?i)\bselect\b`},
	{Name: "Explicit", Pattern: `(?i)\bexplicit\b`},
	{Name: "Where", Pattern: `(?i)\bwhere\b`},
	{Name: "Is", Pattern: `(?i)\bis\b`},
	{Name: "For", Pattern: `(?i)\bfor\b`},
	{Name: "OfType", Pattern: `(?i)\bof type\b`},
	{Name: "Resource", Pattern: `[a-zA-Z0-9_\-]+:[a-zA-Z0-9_\-\.@\|:]+`},
	{Name: "TypeOrRelation", Pattern: `[a-zA-Z0-9_\-]+`},
	{Name: "Wildcard", Pattern: `\*`},
	{Name: "Comma", Pattern: `,`},
	{Name: "EOL", Pattern: `[\n\r]+`},
	{Name: "whitespace", Pattern: `[ \t\n\r]+`},
})

type parser struct {
	*participle.Parser[ast]
}

func newParser() (*parser, error) {
	participleParser, err := participle.Build[ast](
		participle.Lexer(participleLexer),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error generating query parser")
	}

	return &parser{participleParser}, nil
}

func (parser parser) Parse(query string) (*ast, error) {
	ast, err := parser.Parser.ParseString("", query)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing query")
	}

	return ast, nil
}

func NewQueryFromString(queryString string) (Query, error) {
	var query Query

	queryParser, err := newParser()
	if err != nil {
		return Query{}, errors.Wrap(err, "error creating query from string")
	}

	ast, err := queryParser.Parse(queryString)
	if err != nil {
		return Query{}, service.NewInvalidParameterError("q", err.Error())
	}

	if ast.SelectClause == nil {
		return Query{}, service.NewInvalidParameterError("q", "must contain a 'select' clause")
	}

	if ast.SelectClause.ObjectTypesOrRelations == nil && ast.SelectClause.SubjectTypes == nil {
		return Query{}, service.NewInvalidParameterError("q", "incomplete 'select' clause")
	}

	if ast.ForClause != nil && ast.WhereClause != nil {
		return Query{}, service.NewInvalidParameterError("q", "cannot contain both a 'for' clause and a 'where' clause")
	}

	query.Expand = !ast.SelectClause.Explicit

	if ast.SelectClause.SubjectTypes != nil { // Querying for subjects
		if len(ast.SelectClause.SubjectTypes) == 0 {
			return Query{}, service.NewInvalidParameterError("q", "must contain one or more types of subjects to select")
		}

		if ast.SelectClause.ObjectTypesOrRelations == nil || len(ast.SelectClause.ObjectTypesOrRelations) == 0 {
			return Query{}, service.NewInvalidParameterError("q", "must select one or more relations for subjects to match on the object")
		}

		if ast.WhereClause != nil {
			return Query{}, service.NewInvalidParameterError("q", "cannot contain a 'where' clause when selecting subjects")
		}

		query.SelectSubjects = &SelectSubjects{
			Relations:    ast.SelectClause.ObjectTypesOrRelations,
			SubjectTypes: ast.SelectClause.SubjectTypes,
		}

		if ast.ForClause == nil {
			return Query{}, service.NewInvalidParameterError("q", "must contain a 'for' clause")
		}

		objectType, objectId, colonFound := strings.Cut(ast.ForClause.Object, ":")
		if !colonFound {
			return Query{}, service.NewInvalidParameterError("q", "'for' clause contains invalid object")
		}

		query.SelectSubjects.ForObject = &Resource{
			Type: objectType,
			Id:   objectId,
		}
	} else { // Querying for objects
		if ast.SelectClause.ObjectTypesOrRelations == nil || len(ast.SelectClause.ObjectTypesOrRelations) == 0 {
			return Query{}, service.NewInvalidParameterError("q", "must contain one or more types of objects to select")
		}

		if ast.ForClause != nil {
			return Query{}, service.NewInvalidParameterError("q", "cannot contain a 'for' clause when selecting objects")
		}

		query.SelectObjects = &SelectObjects{
			ObjectTypes: ast.SelectClause.ObjectTypesOrRelations,
			Relations:   []string{warrant.Wildcard},
		}

		if ast.WhereClause == nil {
			return Query{}, service.NewInvalidParameterError("q", "must contain a 'where' clause")
		}

		if ast.WhereClause.Relations == nil || len(ast.WhereClause.Relations) == 0 {
			return Query{}, service.NewInvalidParameterError("q", "must contain one or more relations the subject must have on matching objects")
		}

		subjectType, subjectId, colonFound := strings.Cut(ast.WhereClause.Subject, ":")
		if !colonFound {
			return Query{}, service.NewInvalidParameterError("q", "'where' clause contains invalid subject")
		}

		query.SelectObjects.Relations = ast.WhereClause.Relations
		query.SelectObjects.WhereSubject = &Resource{
			Type: subjectType,
			Id:   subjectId,
		}
	}

	return query, nil
}
