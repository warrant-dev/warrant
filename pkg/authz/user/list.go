package authz

import (
	"fmt"
	"net/mail"
	"time"
)

const defaultSortBy = "userId"

type UserListParamParser struct{}

func (parser UserListParamParser) GetDefaultSortBy() string {
	return defaultSortBy
}

func (parser UserListParamParser) GetSupportedSortBys() []string {
	return []string{"userId", "createdAt", "email"}
}

func (parser UserListParamParser) ParseValue(val string, sortBy string) (interface{}, error) {
	switch sortBy {
	case "createdAt":
		afterValue, err := time.Parse(time.RFC3339, val)
		if err != nil || afterValue.Equal(time.Time{}) {
			return nil, fmt.Errorf("must be a valid time in the format %s", time.RFC3339)
		}

		return &afterValue, nil
	case "email":
		if val == "" {
			return "", nil
		}

		afterValue, err := mail.ParseAddress(val)
		if err != nil {
			return nil, fmt.Errorf("must be a valid email")
		}

		return afterValue.Address, nil
	case "userId":
		if val == "" {
			return nil, fmt.Errorf("must not be empty")
		}

		return val, nil
	default:
		return nil, fmt.Errorf("must match type of selected sortBy attribute %s", sortBy)
	}
}
