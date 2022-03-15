package predicates

import (
	"testing"
	"testing/fstest"

	"github.com/guregu/predicates/internal"
	"github.com/ichiban/prolog/engine"
)

func TestFS(t *testing.T) {
	p := internal.NewTestProlog()
	fsys := fstest.MapFS{
		"test.pl": &fstest.MapFile{
			Data: []byte("hello(world)."),
		},
	}
	ff := NewFS(fsys, p.Interpreter)
	p.Interpreter.Register1("consult", ff.Consult)

	t.Run("consult/1", func(t *testing.T) {
		p.MustExec(t, ":- consult(test).")
		p.Expect([]map[string]engine.Term{
			{"X": engine.Atom("world")},
		}, "hello(X).")
	})
}
