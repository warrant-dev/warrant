package authz

import (
	"time"

	context "github.com/warrant-dev/warrant/pkg/context"
	"github.com/warrant-dev/warrant/pkg/database"
)

// Warrant model
type Warrant struct {
	ID              int64               `db:"id"`
	ObjectType      string              `db:"objectType"`
	ObjectId        string              `db:"objectId"`
	Relation        string              `db:"relation"`
	Subject         string              `db:"subject"`
	SubjectType     string              `db:"subjectType"`
	SubjectId       string              `db:"subjectId"`
	SubjectRelation database.NullString `db:"subjectRelation"`
	ContextHash     string              `db:"contextHash"`
	Context         []context.Context   `db:"context"`
	CreatedAt       time.Time           `db:"createdAt"`
	UpdatedAt       time.Time           `db:"updatedAt"`
	DeletedAt       database.NullTime   `db:"deletedAt"`
}

func (warrant *Warrant) ToWarrantSpec() *WarrantSpec {
	warrantSpec := WarrantSpec{
		ObjectType: warrant.ObjectType,
		ObjectId:   warrant.ObjectId,
		Relation:   warrant.Relation,
		Subject: &SubjectSpec{
			ObjectType: warrant.SubjectType,
			ObjectId:   warrant.SubjectId,
			Relation:   warrant.SubjectRelation.String,
		},
		CreatedAt: warrant.CreatedAt,
	}

	if len(warrant.Context) > 0 {
		contextSetSpec := make(context.ContextSetSpec, len(warrant.Context))
		for _, context := range warrant.Context {
			contextSetSpec[context.Name] = context.Value
		}

		warrantSpec.Context = contextSetSpec
	}

	return &warrantSpec
}

func StringToWarrant(warrantString string) (*Warrant, error) {
	warrantSpec, err := StringToWarrantSpec(warrantString)
	if err != nil {
		return nil, err
	}

	return &Warrant{
		ObjectType:      warrantSpec.ObjectType,
		ObjectId:        warrantSpec.ObjectId,
		Relation:        warrantSpec.Relation,
		Subject:         warrantSpec.Subject.String(),
		SubjectType:     warrantSpec.Subject.ObjectType,
		SubjectId:       warrantSpec.Subject.ObjectId,
		SubjectRelation: database.StringToNullString(&warrantSpec.Subject.Relation),
		ContextHash:     warrantSpec.Context.ToHash(),
	}, nil
}
