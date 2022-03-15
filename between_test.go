package predicates

import (
	"testing"

	"github.com/ichiban/prolog/engine"

	"github.com/guregu/predicates/internal"
)

var (
	truth = []map[string]engine.Term{{}}
	okay  = []map[string]engine.Term{{"OK": engine.Atom("true")}}
	fail  = []map[string]engine.Term(nil)
)

func TestBetween(t *testing.T) {
	p := internal.NewTestProlog()
	p.Interpreter.Register3("between", Between)

	t.Run("variable number",
		p.Expect([]map[string]engine.Term{
			{"X": engine.Integer(1)},
			{"X": engine.Integer(2)},
			{"X": engine.Integer(3)},
		}, "between(1,3,X)."))

	t.Run("ground number",
		p.Expect(okay, "between(0,2,1), OK = true."))

	t.Run("wrong order",
		p.Expect(fail, "between(3,1,2)."))
}
