package predicates

import (
	"context"
	"strings"

	"github.com/ichiban/prolog/engine"
)

// DowncaseAtom (downcase_atom/2) converts atom into its lowercase equivalent.
//
//	downcase_atom(+Atom, -LowerCase).
func DowncaseAtom(atom, lowercase engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	var a engine.Atom
	switch atom := env.Resolve(atom).(type) {
	case engine.Variable:
		return engine.Error(engine.InstantiationError(env))
	case engine.Atom:
		a = atom
	default:
		return engine.Error(engine.TypeError(engine.ValidTypeAtom, atom, env))
	}

	switch low := env.Resolve(lowercase).(type) {
	case engine.Atom, engine.Variable:
		transformed := engine.Atom(strings.ToLower(string(a)))
		return engine.Delay(func(context.Context) *engine.Promise {
			return engine.Unify(low, transformed, k, env)
		})
	default:
		return engine.Error(engine.TypeError(engine.ValidTypeAtom, low, env))
	}
}

// UpcaseAtom (upcase_atom/2) converts atom into its uppercase equivalent.
//
//	upcase_atom(+Atom, -UpperCase).
func UpcaseAtom(atom, uppercase engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	var a engine.Atom
	switch atom := env.Resolve(atom).(type) {
	case engine.Variable:
		return engine.Error(engine.InstantiationError(env))
	case engine.Atom:
		a = atom
	default:
		return engine.Error(engine.TypeError(engine.ValidTypeAtom, atom, env))
	}

	switch low := env.Resolve(uppercase).(type) {
	case engine.Atom, engine.Variable:
		transformed := engine.Atom(strings.ToUpper(string(a)))
		return engine.Delay(func(context.Context) *engine.Promise {
			return engine.Unify(low, transformed, k, env)
		})
	default:
		return engine.Error(engine.TypeError(engine.ValidTypeAtom, low, env))
	}
}
