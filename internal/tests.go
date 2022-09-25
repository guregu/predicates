package internal

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ichiban/prolog"
	"github.com/ichiban/prolog/engine"
)

func NewTestProlog() *TestProlog {
	p := &TestProlog{
		Interpreter: prolog.New(os.Stdin, os.Stdout),
	}
	if err := p.Exec(`:- set_prolog_flag(unknown, error).`); err != nil {
		panic(err)
	}
	return p
}

type TestProlog struct {
	*prolog.Interpreter
}

func (p *TestProlog) MustExec(t *testing.T, prog string, args ...interface{}) {
	t.Helper()

	err := p.Exec(prog, args...)
	if err != nil {
		t.Fatal("mustExec:", err)
	}
}

func (p *TestProlog) Expect(want []map[string]engine.Term, query string, args ...interface{}) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		t.Logf("query: %s", query)
		sol, err := p.Query(query)
		if err != nil {
			panic(err)
		}
		defer sol.Close()
		n := 0
		var got []map[string]engine.Term
		for sol.Next() {
			n++
			if got == nil {
				got = make([]map[string]engine.Term, 0, len(want))
			}
			vars := map[string]engine.Term{}
			if err := sol.Scan(vars); err != nil {
				t.Log("scan", err)
			}

			for k := range vars {
				if k[0] == '_' {
					// ignore _vars
					delete(vars, k)
				}
			}

			got = append(got, vars)
			t.Log("solution:", vars)
		}
		if err := sol.Err(); err != nil {
			t.Fatal(err)
		}
		if n != len(want) {
			t.Log("want", len(want), "solutions but got", n)
		}

		if diff := cmp.Diff(want, got, cmpTerm()); diff != "" {
			t.Error("output mismatch (-want +got):\n", diff)
		}
	}
}

func cmpTerm() cmp.Option {
	return cmp.FilterValues(func(x, y interface{}) bool {
		_, ok1 := x.(engine.Term)
		_, ok2 := y.(engine.Term)
		return ok1 && ok2
	}, cmp.Comparer(func(x, y interface{}) bool {
		t1 := x.(engine.Term)
		t2 := y.(engine.Term)

		var b1, b2 bytes.Buffer
		engine.WriteTerm(&b1, t1, &engine.WriteOptions{Quoted: true}, nil)
		engine.WriteTerm(&b2, t2, &engine.WriteOptions{Quoted: true}, nil)

		return b1.String() == b2.String()
	}))
}

// predefined test results
var (
	TestTruth = []map[string]engine.Term{{}}
	TestOK    = []map[string]engine.Term{{"OK": engine.Atom("true")}}
	TestFail  = []map[string]engine.Term(nil)
)
