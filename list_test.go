package predicates

import (
	"testing"

	"github.com/guregu/predicates/internal"
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
