package predicates

import (
	"context"
	"strings"

	"github.com/ichiban/prolog/engine"
)

// IsList (is_list/1) succeeds if the given term is a list.
//
//	is_list(@Term).
func IsList(t engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	switch t := env.Resolve(t).(type) {
	case engine.Variable:
		return engine.Bool(false)
	case engine.Atom:
		if t != "[]" {
			return engine.Bool(false)
		}
		return k(env)
	case *engine.Compound:
		iter := engine.ListIterator{List: t, Env: env}
		for iter.Next() {
		}
		if iter.Err() != nil {
			return engine.Bool(false)
		}
		return k(env)
	default:
		return engine.Bool(false)
	}
}

// AtomicListConcat (atomic_list_concat/3) succeeds if atom represents the members of list joined by seperator.
// This can be used to join strings by passing a ground list, or used to split strings by passing a ground atom.
//
//	atomic_list_concat(+List, +Seperator, -Atom).
//	atomic_list_concat(-List, +Seperator, +Atom).
func AtomicListConcat(list, seperator, atom engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	sep, ok := seperator.(engine.Atom)
	if !ok {
		return engine.Error(engine.TypeError(engine.ValidTypeAtom, seperator, env))
	}

	switch list := env.Resolve(list).(type) {
	case engine.Variable:
		str, ok := env.Resolve(atom).(engine.Atom)
		if !ok {
			return engine.Error(engine.InstantiationError(env))
		}
		split := strings.Split(string(str), string(sep))
		atoms := make([]engine.Term, len(split))
		for i := 0; i < len(split); i++ {
			atoms[i] = engine.Atom(split[i])
		}
		return engine.Delay(func(context.Context) *engine.Promise {
			return engine.Unify(list, engine.List(atoms...), k, env)
		})
	case *engine.Compound:
		if list.Functor != "." || len(list.Args) != 2 {
			return engine.Error(engine.TypeError(engine.ValidTypeList, list, env))
		}
		var sb strings.Builder
		iter := engine.ListIterator{List: list, Env: env}
		for i := 0; iter.Next(); i++ {
			cur := env.Resolve(iter.Current())
			a, ok := cur.(engine.Atom)
			if !ok {
				return engine.Error(engine.TypeError(engine.ValidTypeAtom, a, env))
			}
			if i > 0 {
				sb.WriteString(string(sep))
			}
			sb.WriteString(string(a))
		}
		str := sb.String()
		return engine.Delay(func(context.Context) *engine.Promise {
			return engine.Unify(atom, engine.Atom(str), k, env)
		})
	case engine.Atom:
		if list != "[]" {
			return engine.Error(engine.TypeError(engine.ValidTypeList, list, env))
		}
		return engine.Delay(func(context.Context) *engine.Promise {
			return engine.Unify(atom, engine.Atom(""), k, env)
		})
	default:
		return engine.Error(engine.TypeError(engine.ValidTypeList, list, env))
	}
}
