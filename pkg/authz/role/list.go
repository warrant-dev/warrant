package authz

import (
	"fmt"
	"time"
)

type RoleListParamParser struct{}

func (parser RoleListParamParser) GetDefaultSortBy() string {
	return "roleId"
}

func (parser RoleListParamParser) GetSupportedSortBys() []string {
	return []string{"createdAt", "roleId", "name"}
}

func (parser RoleListParamParser) ParseValue(val string, sortBy string) (interface{}, error) {
	switch sortBy {
	case "createdAt":
		afterValue, err := time.Parse(time.RFC3339, val)
		if err != nil || afterValue.Equal(time.Time{}) {
			return nil, fmt.Errorf("must be a valid time in the format %s", time.RFC3339)
		}

		return &afterValue, nil
	case "roleId":
		if val == "" {
			return nil, fmt.Errorf("must not be empty")
		}

		return val, nil

	case "name":
		if val == "" {
			return "", nil
		}

		return val, nil
	default:
		return nil, fmt.Errorf("must match type of selected sortBy attribute %s", sortBy)
	}
}
