package internal

import "github.com/ichiban/prolog/engine"

/*
MIT License

Copyright (c) 2021 Yutaka Ichibangase

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

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
