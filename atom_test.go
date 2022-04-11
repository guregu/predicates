package predicates

import (
	"testing"

	"github.com/guregu/predicates/internal"
	"github.com/ichiban/prolog/engine"
)

func TestUpcaseAtom(t *testing.T) {
	p := internal.NewTestProlog()
	p.Register2("upcase_atom", UpcaseAtom)

	t.Run("lower is variable", func(t *testing.T) {
		t.Run("atom is lowercase", p.Expect([]map[string]engine.Term{
			{"X": engine.Atom("ABC")},
		}, `upcase_atom('abc', X).`))

		t.Run("atom is uppercase", p.Expect([]map[string]engine.Term{
			{"X": engine.Atom("ABC")},
		}, `upcase_atom('ABC', X).`))
	})

	t.Run("lower is ground", func(t *testing.T) {
		t.Run("atom is lowercase", p.Expect(internal.TestOK,
			`upcase_atom('abc', 'ABC'), OK = true.`))

		t.Run("atom is uppercase", p.Expect(internal.TestOK,
			`upcase_atom('ABC', 'ABC'), OK = true.`))
	})
}

func TestDowncaseAtom(t *testing.T) {
	p := internal.NewTestProlog()
	p.Register2("downcase_atom", DowncaseAtom)

	t.Run("lower is variable", func(t *testing.T) {
		t.Run("atom is lowercase", p.Expect([]map[string]engine.Term{
			{"X": engine.Atom("abc")},
		}, `downcase_atom('abc', X).`))

		t.Run("atom is uppercase", p.Expect([]map[string]engine.Term{
			{"X": engine.Atom("abc")},
		}, `downcase_atom('ABC', X).`))
	})

	t.Run("lower is ground", func(t *testing.T) {
		t.Run("atom is lowercase", p.Expect(internal.TestOK,
			`downcase_atom('abc', 'abc'), OK = true.`))

		t.Run("atom is uppercase", p.Expect(internal.TestOK,
			`downcase_atom('ABC', 'abc'), OK = true.`))
	})
}
