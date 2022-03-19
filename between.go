package predicates

import (
	"context"

	"github.com/ichiban/prolog/engine"

	"github.com/guregu/predicates/internal"
)

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

	if low > high {
		return engine.Bool(false)
	}

	switch value := env.Resolve(value).(type) {
	case engine.Integer:
		if value >= low && value <= high {
			return k(env)
		}
		return engine.Bool(false)
	case engine.Variable:
		return engine.Delay(func(context.Context) *engine.Promise {
			return engine.Unify(value, low, k, env)
		}, func(context.Context) *engine.Promise {
			return Between(low+1, upper, value, k, env)
		})
	default:
		return engine.Error(internal.TypeErrorInteger(value))
	}
}
