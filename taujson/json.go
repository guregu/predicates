// Package taujson provides JSON-related Prolog predicates compatible with Tau Prolog's library(js).
// These predicates use an opaque native object in their first argument.
//
// See: http://tau-prolog.org/documentation#js
package taujson

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ichiban/prolog"
	"github.com/ichiban/prolog/engine"

	"github.com/guregu/predicates/chars"
	"github.com/guregu/predicates/internal"
)

// Register registers this package's predicates to the given interpreter with default names.
func Register(p *prolog.Interpreter) {
	if err := p.Exec(`
		:- built_in(json_atom/2).
		:- built_in(json_prolog/2).
	`); err != nil {
		panic(err)
	}
	p.Register2("json_atom", JSONAtom)
	p.Register2("json_prolog", JSONProlog)
}

// JSONAtom (json_atom/2) succeeds if JS is a native JSON object that represents the JSON-encoded atom.
// This is intended to be compatible with Tau Prolog's library(js).
//
//	json_atom(-JS, +Atom).
//	json_atom(+JS, -Atom).
func JSONAtom(js, atom engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	var raw engine.Compound

	switch js := env.Resolve(js).(type) {
	case engine.Compound:
		raw = js
	case engine.Variable:
	default:
		return engine.Error(engine.TypeError(engine.ValidTypeList, js, env))
	}

	return engine.Delay(func(context.Context) *engine.Promise {
		str, err := jsonOf(raw, env)
		if err != nil {
			return engine.Error(err)
		}

		switch atom := env.Resolve(atom).(type) {
		case engine.Variable:
			if raw == nil {
				return engine.Error(engine.InstantiationError(env))
			}
			t := engine.Atom(str)
			return engine.Unify(atom, t, k, env)
		case engine.Atom:
			if raw == nil {
				t := engine.CharList(string(atom))
				return engine.Unify(js, t, k, env)
			}
			if engine.Atom(str) != atom {
				return engine.Bool(false)
			}
			return k(env)
		default:
			return engine.Error(engine.TypeError(engine.ValidTypeAtom, atom, env))
		}
	})
}

// JSONProlog (json_prolog/2) succeeds if JS is a native JSON object that represents List.
// This is intended to be compatible with Tau Prolog's library(json).
//
//	json_prolog(-JS, +List).
//	json_prolog(+JS, -List).
func JSONProlog(js, value engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	var raw engine.Compound
	switch js := env.Resolve(js).(type) {
	case engine.Compound:
		raw = js
	case engine.Variable:
	default:
		return engine.Error(engine.TypeError(engine.ValidTypeList, js, env))
	}

	return engine.Delay(func(context.Context) *engine.Promise {
		value := env.Resolve(value)

		str, err := jsonOf(raw, env)
		if err != nil {
			return engine.Error(err)
		}

		switch value := value.(type) {
		case engine.Variable:
			if raw == nil {
				return engine.Error(engine.InstantiationError(env))
			}
			t, err := json2prolog([]byte(str))
			if err != nil {
				return engine.Error(err)
			}
			return engine.Unify(value, t, k, env)
		case engine.Atom:
			if value != "[]" {
				return engine.Error(engine.TypeError(engine.ValidTypeList, value, env))
			}
		case engine.Compound:
			// Tau only accepts lists?
			if value.Functor() != "." || value.Arity() != 2 {
				return engine.Error(engine.TypeError(engine.ValidTypeList, value, env))
			}
		default:
			return engine.Error(engine.TypeError(engine.ValidTypeList, value, env))
		}

		enc, err := prolog2json(value, env)
		if err != nil {
			return engine.Error(err)
		}
		jsTerm := engine.CharList(string(enc))
		return engine.Unify(js, jsTerm, k, env)
	})
}

// type Term = engine.Compound

// Term is a native representation of JSON.
// This is intended to match behavior with Tau Prolog.
// Proper JSON predicates coming soon! 😇
// type Term json.RawMessage

// // Unify unifies the native JS object with t.
// func (js *Term) Unify(t engine.Term, occursCheck bool, env *engine.Env) (*engine.Env, bool) {
// 	switch t := env.Resolve(t).(type) {
// 	case *Term:
// 		return env, bytes.Equal(*js, *t)
// 	case engine.Variable:
// 		return t.Unify(js, occursCheck, env)
// 	default:
// 		return env, false
// 	}
// }

