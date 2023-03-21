package authz

import (
	"encoding/json"

	"github.com/warrant-dev/warrant/pkg/service"
)

const (
	ObjectTypeFeature     = "feature"
	ObjectTypePermission  = "permission"
	ObjectTypePricingTier = "pricing-tier"
	ObjectTypeRole        = "role"
	ObjectTypeTenant      = "tenant"
	ObjectTypeUser        = "user"

	RelationAdmin   = "admin"
	RelationManager = "manager"
	RelationMember  = "member"
	RelationParent  = "parent"
	RelationOwner   = "owner"
	RelationEditor  = "editor"
	RelationViewer  = "viewer"

	InheritIfAllOf  = "allOf"
	InheritIfAnyOf  = "anyOf"
	InheritIfNoneOf = "noneOf"
)

type ObjectTypeSpec struct {
	Type      string                  `json:"type" validate:"required,valid_object_type"`
	Source    *Source                 `json:"source,omitempty"`
	Relations map[string]RelationRule `json:"relations" validate:"required,min=1,dive"` // NOTE: map key = name of relation
}

func (spec ObjectTypeSpec) ToObjectType() (*ObjectType, error) {
	definition, err := json.Marshal(spec)
	if err != nil {
		return nil, service.NewInvalidRequestError("invalid request body")
	}

	return &ObjectType{
		TypeId:     spec.Type,
		Definition: string(definition),
	}, nil
}

type Source struct {
	DatabaseType string           `json:"dbType" validate:"required"`
	DatabaseName string           `json:"dbName" validate:"required"`
	Table        string           `json:"table" validate:"required"`
	PrimaryKey   []string         `json:"primaryKey" validate:"min=1"`
	ForeignKeys  []ForeignKeySpec `json:"foreignKeys,omitempty"`
}

type ForeignKeySpec struct {
	Column   string `json:"column" validate:"required"`
	Relation string `json:"relation" validate:"required,valid_relation"`
	Type     string `json:"type" validate:"required,valid_object_type"`
	Subject  string `json:"subject" validate:"required"`
}

// RelationRule type represents the rule or set of rules that imply a particular relation if met
type RelationRule struct {
	InheritIf    string         `json:"inheritIf,omitempty" validate:"required_with=Rules OfType WithRelation,valid_inheritif"`
	Rules        []RelationRule `json:"rules,omitempty" validate:"required_if_oneof=InheritIf anyOf allOf noneOf,omitempty,min=1,dive"` // Required if InheritIf is "anyOf", "allOf", or "noneOf", empty otherwise
	OfType       string         `json:"ofType,omitempty" validate:"required_with=WithRelation,valid_relation"`
	WithRelation string         `json:"withRelation,omitempty" validate:"required_with=OfType,valid_relation"`
}

var UserObjectTypeSpec = ObjectTypeSpec{
	Type: ObjectTypeUser,
	Relations: map[string]RelationRule{
		RelationParent: {
			InheritIf:    RelationParent,
			OfType:       ObjectTypeUser,
			WithRelation: RelationParent,
		},
	},
}

var TenantObjectTypeSpec = ObjectTypeSpec{
	Type: ObjectTypeTenant,
	Relations: map[string]RelationRule{
		RelationAdmin: {},
		RelationManager: {
			InheritIf: RelationAdmin,
		},
		RelationMember: {
			InheritIf: RelationManager,
		},
	},
}

var RoleObjectTypeSpec = ObjectTypeSpec{
	Type: ObjectTypeRole,
	Relations: map[string]RelationRule{
		RelationOwner: {},

		// editor of Role if owner of this Role
		RelationEditor: {
			InheritIf: RelationOwner,
		},

		// viewer of Role if editor of this Role
		RelationViewer: {
			InheritIf: RelationEditor,
		},

		// member of Role if member of Role that is a member of this Role
		RelationMember: {
			InheritIf:    RelationMember,
			OfType:       ObjectTypeRole,
			WithRelation: RelationMember,
		},
	},
}

var PermissionObjectTypeSpec = ObjectTypeSpec{
	Type: ObjectTypePermission,
	Relations: map[string]RelationRule{
		RelationOwner: {},

		// editor of Permission if owner of this Permission
		RelationEditor: {
			InheritIf: RelationOwner,
		},

		// viewer of Permission if editor of this Permission
		RelationViewer: {
			InheritIf: RelationEditor,
		},

		// member of Permission if:
		// 1. member of Permission that is a member of this Permission
		// OR
		// 2. member of Role that is a member of this Permission
		RelationMember: {
			InheritIf: InheritIfAnyOf,
			Rules: []RelationRule{
				{
					InheritIf:    RelationMember,
					OfType:       ObjectTypePermission,
					WithRelation: RelationMember,
				},
				{
					InheritIf:    RelationMember,
					OfType:       ObjectTypeRole,
					WithRelation: RelationMember,
				},
			},
		},
	},
}

var PricingTierObjectTypeSpec = ObjectTypeSpec{
	Type: ObjectTypePricingTier,
	Relations: map[string]RelationRule{
		// member of pricing-tier if member of PricingTier that is a member of this PricingTier
		RelationMember: {
			InheritIf:    RelationMember,
			OfType:       ObjectTypePricingTier,
			WithRelation: RelationMember,
		},
	},
}

var FeatureObjectTypeSpec = ObjectTypeSpec{
	Type: ObjectTypeFeature,
	Relations: map[string]RelationRule{
		// member of feature if member of a member
		RelationMember: {
			InheritIf: InheritIfAnyOf,
			Rules: []RelationRule{
				{
					InheritIf:    RelationMember,
					OfType:       ObjectTypeFeature,
					WithRelation: RelationMember,
				},
				{
					InheritIf:    RelationMember,
					OfType:       ObjectTypePricingTier,
					WithRelation: RelationMember,
				},
			},
		},
	},
}
