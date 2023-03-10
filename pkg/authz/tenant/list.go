package tenant

import (
	"fmt"
	"time"
)

type TenantListParamParser struct{}

func (parser TenantListParamParser) GetDefaultSortBy() string {
	return "tenantId"
}

func (parser TenantListParamParser) GetSupportedSortBys() []string {
	return []string{"createdAt", "name", "tenantId"}
}

func (parser TenantListParamParser) ParseValue(val string, sortBy string) (interface{}, error) {
	switch sortBy {
	case "createdAt":
		afterValue, err := time.Parse(time.RFC3339, val)
		if err != nil || afterValue.Equal(time.Time{}) {
			return nil, fmt.Errorf("must be a valid time in the format %s", time.RFC3339)
		}

		return afterValue, nil
	case "name":
		if val == "" {
			return "", nil
		}

		return val, nil
	case "tenantId":
		if val == "" {
			return nil, fmt.Errorf("must not be empty")
		}

		return val, nil
	default:
		return nil, fmt.Errorf("must match type of selected sortBy attribute %s", sortBy)
	}
}
