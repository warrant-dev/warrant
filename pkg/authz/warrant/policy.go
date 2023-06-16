package authz

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/antonmedv/expr"
	"github.com/pkg/errors"
)

type Policy string

type PolicyContext map[string]interface{}

func (pc PolicyContext) String() string {
	if len(pc) == 0 {
		return ""
	}

	contextKeys := make([]string, 0)
	for key := range pc {
		contextKeys = append(contextKeys, key)
	}
	sort.Strings(contextKeys)

	keyValuePairs := make([]string, 0)
	for _, key := range contextKeys {
		keyValuePairs = append(keyValuePairs, fmt.Sprintf("%s=%v", key, pc[key]))
	}

	return fmt.Sprintf("[%s]", strings.Join(keyValuePairs, " "))
}

func defaultExprOptions(ctx PolicyContext) []expr.Option {
	opts := []expr.Option{
		expr.AllowUndefinedVariables(),
		expr.AsBool(),
	}

	if ctx != nil {
		opts = append(opts, expr.Env(ctx))
	}

	opts = append(opts, expr.Function(
		"expiresIn",
		func(params ...interface{}) (interface{}, error) {
			durationStr := params[0].(string)
			duration, err := time.ParseDuration(durationStr)
			if err != nil {
				return false, fmt.Errorf("invalid duration string %s", durationStr)
			}

			warrantCreatedAt := ctx["warrant"].(*Warrant).CreatedAt
			return bool(time.Now().Before(warrantCreatedAt.Add(duration))), nil
		},
		new(func(string) bool),
	))

	return opts
}

func (policy Policy) Validate() error {
	_, err := expr.Compile(string(policy), defaultExprOptions(nil)...)
	if err != nil {
		return errors.Wrapf(err, "error validating policy '%s'", policy)
	}

	return nil
}

func (policy Policy) Eval(ctx PolicyContext) (bool, error) {
	program, err := expr.Compile(string(policy), defaultExprOptions(ctx)...)
	if err != nil {
		return false, errors.Wrapf(err, "error compiling policy '%s'", policy)
	}

	match, err := expr.Run(program, ctx)
	if err != nil {
		return false, errors.Wrapf(err, "error evaluating policy '%s'", policy)
	}

	return match.(bool), nil
}

func (policy Policy) Hash() string {
	if policy == "" {
		return ""
	}

	hash := sha256.Sum256([]byte(policy))
	return hex.EncodeToString(hash[:])
}
