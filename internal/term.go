package internal

import "github.com/ichiban/prolog/engine"

func IsMap(t engine.Term, env *engine.Env) bool {
	c, ok := t.(engine.Compound)
	if !ok {
		return false
	}
	if c.Functor() == "-" && c.Arity() == 2 {
		return true
	}

	iter := engine.ListIterator{List: c, Env: env}
	for iter.Next() {
		elem := iter.Current()
		cmp, ok := env.Resolve(elem).(engine.Compound)
		if !ok {
			return false
		}
		if cmp.Functor() != "-" || cmp.Arity() != 2 {
			return false
		}
	}
	if err := iter.Err(); err != nil {
		return false
	}
	return true
}