// // WriteTerm writes the Stream to the io.Writer.
// func (js *Term) WriteTerm(w io.Writer, opts *engine.WriteOptions, env *engine.Env) error {
// 	c := engine.Compound{
// 		Functor: "$json",
// 		// Args: []engine.Term{
// 		// 	engine.Atom(string(*js)),
// 		// },
// 	}
// 	return c.WriteTerm(w, opts, env)
// }

// // Compare compares the native JS object to another term.
// func (js *Term) Compare(t engine.Term, env *engine.Env) int64 {
// 	switch t := env.Resolve(t).(type) {
// 	case *Term:
// 		if js == t {
// 			return 0
// 		}
// 		return 1
// 	default:
// 		return 1
// 	}
// }

func prolog2json(t engine.Term, env *engine.Env) ([]byte, error) {
	switch t := env.Resolve(t).(type) {
	case engine.Variable:
		return nil, engine.InstantiationError(env)
	case engine.Atom:
		if t == "[]" {
			return []byte("[]"), nil
		}
		return json.Marshal(string(t))
	case engine.Integer:
		return json.Marshal(int64(t))
	case engine.Float:
		return json.Marshal(float64(t))
	case engine.Compound:
		if internal.IsMap(t, env) {
			m := make(map[string]json.RawMessage)
			iter := engine.ListIterator{List: t, Env: env}
			for iter.Next() {
				cur := env.Resolve(iter.Current())
				cmp := cur.(engine.Compound)
				k := string(env.Resolve(cmp.Arg(0)).(engine.Atom))
				v, err := prolog2json(env.Resolve(cmp.Arg(1)), env)
				if err != nil {
					return nil, err
				}
				m[k] = json.RawMessage(v)
			}
			if err := iter.Err(); err != nil {
				return nil, err
			}
			return json.Marshal(m)
		}

		if t.Functor() == "." && t.Arity() == 2 {
			list := make([]json.RawMessage, 0)
			iter := engine.ListIterator{List: t, Env: env}
			for iter.Next() {
				cur := env.Resolve(iter.Current())
				v, err := prolog2json(cur, env)
				if err != nil {
					return nil, err
				}
				list = append(list, json.RawMessage(v))
			}
			if err := iter.Err(); err != nil {
				return nil, err
			}
			return json.Marshal(list)
		}

		var sb strings.Builder
		if err := engine.WriteTerm(&sb, t, &engine.WriteOptions{}, env); err != nil {
			return nil, err
		}
		return json.Marshal(sb.String())
	}
	return nil, nil
}

func json2prolog(raw []byte) (engine.Term, error) {
	var v any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}

	t := iface2prolog(v)
	return t, nil
}

func iface2prolog(v any) engine.Term {
	var list []engine.Term

	switch x := v.(type) {
	case map[string]any:
		for k, v := range x {
			t := iface2prolog(v)
			list = append(list, engine.Atom("-").Apply(engine.Atom(k), t))
		}
		return engine.List(list...)
	case []any:
		for _, member := range x {
			t := iface2prolog(member)
			list = append(list, t)
		}
		return engine.List(list...)
	case int64:
		return engine.Integer(x)
	case float64:
		return engine.Float(x)
	case json.Number:
		// TODO: less dumb
		s := string(x)
		if strings.ContainsRune(s, '.') {
			n, err := strconv.ParseFloat(s, 64)
			if err != nil {
				panic(err)
			}
			return engine.Float(n)
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			panic(err)
		}
		return engine.Integer(n)
	case string:
		return engine.Atom(x)
	case bool:
		if x {
			return engine.Atom("@").Apply(engine.Atom("true"))
		}
		return engine.Atom("@").Apply(engine.Atom("false"))
	case nil:
		// I guess JSON null is Prolog []?
		return engine.Atom("[]")
	}

	panic(fmt.Errorf("unhandled iface: %T", v))
}

func jsonOf(t engine.Term, env *engine.Env) (string, error) {
	if t == nil {
		return "", nil
	}
	return chars.Value[string](t, env)
}
