package internal

import "github.com/ichiban/prolog/engine"

// helper functions for creating runtime errors.
// Mostly copied from ichiban/prolog/engine internals (for easy future merging?).

func TypeErrorInteger(culprit engine.Term) *engine.Exception {
	return engine.TypeError("integer", culprit)
}

func TypeErrorCompound(culprit engine.Term) *engine.Exception {
	return engine.TypeError("compound", culprit)
}

func TypeErrorPair(culprit engine.Term) *engine.Exception {
	return engine.TypeError("pair", culprit)
}

func TypeErrorAtom(culprit engine.Term) *engine.Exception {
	return engine.TypeError("pair", culprit)
}

func EvaluationError(error, info engine.Term) *engine.Exception {
	return &engine.Exception{
		Term: &engine.Compound{
			Functor: "error",
			Args: []engine.Term{
				&engine.Compound{
					Functor: "evaluation_error",
					Args:    []engine.Term{error},
				},
				info,
			},
		},
	}
}

// http://www.gprolog.org/manual/html_node/gprolog020.html#sec44
func EvaluationErrorIntOverflow() *engine.Exception {
	return EvaluationError(engine.Atom("int_overflow"), engine.Atom("integer overflow."))
}
