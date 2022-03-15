package predicates

import (
	"context"

	"github.com/ichiban/prolog/engine"

	"github.com/guregu/predicates/internal"
)

// Between (between/3) is true when lower, upper, and value are all integers, and lower <= value <= upper.
// If value is a variable, it is unified with successive integers from lower to upper.
// between(+Lower, +Upper, -Value).
func Between(lower, upper, value engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	var low, high engine.Integer

	switch lower := env.Resolve(lower).(type) {
	case engine.Integer:
		low = lower
	case engine.Variable:
		return engine.Error(engine.ErrInstantiation)
	default:
		return engine.Error(internal.TypeErrorInteger(lower))
	}

	switch upper := env.Resolve(upper).(type) {
	case engine.Integer:
		high = upper
	case engine.Variable:
		return engine.Error(engine.ErrInstantiation)
	default:
		return engine.Error(internal.TypeErrorInteger(upper))
	}

	switch value := env.Resolve(value).(type) {
	case engine.Integer:
		if value >= low && value <= high {
			return k(env)
		}
		return engine.Bool(false)
	case engine.Variable:
		return engine.Delay(func(context.Context) *engine.Promise {
			i := low - 1
			return engine.Repeat(func(context.Context) *engine.Promise {
				i++
				switch {
				case i-1 > i:
					return engine.Error(internal.EvaluationErrorIntOverflow())
				case i > high:
					return engine.Bool(true)
				}
				return engine.Unify(value, i, k, env)
			})
		})
	}
	return engine.Error(internal.TypeErrorInteger(value))
}
