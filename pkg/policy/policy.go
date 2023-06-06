package policy

import (
	"github.com/antonmedv/expr"
	"github.com/pkg/errors"
)

func defaultExprOptions(ctx ContextSpec) []expr.Option {
	opts := []expr.Option{
		expr.AllowUndefinedVariables(),
		expr.AsBool(),
	}

	if ctx != nil {
		opts = append(opts, expr.Env(ctx))
	}

	return opts
}

func Validate(policy string) error {
	_, err := expr.Compile(policy, defaultExprOptions(nil)...)
	if err != nil {
		return errors.Wrapf(err, "error validating policy %s", policy)
	}

	return nil
}

func Eval(policy string, ctx ContextSpec) (bool, error) {
	program, err := expr.Compile(policy, defaultExprOptions(ctx)...)
	if err != nil {
		return false, errors.Wrapf(err, "error compiling policy %s", policy)
	}

	match, err := expr.Run(program, ctx)
	if err != nil {
		return false, errors.Wrapf(err, "error evaluating policy %s", policy)
	}

	return match.(bool), nil
}
