package authz

import (
	"fmt"
	"time"

	context "github.com/warrant-dev/warrant/pkg/context"
	"github.com/warrant-dev/warrant/pkg/database"
)

type WarrantModel interface {
	GetID() int64
	GetObjectType() string
	GetObjectId() string
	GetRelation() string
	GetSubjectType() string
	GetSubjectId() string
	GetSubjectRelation() database.NullString
	GetContextHash() string
	GetContext() []context.Context
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() database.NullTime
	ToWarrantSpec() *WarrantSpec
}

// Warrant model
type Warrant struct {
	ID              int64                  `mysql:"id" postgres:"id"`
	ObjectType      string                 `mysql:"objectType" postgres:"object_type"`
	ObjectId        string                 `mysql:"objectId" postgres:"object_id"`
	Relation        string                 `mysql:"relation" postgres:"relation"`
	SubjectType     string                 `mysql:"subjectType" postgres:"subject_type"`
	SubjectId       string                 `mysql:"subjectId" postgres:"subject_id"`
	SubjectRelation database.NullString    `mysql:"subjectRelation" postgres:"subject_relation"`
	ContextHash     string                 `mysql:"contextHash" postgres:"context_hash"`
	Context         []context.ContextModel `mysql:"context" postgres:"context"`
	CreatedAt       time.Time              `mysql:"createdAt" postgres:"created_at"`
	UpdatedAt       time.Time              `mysql:"updatedAt" postgres:"updated_at"`
	DeletedAt       database.NullTime      `mysql:"deletedAt" postgres:"deleted_at"`
}

func (warrant Warrant) GetID() int64 {
	return warrant.ID
}

func (warrant Warrant) GetObjectType() string {
	return warrant.ObjectType
}

func (warrant Warrant) GetObjectId() string {
	return warrant.ObjectId
}

func (warrant Warrant) GetRelation() string {
	return warrant.Relation
}

func (warrant Warrant) GetSubjectType() string {
	return warrant.SubjectType
}

func (warrant Warrant) GetSubjectId() string {
	return warrant.SubjectId
}

func (warrant Warrant) GetSubjectRelation() database.NullString {
	return warrant.SubjectRelation
}

func (warrant Warrant) GetContextHash() string {
	return warrant.ContextHash
}

func (warrant Warrant) GetContext() []context.ContextModel {
	return warrant.Context
}

func (warrant Warrant) GetCreatedAt() time.Time {
	return warrant.CreatedAt
}

func (warrant Warrant) GetUpdatedAt() time.Time {
	return warrant.UpdatedAt
}

func (warrant Warrant) GetDeletedAt() database.NullTime {
	return warrant.DeletedAt
}

func (warrant Warrant) ToWarrantSpec() *WarrantSpec {
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
			contextSetSpec[context.GetName()] = context.GetValue()
		}

		warrantSpec.Context = contextSetSpec
	}

	return &warrantSpec
}

func (warrant Warrant) String() string {
	str := fmt.Sprintf("%s:%s#%s@%s:%s", warrant.ObjectType, warrant.ObjectId, warrant.Relation, warrant.SubjectType, warrant.SubjectId)

	if warrant.SubjectRelation.String != "" {
		str = fmt.Sprintf("%s#%s", str, warrant.SubjectRelation.String)
	}

	if warrant.ContextHash != "" {
		str = fmt.Sprintf("%s[%s]", str, warrant.ContextHash)
	}

	return str
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
		SubjectType:     warrantSpec.Subject.ObjectType,
		SubjectId:       warrantSpec.Subject.ObjectId,
		SubjectRelation: database.StringToNullString(&warrantSpec.Subject.Relation),
		ContextHash:     warrantSpec.Context.ToHash(),
	}, nil
}
