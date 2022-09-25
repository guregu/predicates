// Package chars contains convenience functions for working with Prolog strings (list of characters).
package chars

import (
	"strings"
	"unicode/utf8"

	"github.com/ichiban/prolog/engine"
)

// Chars is a list of characters in native Go representation.
type Chars interface {
	~string | ~[]rune
}

// List returns a new list of Prolog strings constructed from strs.
func List[T Chars](strs ...T) engine.Term {
	ts := make([]engine.Term, len(strs))
	for i, str := range strs {
		ts[i] = String(str)
	}
	return engine.List(ts...)
}

// String returns a new Prolog string constructed from str.
func String[T Chars](str T) engine.Term {
	rs := []rune(str)
	ts := make([]engine.Term, len(rs))
	for i, r := range rs {
		ts[i] = engine.Atom(r)
	}
	return engine.List(ts...)
}

// Value resolves str, which must be a Prolog string, and returns a Go string or an error.
func Value[T Chars](str engine.Term, env *engine.Env) (T, error) {
	list := env.Resolve(str)
	var empty T
	var sb strings.Builder
	iter := engine.ListIterator{List: list, Env: env}
	for iter.Next() {
		elem := env.Resolve(iter.Current())
		switch x := elem.(type) {
		case engine.Variable:
			return empty, engine.InstantiationError(env)
		case engine.Atom:
			char, size := utf8.DecodeRuneInString(string(x))
			if char == utf8.RuneError ||
				size == 0 ||
				size != len(x) {
				// not a list of single characters
				return empty, engine.TypeError(engine.ValidTypeCharacter, x, env)
			}
			sb.WriteRune(char)
		default:
			return empty, engine.TypeError(engine.ValidTypeCharacter, list, env)
		}
	}
	return T(sb.String()), iter.Err()
}

// Values resolves list, which must be a list of strings, and converts all its members into Go strings.
func Values[T Chars](list engine.Term, env *engine.Env) ([]T, error) {
	list = env.Resolve(list)
	var vs []T
	iter := engine.ListIterator{List: list, Env: env}
	for iter.Next() {
		elem := env.Resolve(iter.Current())
		switch x := elem.(type) {
		case engine.Variable:
			return nil, engine.InstantiationError(env)
		case engine.Compound:
			v, err := Value[T](x, env)
			if err != nil {
				return nil, err
			}
			vs = append(vs, v)
		default:
			return nil, engine.TypeError(engine.ValidTypeCharacter, list, env)
		}
	}
	return vs, iter.Err()
}
