package internal

import (
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
			got = append(got, vars)
			t.Log("solution:", vars)
		}
		if n != len(want) {
			t.Log("want", len(want), "solutions but got", n)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Error("output mismatch (-want +got):\n", diff)
		}
	}
}
