package predicates

import (
	"testing"

	"github.com/guregu/predicates/internal"
	"github.com/ichiban/prolog/engine"
)

func TestIsList(t *testing.T) {
	p := internal.NewTestProlog()
	p.Register1("is_list", IsList)

	t.Run("term is empty list", p.Expect(internal.TestOK,
		`is_list([]), OK = true.`))

	t.Run("term is list", p.Expect(internal.TestOK,
		`is_list([a, b, c]), OK = true.`))

	t.Run("term is unground list", p.Expect(internal.TestOK,
		`is_list([_]), OK = true.`))

	// this is true on SWI and false on Tau, so let's side with Tau for now
	t.Run("term is [_|_]", p.Expect(internal.TestFail,
		`is_list([_|_]), OK = true.`))

	t.Run("term is atom", p.Expect(internal.TestFail,
		`is_list(foo), OK = true.`))

	t.Run("term is number", p.Expect(internal.TestFail,
		`is_list(555), OK = true.`))

	t.Run("term is variable", p.Expect(internal.TestFail,
		`is_list(X), OK = true.`))
}

func TestAtomicListConcat(t *testing.T) {
	p := internal.NewTestProlog()
	p.Register3("atomic_list_concat", AtomicListConcat)

	t.Run("list is ground", func(t *testing.T) {
		t.Run("atom is variable", p.Expect([]map[string]engine.Term{
			{"X": engine.Atom("a-b")},
		}, `atomic_list_concat([a, b], '-', X).`))

		t.Run("atom is variable and seperator is empty", p.Expect([]map[string]engine.Term{
			{"X": engine.Atom("ab")},
		}, `atomic_list_concat([a, b], '', X).`))

		t.Run("atom is ground", p.Expect(internal.TestOK,
			`atomic_list_concat([a, b], '/', 'a/b'), OK = true.`))

		t.Run("list is empty and atom is variable", p.Expect([]map[string]engine.Term{
			{"X": engine.Atom("")},
		}, `atomic_list_concat([], '-', X).`))

		t.Run("list is empty and atom is ground", p.Expect(internal.TestOK,
			`atomic_list_concat([], '/', ''), OK = true.`))
	})

	t.Run("atom is ground", func(t *testing.T) {
		t.Run("list is variable", p.Expect([]map[string]engine.Term{
			{"X": engine.List(engine.Atom("a"), engine.Atom("b"), engine.Atom("c"))},
		}, `atomic_list_concat(X, '-', 'a-b-c').`))

		t.Run("list is variable and seperator is empty", p.Expect([]map[string]engine.Term{
			{"X": engine.List(engine.Atom("a"), engine.Atom("b"), engine.Atom("c"))},
		}, `atomic_list_concat(X, '', 'abc').`))

		t.Run("atom is empty", p.Expect([]map[string]engine.Term{
			{"X": engine.Atom("")},
		}, `atomic_list_concat([], '-', X).`))

		// seems like Tau and SWI both bind [''] instead of [] to the list here
		t.Run("atom is empty and list is var", p.Expect([]map[string]engine.Term{
			{"X": engine.List(engine.Atom(""))},
		}, `atomic_list_concat(X, '/', '').`))
	})
}
