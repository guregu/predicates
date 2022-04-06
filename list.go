package predicates

import (
	"log"

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
			log.Println("ITER ERR", iter.Err())
			return engine.Bool(false)
		}
		return k(env)
	default:
		return engine.Bool(false)
	}
}
