package taujson

import (
	"testing"

	"github.com/ichiban/prolog/engine"

	"github.com/guregu/predicates/internal"
)

func TestJSON(t *testing.T) {
	p := internal.NewTestProlog()
	Register(p.Interpreter)

	t.Run("Prolog → JSON", func(t *testing.T) {
		t.Run("value is list", p.Expect([]map[string]engine.Term{
			{"JSON": engine.Atom(`["a",1]`)},
		}, `json_prolog(_JS, [a, 1]), json_atom(_JS, JSON).`))

		t.Run("value is empty list", p.Expect([]map[string]engine.Term{
			{"JSON": engine.Atom(`[]`)},
		}, `json_prolog(_JS, []), json_atom(_JS, JSON).`))

		t.Run("value is map", p.Expect([]map[string]engine.Term{
			{"JSON": engine.Atom(`{"a":1}`), "OK": engine.Atom("true")},
		}, `json_prolog(_JS, [a-1]), json_atom(_JS, JSON), json_prolog(_JS, [a-1]), OK = true.`))

		t.Run("value is map of maps", p.Expect([]map[string]engine.Term{
			{"JSON": engine.Atom(`{"a":{"b":{"c":{"d":{"e":555}}}}}`)},
		}, `json_prolog(_JS, [a-[b-[c-[d-[e-555]]]]]), json_atom(_JS, JSON).`))

		t.Run("value is non-map non-list compound", p.Expect([]map[string]engine.Term{
			{"JSON": engine.Atom(`{"a":"b(c)"}`)},
		}, `json_prolog(_JS, [a-b(c)]), json_atom(_JS, JSON).`))
	})

	t.Run("JSON → Prolog", func(t *testing.T) {
		t.Run("value is list", p.Expect([]map[string]engine.Term{
			{"X": engine.List(engine.Atom("a"), engine.Integer(1))},
		}, `json_atom(_JS, '["a",1]'), json_prolog(_JS, X).`))

		t.Run("value is empty list", p.Expect([]map[string]engine.Term{
			{"X": engine.Atom("[]")},
		}, `json_atom(_JS, '[]'), json_prolog(_JS, X).`))

		t.Run("value is map", p.Expect([]map[string]engine.Term{
			{"X": engine.List(engine.Atom("-").Apply(engine.Atom("a"), engine.Integer(1)))},
		}, `json_atom(_JS, '{"a":1}'), json_prolog(_JS, X).`))

		t.Run("value is map of maps", p.Expect([]map[string]engine.Term{
			{"X": engine.List(engine.Atom("-").Apply(engine.Atom("a"),
				engine.List(engine.Atom("-").Apply(engine.Atom("b"),
					engine.List(engine.Atom("-").Apply(engine.Atom("c"),
						engine.List(engine.Atom("-").Apply(engine.Atom("d"),
							engine.List(engine.Atom("-").Apply(engine.Atom("e"), engine.Integer(555)))))))))))},
		}, `json_atom(_JS, '{"a":{"b":{"c":{"d":{"e":555}}}}}'), json_prolog(_JS, X).`))

		t.Run("value is null", p.Expect([]map[string]engine.Term{
			{"X": engine.List(engine.Atom("-").Apply(engine.Atom("a"), engine.Atom("[]")))},
		}, `json_atom(_JS, '{"a":null}'), json_prolog(_JS, X).`))
	})
}
