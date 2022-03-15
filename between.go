package predicates

import (
	"context"
	"math"

	"github.com/ichiban/prolog/engine"

	"github.com/guregu/predicates/internal"
)

// Between (between/3) succeeds iff low <= value <= high.
// If value is a variable, it is unified with successive integers from low to high.
// between(+Low, +High, ?Value).
// See: https://www.swi-prolog.org/pldoc/doc_for?object=between/3
func Between(low, high, value engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	var lo, hi engine.Integer

	switch l := env.Resolve(low).(type) {
	case engine.Integer:
		lo = l
	case engine.Variable:
		return engine.Error(engine.ErrInstantiation)
	default:
		return engine.Error(internal.TypeErrorInteger(low))
	}

	switch h := env.Resolve(high).(type) {
	case engine.Integer:
		hi = h
	case engine.Variable:
		return engine.Error(engine.ErrInstantiation)
	case engine.Atom:
		switch h {
		case "infinite", "inf":
			hi = math.MaxInt64 // is this dumb?
		default:
			return engine.Error(internal.TypeErrorInteger(high))
		}
	default:
		return engine.Error(internal.TypeErrorInteger(high))
	}

	switch x := env.Resolve(value).(type) {
	case engine.Integer:
		if x >= lo && x <= hi {
			return k(env)
		}
		return engine.Bool(false)
	case engine.Variable:
		return engine.Delay(func(context.Context) *engine.Promise {
			i := lo - 1
			return engine.Repeat(func(context.Context) *engine.Promise {
				i++
				switch {
				case i-1 > i:
					return engine.Error(internal.EvaluationErrorIntOverflow())
				case i > hi:
					return engine.Bool(true)
				}
				return engine.Unify(value, i, k, env)
			})
		})
	}
	return engine.Error(internal.TypeErrorInteger(value))
}
