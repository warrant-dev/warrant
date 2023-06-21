package authz

import (
	"fmt"
	"time"
)

const DefaultSortBy = "objectType"
const DefaultSortByColumn = "typeId"

type ObjectTypeListParamParser struct{}

func (parser ObjectTypeListParamParser) GetDefaultSortBy() string {
	return DefaultSortBy
}

func (parser ObjectTypeListParamParser) GetSupportedSortBys() []string {
	return []string{"createdAt", "objectType"}
}

func (parser ObjectTypeListParamParser) ParseValue(val string, sortBy string) (interface{}, error) {
	switch sortBy {
	case "createdAt":
		afterValue, err := time.Parse(time.RFC3339, val)
		if err != nil || afterValue.Equal(time.Time{}) {
			return nil, fmt.Errorf("must be a valid time in the format %s", time.RFC3339)
		}

		return &afterValue, nil
	case "objectType":
		if val == "" {
			return nil, fmt.Errorf("must not be empty")
		}

		return val, nil
	default:
		return nil, fmt.Errorf("must match type of selected sortBy attribute %s", sortBy)
	}
}
